// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	weatherpb "github.com/mcpany/core/proto/examples/weather/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/tool"
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

	// Try using runtime.Caller to find the test file location
	_, filename, _, ok := runtime.Caller(0)
	var path string
	if ok {
		baseDir := filepath.Dir(filename)
		// Assume repo root is 3 levels up from server/pkg/tool
		path = filepath.Join(baseDir, "../../../build/all.protoset")
	} else {
		path = "../../../build/all.protoset"
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Fallback to absolute path in container
		path = "/app/build/all.protoset"
		if _, err := os.Stat(path); os.IsNotExist(err) {
			// Debugging info for CI failure
			cwd, _ := os.Getwd()
			t.Logf("Current working directory: %s", cwd)
			t.Logf("Failed to find protoset at computed path: %s", path)

			// Check what IS available
			if entries, err := os.ReadDir("../../.."); err == nil {
				var names []string
				for _, e := range entries {
					names = append(names, e.Name())
				}
				t.Logf("Entries in ../../..: %v", names)
			} else {
				t.Logf("Failed to read ../../..: %v", err)
			}
			if entries, err := os.ReadDir("../../../build"); err == nil {
				var names []string
				for _, e := range entries {
					names = append(names, e.Name())
				}
				t.Logf("Entries in ../../../build: %v", names)
			}
		}
	}
	b, err := os.ReadFile(path)
	require.NoError(t, err, "Failed to read protoset file at %s. Ensure 'make gen' has been run.", path)

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
	t.Parallel()
	pm := pool.NewManager()
	serviceID := "test-service"
	toolProto := &v1.Tool{}
	methodDesc := findMethodDescriptor(t, "WeatherService", "GetWeather")

	grpcTool := tool.NewGRPCTool(toolProto, pm, serviceID, methodDesc, nil, nil)
	require.NotNil(t, grpcTool)
	assert.Equal(t, toolProto, grpcTool.Tool())
}

// mockWeatherServer is a mock implementation of the WeatherServiceServer for testing.
type mockWeatherServer struct {
	weatherpb.UnimplementedWeatherServiceServer
	getWeatherFunc func(ctx context.Context, req *weatherpb.GetWeatherRequest) (*weatherpb.GetWeatherResponse, error)
}

func (s *mockWeatherServer) GetWeather(ctx context.Context, req *weatherpb.GetWeatherRequest) (*weatherpb.GetWeatherResponse, error) {
	if s.getWeatherFunc != nil {
		return s.getWeatherFunc(ctx, req)
	}
	return weatherpb.GetWeatherResponse_builder{Weather: "sunny"}.Build(), nil
}

// setupGrpcTest sets up a mock gRPC server and returns a client connection to it.
func setupGrpcTest(t *testing.T, srv weatherpb.WeatherServiceServer) *grpc.ClientConn {
	t.Helper()
	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	weatherpb.RegisterWeatherServiceServer(s, srv)
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
		_ = conn.Close()
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
	t.Parallel()
	methodDesc := findMethodDescriptor(t, "WeatherService", "GetWeather")
	toolProto := &v1.Tool{}
	toolProto.SetUnderlyingMethodFqn(string(methodDesc.FullName()))

	t.Run("successful execution", func(t *testing.T) {
		server := &mockWeatherServer{
			getWeatherFunc: func(_ context.Context, req *weatherpb.GetWeatherRequest) (*weatherpb.GetWeatherResponse, error) {
				assert.Equal(t, "London", req.GetLocation())
				return weatherpb.GetWeatherResponse_builder{Weather: "sunny"}.Build(), nil
			},
		}
		conn := setupGrpcTest(t, server)
		wrapper := client.NewGrpcClientWrapper(conn, nil, nil)

		pm := pool.NewManager()
		mockPool := &mockGrpcPool{
			getFunc: func(_ context.Context) (*client.GrpcClientWrapper, error) {
				return wrapper, nil
			},
		}
		pm.Register("grpc-test", mockPool)

		grpcTool := tool.NewGRPCTool(toolProto, pm, "grpc-test", methodDesc, nil, nil)
		inputs := json.RawMessage(`{"location": "London"}`)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		result, err := grpcTool.Execute(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, map[string]any{"weather": "sunny"}, result)
	})

	t.Run("pool get error", func(t *testing.T) {
		pm := pool.NewManager()
		mockPool := &mockGrpcPool{
			getFunc: func(_ context.Context) (*client.GrpcClientWrapper, error) {
				return nil, errors.New("pool error")
			},
		}
		pm.Register("grpc-test", mockPool)

		grpcTool := tool.NewGRPCTool(toolProto, pm, "grpc-test", methodDesc, nil, nil)
		req := &tool.ExecutionRequest{ToolInputs: json.RawMessage(`{}`)}
		_, err := grpcTool.Execute(context.Background(), req)
		assert.Error(t, err)
	})

	t.Run("invoke error", func(t *testing.T) {
		server := &mockWeatherServer{
			getWeatherFunc: func(_ context.Context, _ *weatherpb.GetWeatherRequest) (*weatherpb.GetWeatherResponse, error) {
				return nil, errors.New("invoke error")
			},
		}
		conn := setupGrpcTest(t, server)
		wrapper := client.NewGrpcClientWrapper(conn, nil, nil)

		pm := pool.NewManager()
		mockPool := &mockGrpcPool{
			getFunc: func(_ context.Context) (*client.GrpcClientWrapper, error) {
				return wrapper, nil
			},
		}
		pm.Register("grpc-test", mockPool)

		grpcTool := tool.NewGRPCTool(toolProto, pm, "grpc-test", methodDesc, nil, nil)
		inputs := json.RawMessage(`{"location": "London"}`)
		req := &tool.ExecutionRequest{ToolInputs: inputs}
		_, err := grpcTool.Execute(context.Background(), req)
		assert.Error(t, err)
	})

	t.Run("invalid input json", func(t *testing.T) {
		pm := pool.NewManager()
		grpcTool := tool.NewGRPCTool(toolProto, pm, "grpc-test", methodDesc, nil, nil)
		inputs := json.RawMessage(`{invalid}`)
		req := &tool.ExecutionRequest{ToolInputs: inputs}
		_, err := grpcTool.Execute(context.Background(), req)
		assert.Error(t, err)
	})
}
