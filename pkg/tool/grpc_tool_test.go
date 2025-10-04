/*
 * Copyright 2025 Author(s) of MCP-XY
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tool_test

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"os"
	"testing"

	"github.com/mcpxy/core/pkg/client"
	"github.com/mcpxy/core/pkg/pool"
	"github.com/mcpxy/core/pkg/tool"
	calculatorpb "github.com/mcpxy/core/proto/examples/calculator/v1"
	v1 "github.com/mcpxy/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func findMethodDescriptor(t *testing.T, serviceName, methodName string) protoreflect.MethodDescriptor {
	t.Helper()
	b, err := os.ReadFile("../../build/all.protoset")
	require.NoError(t, err, "Failed to read protoset file. Ensure 'make gen-local' has been run.")

	fds := &descriptorpb.FileDescriptorSet{}
	err = proto.Unmarshal(b, fds)
	require.NoError(t, err, "Failed to unmarshal protoset file")

	files, err := protodesc.NewFiles(fds)
	require.NoError(t, err)

	var methodDesc protoreflect.MethodDescriptor
	files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		services := fd.Services()
		for i := 0; i < services.Len(); i++ {
			service := services.Get(i)
			if string(service.Name()) == serviceName {
				method := service.Methods().ByName(protoreflect.Name(methodName))
				if method != nil {
					methodDesc = method
					return false // stop iterating
				}
			}
		}
		return true
	})

	require.NotNil(t, methodDesc, "method %s not found in service %s", methodName, serviceName)
	return methodDesc
}

func TestNewGRPCTool(t *testing.T) {
	pm := pool.NewManager()
	serviceKey := "test-service"
	toolProto := &v1.Tool{}
	methodDesc := findMethodDescriptor(t, "CalculatorService", "Add")

	grpcTool := tool.NewGRPCTool(toolProto, pm, serviceKey, methodDesc)
	require.NotNil(t, grpcTool)
	assert.Equal(t, toolProto, grpcTool.Tool())
}

// mockCalculatorServer is a mock implementation of the CalculatorServiceServer for testing.
type mockCalculatorServer struct {
	calculatorpb.UnimplementedCalculatorServiceServer
	addFunc func(ctx context.Context, req *calculatorpb.AddRequest) (*calculatorpb.AddResponse, error)
}

func (s *mockCalculatorServer) Add(ctx context.Context, req *calculatorpb.AddRequest) (*calculatorpb.AddResponse, error) {
	if s.addFunc != nil {
		return s.addFunc(ctx, req)
	}
	res := &calculatorpb.AddResponse{}
	res.SetResult(req.GetA() + req.GetB())
	return res, nil
}

// setupGrpcTest sets up a mock gRPC server and returns a client connection to it.
func setupGrpcTest(t *testing.T, srv calculatorpb.CalculatorServiceServer) *grpc.ClientConn {
	t.Helper()
	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	calculatorpb.RegisterCalculatorServiceServer(s, srv)
	go func() {
		if err := s.Serve(lis); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
	conn, err := grpc.NewClient("passthrough:///bufnet", grpc.WithContextDialer(dialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	t.Cleanup(func() {
		s.Stop()
		conn.Close()
	})

	return conn
}

// mockGrpcPool implements the pool.Pool interface for testing.
type mockGrpcPool struct {
	pool.Pool[*client.GrpcClientWrapper]
	getFunc func(ctx context.Context) (*client.GrpcClientWrapper, error)
	putFunc func(c *client.GrpcClientWrapper)
}

func (m *mockGrpcPool) Get(ctx context.Context) (*client.GrpcClientWrapper, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx)
	}
	return nil, nil
}

func (m *mockGrpcPool) Put(c *client.GrpcClientWrapper) {
	if m.putFunc != nil {
		m.putFunc(c)
	}
}

func TestGRPCTool_Execute(t *testing.T) {
	methodDesc := findMethodDescriptor(t, "CalculatorService", "Add")
	toolProto := &v1.Tool{}
	toolProto.SetUnderlyingMethodFqn(string(methodDesc.FullName()))

	t.Run("successful execution", func(t *testing.T) {
		server := &mockCalculatorServer{
			addFunc: func(ctx context.Context, req *calculatorpb.AddRequest) (*calculatorpb.AddResponse, error) {
				assert.Equal(t, int32(10), req.GetA())
				assert.Equal(t, int32(20), req.GetB())
				res := &calculatorpb.AddResponse{}
				res.SetResult(30)
				return res, nil
			},
		}
		conn := setupGrpcTest(t, server)
		wrapper := &client.GrpcClientWrapper{ClientConn: conn}

		pm := pool.NewManager()
		mockPool := &mockGrpcPool{
			getFunc: func(ctx context.Context) (*client.GrpcClientWrapper, error) {
				return wrapper, nil
			},
		}
		pm.Register("grpc-test", mockPool)

		grpcTool := tool.NewGRPCTool(toolProto, pm, "grpc-test", methodDesc)
		inputs := json.RawMessage(`{"a": 10, "b": 20}`)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		result, err := grpcTool.Execute(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, map[string]any{"result": float64(30)}, result) // JSON unmarshals numbers to float64
	})

	t.Run("pool get error", func(t *testing.T) {
		pm := pool.NewManager()
		mockPool := &mockGrpcPool{
			getFunc: func(ctx context.Context) (*client.GrpcClientWrapper, error) {
				return nil, errors.New("pool error")
			},
		}
		pm.Register("grpc-test", mockPool)

		grpcTool := tool.NewGRPCTool(toolProto, pm, "grpc-test", methodDesc)
		req := &tool.ExecutionRequest{ToolInputs: json.RawMessage(`{}`)}
		_, err := grpcTool.Execute(context.Background(), req)
		assert.Error(t, err)
	})

	t.Run("invoke error", func(t *testing.T) {
		server := &mockCalculatorServer{
			addFunc: func(ctx context.Context, req *calculatorpb.AddRequest) (*calculatorpb.AddResponse, error) {
				return nil, errors.New("invoke error")
			},
		}
		conn := setupGrpcTest(t, server)
		wrapper := &client.GrpcClientWrapper{ClientConn: conn}

		pm := pool.NewManager()
		mockPool := &mockGrpcPool{
			getFunc: func(ctx context.Context) (*client.GrpcClientWrapper, error) {
				return wrapper, nil
			},
		}
		pm.Register("grpc-test", mockPool)

		grpcTool := tool.NewGRPCTool(toolProto, pm, "grpc-test", methodDesc)
		inputs := json.RawMessage(`{"a": 10, "b": 20}`)
		req := &tool.ExecutionRequest{ToolInputs: inputs}
		_, err := grpcTool.Execute(context.Background(), req)
		assert.Error(t, err)
	})

	t.Run("invalid input json", func(t *testing.T) {
		pm := pool.NewManager()
		grpcTool := tool.NewGRPCTool(toolProto, pm, "grpc-test", methodDesc)
		inputs := json.RawMessage(`{invalid}`)
		req := &tool.ExecutionRequest{ToolInputs: inputs}
		_, err := grpcTool.Execute(context.Background(), req)
		assert.Error(t, err)
	})
}
