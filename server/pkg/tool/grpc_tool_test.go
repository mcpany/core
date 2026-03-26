// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"context"
	"encoding/json"
	"errors"
	"net"
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
	"google.golang.org/protobuf/reflect/protoreflect"
)

func findMethodDescriptor(t *testing.T, serviceName, methodName string) protoreflect.MethodDescriptor {
	t.Helper()

	service := weatherpb.File_proto_examples_weather_v1_weather_proto.Services().ByName(protoreflect.Name(serviceName))
	require.NotNil(t, service, "service %s not found", serviceName)
	methodDesc := service.Methods().ByName(protoreflect.Name(methodName))

	require.NotNil(t, methodDesc, "method %s not found in service %s", methodName, serviceName)
	return methodDesc
}

// TestNewGRPCTool ...
// Summary: TestNewGRPCTool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// GetWeather ...
// Summary: GetWeather
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// Get ...
// Summary: Get
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	if m.getFunc != nil {
		return m.getFunc(ctx)
	}
	return nil, nil
}

// Put ...
// Summary: Put
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	if m.putFunc != nil {
		m.putFunc(c)
	}
}

// TestGRPCTool_Execute ...
// Summary: TestGRPCTool_Execute
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
