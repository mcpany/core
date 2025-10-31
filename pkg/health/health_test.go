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

package health

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	configv1 "github.com/mcpxy/core/proto/config/v1"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/protobuf/types/known/durationpb"
)

// mockHealthServer is a mock implementation of the gRPC health check server.
type mockHealthServer struct {
	grpc_health_v1.UnimplementedHealthServer
	status grpc_health_v1.HealthCheckResponse_ServingStatus
}

func (s *mockHealthServer) Check(ctx context.Context, in *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{Status: s.status}, nil
}

func (s *mockHealthServer) Watch(in *grpc_health_v1.HealthCheckRequest, srv grpc_health_v1.Health_WatchServer) error {
	return srv.Send(&grpc_health_v1.HealthCheckResponse{Status: s.status})
}

// newMockGRPCHealthServer starts a gRPC server with the mock health service.
func newMockGRPCHealthServer(status grpc_health_v1.HealthCheckResponse_ServingStatus) (*grpc.Server, net.Listener) {
	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	grpc_health_v1.RegisterHealthServer(s, &mockHealthServer{status: status})
	go s.Serve(lis)
	return s, lis
}

func TestCheck(t *testing.T) {
	ctx := context.Background()

	t.Run("NilConfig", func(t *testing.T) {
		assert.False(t, Check(ctx, nil), "Check with nil config should return false")
	})
}

func TestCheckHTTPHealth(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		serverURL := server.URL
		serverAddr := server.Listener.Addr().String()

		healthCheck := configv1.HttpHealthCheck_builder{
			Url:          &serverURL,
			ExpectedCode: lo.ToPtr(int32(http.StatusOK)),
		}.Build()
		httpService := configv1.HttpUpstreamService_builder{
			Address:     &serverAddr,
			HealthCheck: healthCheck,
		}.Build()
		upstreamConfig := configv1.UpstreamServiceConfig_builder{
			HttpService: httpService,
		}.Build()

		assert.True(t, Check(ctx, upstreamConfig), "HTTP health check should be successful")
	})

	t.Run("UnexpectedStatusCode", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		serverURL := server.URL
		serverAddr := server.Listener.Addr().String()

		healthCheck := configv1.HttpHealthCheck_builder{
			Url:          &serverURL,
			ExpectedCode: lo.ToPtr(int32(http.StatusOK)),
		}.Build()
		httpService := configv1.HttpUpstreamService_builder{
			Address:     &serverAddr,
			HealthCheck: healthCheck,
		}.Build()

		assert.False(t, checkHTTPHealth(ctx, httpService), "HTTP health check should fail with unexpected status code")
	})

	t.Run("ServerUnreachable", func(t *testing.T) {
		unreachableURL := "http://localhost:12345"
		unreachableAddr := "localhost:12345"

		healthCheck := configv1.HttpHealthCheck_builder{
			Url:          &unreachableURL,
			ExpectedCode: lo.ToPtr(int32(http.StatusOK)),
			Timeout:      durationpb.New(10 * time.Millisecond),
		}.Build()
		httpService := configv1.HttpUpstreamService_builder{
			Address:     &unreachableAddr,
			HealthCheck: healthCheck,
		}.Build()

		assert.False(t, checkHTTPHealth(ctx, httpService), "HTTP health check should fail for unreachable server")
	})
}

