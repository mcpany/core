/*
 * Copyright 2025 Author(s) of MCP Any
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

	"github.com/alexliesenfeld/health"
	configv1 "github.com/mcpany/core/proto/config/v1"
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

// mockHealthServerWithFailure is a mock implementation of the gRPC health check server that returns an error.
type mockHealthServerWithFailure struct {
	grpc_health_v1.UnimplementedHealthServer
}

func (s *mockHealthServerWithFailure) Check(ctx context.Context, in *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return nil, assert.AnError
}

// newMockGRPCHealthServerWithFailure starts a gRPC server with the mock health service that returns an error.
func newMockGRPCHealthServerWithFailure() (*grpc.Server, net.Listener) {
	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	grpc_health_v1.RegisterHealthServer(s, &mockHealthServerWithFailure{})
	go s.Serve(lis)
	return s, lis
}

func TestNewChecker(t *testing.T) {
	ctx := context.Background()

	t.Run("NilConfig", func(t *testing.T) {
		assert.Nil(t, NewChecker(nil), "NewChecker with nil config should return nil")
	})

	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		serverURL := server.URL
		serverAddr := server.Listener.Addr().String()

		upstreamConfig := configv1.UpstreamServiceConfig_builder{
			Name: lo.ToPtr("test-service"),
			HttpService: configv1.HttpUpstreamService_builder{
				Address: &serverAddr,
				HealthCheck: configv1.HttpHealthCheck_builder{
					Url:          &serverURL,
					ExpectedCode: lo.ToPtr(int32(http.StatusOK)),
				}.Build(),
			}.Build(),
		}.Build()

		checker := NewChecker(upstreamConfig)
		assert.NotNil(t, checker)
		assert.Equal(t, health.StatusUp, checker.Check(ctx).Status)
	})

	t.Run("UnexpectedStatusCode", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		serverURL := server.URL
		serverAddr := server.Listener.Addr().String()

		upstreamConfig := configv1.UpstreamServiceConfig_builder{
			Name: lo.ToPtr("test-service"),
			HttpService: configv1.HttpUpstreamService_builder{
				Address: &serverAddr,
				HealthCheck: configv1.HttpHealthCheck_builder{
					Url:          &serverURL,
					ExpectedCode: lo.ToPtr(int32(http.StatusOK)),
				}.Build(),
			}.Build(),
		}.Build()

		checker := NewChecker(upstreamConfig)
		assert.NotNil(t, checker)
		assert.Equal(t, health.StatusDown, checker.Check(ctx).Status)
	})

	t.Run("ServerUnreachable", func(t *testing.T) {
		unreachableURL := "http://localhost:12345"
		unreachableAddr := "localhost:12345"

		upstreamConfig := configv1.UpstreamServiceConfig_builder{
			Name: lo.ToPtr("test-service"),
			HttpService: configv1.HttpUpstreamService_builder{
				Address: &unreachableAddr,
				HealthCheck: configv1.HttpHealthCheck_builder{
					Url:          &unreachableURL,
					ExpectedCode: lo.ToPtr(int32(http.StatusOK)),
					Timeout:      durationpb.New(10 * time.Millisecond),
				}.Build(),
			}.Build(),
		}.Build()

		checker := NewChecker(upstreamConfig)
		assert.NotNil(t, checker)
		assert.Equal(t, health.StatusDown, checker.Check(ctx).Status)
	})

	t.Run("GRPCNilHealthCheck", func(t *testing.T) {
		server, lis := newMockGRPCHealthServer(grpc_health_v1.HealthCheckResponse_SERVING)
		defer server.Stop()

		upstreamConfig := configv1.UpstreamServiceConfig_builder{
			Name: lo.ToPtr("test-service"),
			GrpcService: configv1.GrpcUpstreamService_builder{
				Address: lo.ToPtr(lis.Addr().String()),
			}.Build(),
		}.Build()

		checker := NewChecker(upstreamConfig)
		assert.NotNil(t, checker)
		assert.Equal(t, health.StatusUp, checker.Check(ctx).Status)
	})

	t.Run("HealthClientCheckFailure", func(t *testing.T) {
		server, lis := newMockGRPCHealthServerWithFailure()
		defer server.Stop()

		upstreamConfig := configv1.UpstreamServiceConfig_builder{
			Name: lo.ToPtr("test-service"),
			GrpcService: configv1.GrpcUpstreamService_builder{
				Address: lo.ToPtr(lis.Addr().String()),
				HealthCheck: configv1.GrpcHealthCheck_builder{
					Service: lo.ToPtr("test-service"),
				}.Build(),
			}.Build(),
		}.Build()

		checker := NewChecker(upstreamConfig)
		assert.NotNil(t, checker)
		assert.Equal(t, health.StatusDown, checker.Check(ctx).Status)
	})

	t.Run("HTTPNilHealthCheck", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		serverAddr := server.Listener.Addr().String()

		upstreamConfig := configv1.UpstreamServiceConfig_builder{
			Name: lo.ToPtr("test-service"),
			HttpService: configv1.HttpUpstreamService_builder{
				Address: &serverAddr,
			}.Build(),
		}.Build()

		checker := NewChecker(upstreamConfig)
		assert.NotNil(t, checker)
		assert.Equal(t, health.StatusUp, checker.Check(ctx).Status)
	})

	t.Run("InvalidRequestURL", func(t *testing.T) {
		invalidURL := " a"
		serverAddr := "localhost:12345"

		upstreamConfig := configv1.UpstreamServiceConfig_builder{
			Name: lo.ToPtr("test-service"),
			HttpService: configv1.HttpUpstreamService_builder{
				Address: &serverAddr,
				HealthCheck: configv1.HttpHealthCheck_builder{
					Url:          &invalidURL,
					ExpectedCode: lo.ToPtr(int32(http.StatusOK)),
				}.Build(),
			}.Build(),
		}.Build()

		checker := NewChecker(upstreamConfig)
		assert.NotNil(t, checker)
		assert.Equal(t, health.StatusDown, checker.Check(ctx).Status)
	})
}

func TestCheckGRPCHealth(t *testing.T) {
	ctx := context.Background()

	t.Run("Serving", func(t *testing.T) {
		server, lis := newMockGRPCHealthServer(grpc_health_v1.HealthCheckResponse_SERVING)
		defer server.Stop()

		upstreamConfig := configv1.UpstreamServiceConfig_builder{
			Name: lo.ToPtr("test-service"),
			GrpcService: configv1.GrpcUpstreamService_builder{
				Address: lo.ToPtr(lis.Addr().String()),
				HealthCheck: configv1.GrpcHealthCheck_builder{
					Service: lo.ToPtr("test-service"),
				}.Build(),
			}.Build(),
		}.Build()

		checker := NewChecker(upstreamConfig)
		assert.NotNil(t, checker)
		assert.Equal(t, health.StatusUp, checker.Check(ctx).Status)
	})

	t.Run("NotServing", func(t *testing.T) {
		server, lis := newMockGRPCHealthServer(grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		defer server.Stop()

		upstreamConfig := configv1.UpstreamServiceConfig_builder{
			Name: lo.ToPtr("test-service"),
			GrpcService: configv1.GrpcUpstreamService_builder{
				Address: lo.ToPtr(lis.Addr().String()),
				HealthCheck: configv1.GrpcHealthCheck_builder{
					Service: lo.ToPtr("test-service"),
				}.Build(),
			}.Build(),
		}.Build()

		checker := NewChecker(upstreamConfig)
		assert.NotNil(t, checker)
		assert.Equal(t, health.StatusDown, checker.Check(ctx).Status)
	})

	t.Run("ServerUnreachable", func(t *testing.T) {
		upstreamConfig := configv1.UpstreamServiceConfig_builder{
			Name: lo.ToPtr("test-service"),
			GrpcService: configv1.GrpcUpstreamService_builder{
				Address: lo.ToPtr("localhost:12345"),
				HealthCheck: configv1.GrpcHealthCheck_builder{
					Service: lo.ToPtr("test-service"),
				}.Build(),
			}.Build(),
		}.Build()

		checker := NewChecker(upstreamConfig)
		assert.NotNil(t, checker)
		assert.Equal(t, health.StatusDown, checker.Check(ctx).Status)
	})
}

func TestCheckConnection(t *testing.T) {
	t.Run("ConnectionSuccess", func(t *testing.T) {
		lis, err := net.Listen("tcp", "localhost:0")
		assert.NoError(t, err)
		defer lis.Close()
		assert.NoError(t, checkConnection(lis.Addr().String()), "checkConnection should succeed for a listening port")
	})

	t.Run("ConnectionFailure", func(t *testing.T) {
		assert.Error(t, checkConnection("localhost:12345"), "checkConnection should fail for a non-listening port")
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
		want   health.AvailabilityStatus
	}{
		{
			name: "OpenAPI Service",
			config: configv1.UpstreamServiceConfig_builder{
				Name:           lo.ToPtr("openapi-service"),
				OpenapiService: configv1.OpenapiUpstreamService_builder{Address: &addr}.Build(),
			}.Build(),
			want: health.StatusUp,
		},
		{
			name: "Command Line Service with Health Check",
			config: configv1.UpstreamServiceConfig_builder{
				Name: lo.ToPtr("cmd-service-with-health-check"),
				CommandLineService: configv1.CommandLineUpstreamService_builder{
					HealthCheck: &configv1.CommandLineHealthCheck{},
				}.Build(),
			}.Build(),
			want: health.StatusUp,
		},
		{
			name: "WebSocket Service",
			config: configv1.UpstreamServiceConfig_builder{
				Name:             lo.ToPtr("websocket-service"),
				WebsocketService: configv1.WebsocketUpstreamService_builder{Address: &addr}.Build(),
			}.Build(),
			want: health.StatusUp,
		},
		{
			name: "WebRTC Service",
			config: configv1.UpstreamServiceConfig_builder{
				Name:          lo.ToPtr("webrtc-service"),
				WebrtcService: configv1.WebrtcUpstreamService_builder{Address: &addr}.Build(),
			}.Build(),
			want: health.StatusUp,
		},
		{
			name: "MCP Service HTTP",
			config: configv1.UpstreamServiceConfig_builder{
				Name: lo.ToPtr("mcp-http-service"),
				McpService: configv1.McpUpstreamService_builder{
					HttpConnection: configv1.McpStreamableHttpConnection_builder{HttpAddress: &addr}.Build(),
				}.Build(),
			}.Build(),
			want: health.StatusUp,
		},
		{
			name: "MCP Service Stdio",
			config: configv1.UpstreamServiceConfig_builder{
				Name: lo.ToPtr("mcp-stdio-service"),
				McpService: configv1.McpUpstreamService_builder{
					StdioConnection: configv1.McpStdioConnection_builder{Command: lo.ToPtr("echo")}.Build(),
				}.Build(),
			}.Build(),
			want: health.StatusUp,
		},
		{
			name: "Command Line Service",
			config: configv1.UpstreamServiceConfig_builder{
				Name:               lo.ToPtr("cmd-service"),
				CommandLineService: (&configv1.CommandLineUpstreamService_builder{}).Build(),
			}.Build(),
			want: health.StatusUp,
		},
		{
			name: "MCP Service No Connection",
			config: configv1.UpstreamServiceConfig_builder{
				Name:       lo.ToPtr("mcp-no-connection"),
				McpService: &configv1.McpUpstreamService{},
			}.Build(),
			want: health.StatusDown,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			checker := NewChecker(tc.config)
			assert.NotNil(t, checker)
			assert.Equal(t, tc.want, checker.Check(ctx).Status)
		})
	}
}

func TestCommandLineHealthCheck(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name   string
		config *configv1.CommandLineUpstreamService
		want   health.AvailabilityStatus
	}{
		{
			name: "Successful Health Check",
			config: configv1.CommandLineUpstreamService_builder{
				HealthCheck: configv1.CommandLineHealthCheck_builder{
					Method:                     lo.ToPtr("echo health_check"),
					ExpectedResponseContains: lo.ToPtr("health_check"),
				}.Build(),
			}.Build(),
			want: health.StatusUp,
		},
		{
			name: "Failed Health Check - Command Fails",
			config: configv1.CommandLineUpstreamService_builder{
				HealthCheck: configv1.CommandLineHealthCheck_builder{
					Method: lo.ToPtr("false"), // This command always fails
				}.Build(),
			}.Build(),
			want: health.StatusDown,
		},
		{
			name: "Failed Health Check - Unexpected Output",
			config: configv1.CommandLineUpstreamService_builder{
				HealthCheck: configv1.CommandLineHealthCheck_builder{
					Method:                     lo.ToPtr("echo"),
					Prompt:                     lo.ToPtr("wrong_output"),
					ExpectedResponseContains: lo.ToPtr("expected_output"),
				}.Build(),
			}.Build(),
			want: health.StatusDown,
		},
		{
			name: "Health Check Timeout",
			config: configv1.CommandLineUpstreamService_builder{
				HealthCheck: configv1.CommandLineHealthCheck_builder{
					Method:  lo.ToPtr("sleep 2"),
					Timeout: durationpb.New(1 * time.Second),
				}.Build(),
			}.Build(),
			want: health.StatusDown,
		},
		{
			name: "No Health Check Configured",
			config: configv1.CommandLineUpstreamService_builder{
				HealthCheck: nil,
			}.Build(),
			want: health.StatusUp,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			check := commandLineCheck("test-service", tc.config)
			result := check.Check(ctx)
			if tc.want == health.StatusUp {
				assert.NoError(t, result)
			} else {
				assert.Error(t, result)
			}
		})
	}
}
