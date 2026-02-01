// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"os"
	"strings"
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

// Duplicate helpers from grpc_tool_test.go since they are not exported

func findMethodDescriptorSec(t *testing.T, serviceName, methodName string) protoreflect.MethodDescriptor {
	t.Helper()
	path := "../../../build/all.protoset"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		path = "/app/build/all.protoset"
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

// mockWeatherServerSec is a mock implementation of the WeatherServiceServer for testing.
type mockWeatherServerSec struct {
	weatherpb.UnimplementedWeatherServiceServer
	getWeatherFunc func(ctx context.Context, req *weatherpb.GetWeatherRequest) (*weatherpb.GetWeatherResponse, error)
}

func (s *mockWeatherServerSec) GetWeather(ctx context.Context, req *weatherpb.GetWeatherRequest) (*weatherpb.GetWeatherResponse, error) {
	if s.getWeatherFunc != nil {
		return s.getWeatherFunc(ctx, req)
	}
	return weatherpb.GetWeatherResponse_builder{Weather: "sunny"}.Build(), nil
}

// setupGrpcTestSec sets up a mock gRPC server and returns a client connection to it.
func setupGrpcTestSec(t *testing.T, srv weatherpb.WeatherServiceServer) *grpc.ClientConn {
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

// mockGrpcPoolSec implements the pool.Pool interface for testing.
type mockGrpcPoolSec struct {
	pool.Pool[*client.GrpcClientWrapper]
	getFunc func(ctx context.Context) (*client.GrpcClientWrapper, error)
	putFunc func(c *client.GrpcClientWrapper)
}

func (m *mockGrpcPoolSec) Get(ctx context.Context) (*client.GrpcClientWrapper, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx)
	}
	return nil, nil
}

func (m *mockGrpcPoolSec) Put(c *client.GrpcClientWrapper) {
	if m.putFunc != nil {
		m.putFunc(c)
	}
}

func TestGRPCTool_InformationLeakage(t *testing.T) {
	t.Parallel()
	methodDesc := findMethodDescriptorSec(t, "WeatherService", "GetWeather")
	toolProto := &v1.Tool{}
	toolProto.SetUnderlyingMethodFqn(string(methodDesc.FullName()))

	// Simulate a sensitive error message (DSN leak)
	sensitiveError := errors.New("upstream connect failed: postgres://admin:supersecret@10.0.0.1:5432/db")

	server := &mockWeatherServerSec{
		getWeatherFunc: func(_ context.Context, _ *weatherpb.GetWeatherRequest) (*weatherpb.GetWeatherResponse, error) {
			return nil, sensitiveError
		},
	}
	conn := setupGrpcTestSec(t, server)
	wrapper := client.NewGrpcClientWrapper(conn, nil, nil)

	pm := pool.NewManager()
	mockPool := &mockGrpcPoolSec{
		getFunc: func(_ context.Context) (*client.GrpcClientWrapper, error) {
			return wrapper, nil
		},
	}
	pm.Register("grpc-test", mockPool)

	grpcTool := tool.NewGRPCTool(toolProto, pm, "grpc-test", methodDesc, nil, nil)
	inputs := json.RawMessage(`{"location": "London"}`)
	req := &tool.ExecutionRequest{ToolInputs: inputs}

	_, err := grpcTool.Execute(context.Background(), req)

	// Verification: The error should NOT contain the secret
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "supersecret", "Error should NOT contain secret after fix")
	assert.Contains(t, err.Error(), "[REDACTED]", "Error should contain [REDACTED] placeholder")
	// The RedactDSN function might leave "postgres://" but redact the credentials
	// Verify truncation if we test length (not testing length here as the mock error is short)
}

func TestGRPCTool_ErrorTruncation(t *testing.T) {
	t.Parallel()
	methodDesc := findMethodDescriptorSec(t, "WeatherService", "GetWeather")
	toolProto := &v1.Tool{}
	toolProto.SetUnderlyingMethodFqn(string(methodDesc.FullName()))

	// Simulate a very long stack trace
	longErrorMsg := "Stack trace: " + strings.Repeat("at bad.code.path(file.go:123)\n", 100)
	longError := errors.New(longErrorMsg)

	server := &mockWeatherServerSec{
		getWeatherFunc: func(_ context.Context, _ *weatherpb.GetWeatherRequest) (*weatherpb.GetWeatherResponse, error) {
			return nil, longError
		},
	}
	conn := setupGrpcTestSec(t, server)
	wrapper := client.NewGrpcClientWrapper(conn, nil, nil)

	pm := pool.NewManager()
	mockPool := &mockGrpcPoolSec{
		getFunc: func(_ context.Context) (*client.GrpcClientWrapper, error) {
			return wrapper, nil
		},
	}
	pm.Register("grpc-test-trunc", mockPool)

	grpcTool := tool.NewGRPCTool(toolProto, pm, "grpc-test-trunc", methodDesc, nil, nil)
	inputs := json.RawMessage(`{"location": "London"}`)
	req := &tool.ExecutionRequest{ToolInputs: inputs}

	_, err := grpcTool.Execute(context.Background(), req)

	require.Error(t, err)
	// Max length is 500 + length of "failed to invoke grpc method: " (29) = 529?
	// But fmt.Errorf adds the prefix. The errMsg itself is truncated to 500.
	// So the error message ending should contain "(truncated)".
	assert.Contains(t, err.Error(), "(truncated)")
	assert.True(t, len(err.Error()) < 600, "Error message should be truncated (actual len: %d)", len(err.Error()))
}