func TestCheckGRPCHealth(t *testing.T) {
	ctx := context.Background()

	t.Run("Serving", func(t *testing.T) {
		server, lis := newMockGRPCHealthServer(grpc_health_v1.HealthCheckResponse_SERVING)
		defer server.Stop()

		serviceName := "test-service"
		healthCheck := configv1.GrpcHealthCheck_builder{
			Service: &serviceName,
		}.Build()
		grpcService := configv1.GrpcUpstreamService_builder{
			Address:     lo.ToPtr(lis.Addr().String()),
			HealthCheck: healthCheck,
		}.Build()
		upstreamConfig := configv1.UpstreamServiceConfig_builder{
			GrpcService: grpcService,
		}.Build()

		assert.True(t, Check(ctx, upstreamConfig), "gRPC health check should be successful for SERVING status")
	})

	t.Run("NotServing", func(t *testing.T) {
		server, lis := newMockGRPCHealthServer(grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		defer server.Stop()

		serviceName := "test-service"
		healthCheck := configv1.GrpcHealthCheck_builder{
			Service: &serviceName,
		}.Build()
		grpcService := configv1.GrpcUpstreamService_builder{
			Address:     lo.ToPtr(lis.Addr().String()),
			HealthCheck: healthCheck,
		}.Build()

		assert.False(t, checkGRPCHealth(ctx, grpcService), "gRPC health check should fail for NOT_SERVING status")
	})

	t.Run("ServerUnreachable", func(t *testing.T) {
		serviceName := "test-service"
		healthCheck := configv1.GrpcHealthCheck_builder{
			Service: &serviceName,
		}.Build()
		grpcService := configv1.GrpcUpstreamService_builder{
			Address:     lo.ToPtr("localhost:12345"),
			HealthCheck: healthCheck,
		}.Build()

		assert.False(t, checkGRPCHealth(ctx, grpcService), "gRPC health check should fail for unreachable server")
	})
}

func TestCheckConnection(t *testing.T) {
	ctx := context.Background()

	t.Run("ConnectionSuccess", func(t *testing.T) {
		lis, err := net.Listen("tcp", "localhost:0")
		assert.NoError(t, err)
		defer lis.Close()
		assert.True(t, checkConnection(ctx, lis.Addr().String()), "checkConnection should succeed for a listening port")
	})

	t.Run("ConnectionFailure", func(t *testing.T) {
		assert.False(t, checkConnection(ctx, "localhost:12345"), "checkConnection should fail for a non-listening port")
	})
}

func TestCheckVariousServices(t *testing.T) {
	ctx := context.Background()

	// Setup a simple listening server for connection checks
	lis, err := net.Listen("tcp", "localhost:0")
	assert.NoError(t, err)
	defer lis.Close()
	addr := lis.Addr().String()

	testCases := []struct {
		name   string
		config *configv1.UpstreamServiceConfig
		want   bool
	}{
		{
			name: "OpenAPI Service",
			config: configv1.UpstreamServiceConfig_builder{
				OpenapiService: configv1.OpenapiUpstreamService_builder{Address: &addr}.Build(),
			}.Build(),
			want: true,
		},
		{
			name: "WebSocket Service",
			config: configv1.UpstreamServiceConfig_builder{
				WebsocketService: configv1.WebsocketUpstreamService_builder{Address: &addr}.Build(),
			}.Build(),
			want: true,
		},
		{
			name: "WebRTC Service",
			config: configv1.UpstreamServiceConfig_builder{
				WebrtcService: configv1.WebrtcUpstreamService_builder{Address: &addr}.Build(),
			}.Build(),
			want: true,
		},
		{
			name: "MCP Service HTTP",
			config: configv1.UpstreamServiceConfig_builder{
				McpService: configv1.McpUpstreamService_builder{
					HttpConnection: configv1.McpStreamableHttpConnection_builder{HttpAddress: &addr}.Build(),
				}.Build(),
			}.Build(),
			want: true,
		},
		{
			name: "MCP Service Stdio",
			config: configv1.UpstreamServiceConfig_builder{
				McpService: configv1.McpUpstreamService_builder{
					StdioConnection: configv1.McpStdioConnection_builder{Command: lo.ToPtr("echo")}.Build(),
				}.Build(),
			}.Build(),
			want: true,
		},
		{
			name: "Command Line Service",
			config: configv1.UpstreamServiceConfig_builder{
				CommandLineService: (&configv1.CommandLineUpstreamService_builder{}).Build(),
			}.Build(),
			want: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, Check(ctx, tc.config))
		})
	}
}
