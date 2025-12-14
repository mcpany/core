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
	"strings"
	"testing"
	"time"

	"github.com/alexliesenfeld/health"
	"github.com/coder/websocket"
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

func (s *mockHealthServer) Check(
	_ context.Context,
	_ *grpc_health_v1.HealthCheckRequest,
) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{Status: s.status}, nil
}

func (s *mockHealthServer) Watch(
	_ *grpc_health_v1.HealthCheckRequest,
	srv grpc_health_v1.Health_WatchServer,
) error {
	return srv.Send(&grpc_health_v1.HealthCheckResponse{Status: s.status})
}

// newMockGRPCHealthServer starts a gRPC server with the mock health service.
func newMockGRPCHealthServer(
	status grpc_health_v1.HealthCheckResponse_ServingStatus,
) (*grpc.Server, net.Listener) {
	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	grpc_health_v1.RegisterHealthServer(s, &mockHealthServer{status: status})
	go func() { _ = s.Serve(lis) }()
	return s, lis
}



func TestNewChecker(t *testing.T) {
	ctx := context.Background()

	t.Run("NilConfig", func(t *testing.T) {
		assert.Nil(t, NewChecker(nil), "NewChecker with nil config should return nil")
	})

	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		addr := server.Listener.Addr().String()
		upstreamConfig := configv1.UpstreamServiceConfig_builder{
			Name:        lo.ToPtr("test-service"),
			HttpService: configv1.HttpUpstreamService_builder{Address: &addr}.Build(),
		}.Build()

		checker := NewChecker(upstreamConfig)
		assert.NotNil(t, checker)
		assert.Equal(t, health.StatusUp, checker.Check(ctx).Status)
	})

	t.Run("Failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		addr := server.Listener.Addr().String()
		upstreamConfig := configv1.UpstreamServiceConfig_builder{
			Name: lo.ToPtr("test-service"),
			HttpService: configv1.HttpUpstreamService_builder{
				Address: &addr,
				HealthCheck: configv1.HttpHealthCheck_builder{
					Url:          lo.ToPtr(server.URL),
					ExpectedCode: lo.ToPtr(int32(http.StatusOK)),
				}.Build(),
			}.Build(),
		}.Build()

		checker := NewChecker(upstreamConfig)
		assert.NotNil(t, checker)
		assert.Equal(t, health.StatusDown, checker.Check(ctx).Status)
	})

	t.Run("WriteOnly", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, err := websocket.Accept(w, r, nil)
			if err != nil {
				return
			}
			defer func() { _ = conn.Close(websocket.StatusInternalError, "") }()

			_, msg, err := conn.Read(r.Context())
			if err != nil {
				return
			}

			if string(msg) == "ping" {
				// Do nothing, just close the connection
				return
			}
		}))
		defer server.Close()

		addr := strings.TrimPrefix(server.URL, "http://")
		upstreamConfig := configv1.UpstreamServiceConfig_builder{
			Name: lo.ToPtr("websocket-service"),
			WebsocketService: configv1.WebsocketUpstreamService_builder{
				Address: &addr,
				HealthCheck: configv1.WebsocketHealthCheck_builder{
					Message: lo.ToPtr("ping"),
				}.Build(),
			}.Build(),
		}.Build()

		checker := NewChecker(upstreamConfig)
		assert.NotNil(t, checker)
		assert.Equal(t, health.StatusUp, checker.Check(ctx).Status)
	})

	t.Run("HTTPHealthCheckWithCustomPOSTMethod", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		serverURL := server.URL
		serverAddr := server.Listener.Addr().String()

		upstreamConfig := configv1.UpstreamServiceConfig_builder{
			Name: lo.ToPtr("test-service-post"),
			HttpService: configv1.HttpUpstreamService_builder{
				Address: &serverAddr,
				HealthCheck: configv1.HttpHealthCheck_builder{
					Url:          &serverURL,
					Method:       lo.ToPtr("POST"),
					ExpectedCode: lo.ToPtr(int32(http.StatusOK)),
				}.Build(),
			}.Build(),
		}.Build()

		checker := NewChecker(upstreamConfig)
		assert.NotNil(t, checker)
		result := checker.Check(ctx)
		assert.Equal(t, health.StatusUp, result.Status)
	})

	t.Run("WebRTC Service WebSocket Health Check Success", func(t *testing.T) {
		// Mock WebSocket server
		mockWSServer := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c, err := websocket.Accept(w, r, nil)
				if err != nil {
					return
				}
				defer func() { _ = c.Close(websocket.StatusInternalError, "the sky is falling") }()
			}),
		)
		defer mockWSServer.Close()
		wsURL := "ws" + strings.TrimPrefix(mockWSServer.URL, "http")

		config := (&configv1.UpstreamServiceConfig_builder{
			Name: lo.ToPtr("webrtc-service-ws-healthy"),
			WebrtcService: (&configv1.WebrtcUpstreamService_builder{
				Address: lo.ToPtr(mockWSServer.Listener.Addr().String()),
				HealthCheck: (&configv1.WebRTCHealthCheck_builder{
					Websocket: (&configv1.WebsocketHealthCheck_builder{
						Url: lo.ToPtr(wsURL),
					}).Build(),
				}).Build(),
			}).Build(),
		}).Build()
		checker := NewChecker(config)
		assert.NotNil(t, checker)
		assert.Equal(t, health.StatusUp, checker.Check(ctx).Status)
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
		defer func() { _ = lis.Close() }()
		assert.NoError(
			t,
			checkConnection(lis.Addr().String()),
			"checkConnection should succeed for a listening port",
		)
	})

	t.Run("ConnectionFailure", func(t *testing.T) {
		assert.Error(
			t,
			checkConnection("localhost:12345"),
			"checkConnection should fail for a non-listening port",
		)
	})
}

func TestCheckVariousServices(t *testing.T) {
	ctx := context.Background()

	// Setup a simple listening server for connection checks
	lis, err := net.Listen("tcp", "localhost:0")
	assert.NoError(t, err)
	defer func() { _ = lis.Close() }()
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
			name: "Command Line Service Health Check Success",
			config: configv1.UpstreamServiceConfig_builder{
				Name: lo.ToPtr("cmd-service-health-check-success"),
				CommandLineService: configv1.CommandLineUpstreamService_builder{
					Command: lo.ToPtr("echo"),
					HealthCheck: configv1.CommandLineHealthCheck_builder{
						Method:                   lo.ToPtr("hello"),
						ExpectedResponseContains: lo.ToPtr("hello"),
					}.Build(),
				}.Build(),
			}.Build(),
			want: health.StatusUp,
		},
		{
			name: "Command Line Service Health Check Failure",
			config: configv1.UpstreamServiceConfig_builder{
				Name: lo.ToPtr("cmd-service-health-check-failure"),
				CommandLineService: configv1.CommandLineUpstreamService_builder{
					Command: lo.ToPtr("echo"),
					HealthCheck: configv1.CommandLineHealthCheck_builder{
						Method:                   lo.ToPtr("hello"),
						ExpectedResponseContains: lo.ToPtr("world"),
					}.Build(),
				}.Build(),
			}.Build(),
			want: health.StatusDown,
		},
		{
			name: "Command Line Service Health Check Non-Zero Exit Code",
			config: configv1.UpstreamServiceConfig_builder{
				Name: lo.ToPtr("cmd-service-health-check-non-zero-exit"),
				CommandLineService: configv1.CommandLineUpstreamService_builder{
					Command: lo.ToPtr("false"),
					HealthCheck: configv1.CommandLineHealthCheck_builder{
						Method: lo.ToPtr(""),
					}.Build(),
				}.Build(),
			}.Build(),
			want: health.StatusDown,
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
			name: "WebSocket Service Unreachable",
			config: configv1.UpstreamServiceConfig_builder{
				Name: lo.ToPtr("websocket-service-unreachable"),
				WebsocketService: configv1.WebsocketUpstreamService_builder{
					Address: lo.ToPtr("localhost:12345"),
				}.Build(),
			}.Build(),
			want: health.StatusDown,
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
			name: "WebRTC Service Unreachable",
			config: configv1.UpstreamServiceConfig_builder{
				Name:          lo.ToPtr("webrtc-service-unreachable"),
				WebrtcService: configv1.WebrtcUpstreamService_builder{Address: lo.ToPtr("localhost:12345")}.Build(),
			}.Build(),
			want: health.StatusDown,
		},
		{
			name: "MCP Service HTTP",
			config: configv1.UpstreamServiceConfig_builder{
				Name: lo.ToPtr("mcp-http-service"),
				McpService: configv1.McpUpstreamService_builder{
					HttpConnection: configv1.McpStreamableHttpConnection_builder{
						HttpAddress: &addr,
					}.Build(),
				}.Build(),
			}.Build(),
			want: health.StatusUp,
		},
		{
			name: "MCP Service Stdio",
			config: configv1.UpstreamServiceConfig_builder{
				Name: lo.ToPtr("mcp-stdio-service"),
				McpService: configv1.McpUpstreamService_builder{
					StdioConnection: configv1.McpStdioConnection_builder{
						Command: lo.ToPtr("echo"),
					}.Build(),
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

func TestWebsocketCheck(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, err := websocket.Accept(w, r, nil)
			if err != nil {
				return
			}
			defer func() { _ = conn.Close(websocket.StatusInternalError, "") }()

			_, msg, err := conn.Read(r.Context())
			if err != nil {
				return
			}

			if string(msg) == "ping" {
				_ = conn.Write(r.Context(), websocket.MessageText, []byte("pong"))
			}
		}))
		defer server.Close()

		addr := strings.TrimPrefix(server.URL, "http://")
		upstreamConfig := configv1.UpstreamServiceConfig_builder{
			Name: lo.ToPtr("websocket-service"),
			WebsocketService: configv1.WebsocketUpstreamService_builder{
				Address: &addr,
				HealthCheck: configv1.WebsocketHealthCheck_builder{
					Message:                  lo.ToPtr("ping"),
					ExpectedResponseContains: lo.ToPtr("pong"),
				}.Build(),
			}.Build(),
		}.Build()

		checker := NewChecker(upstreamConfig)
		assert.NotNil(t, checker)
		assert.Equal(t, health.StatusUp, checker.Check(ctx).Status)
	})

	t.Run("ResponseMismatch", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, err := websocket.Accept(w, r, nil)
			if err != nil {
				return
			}
			defer func() { _ = conn.Close(websocket.StatusInternalError, "") }()

			_ = conn.Write(r.Context(), websocket.MessageText, []byte("unexpected"))
		}))
		defer server.Close()

		addr := strings.TrimPrefix(server.URL, "http://")
		upstreamConfig := configv1.UpstreamServiceConfig_builder{
			Name: lo.ToPtr("websocket-service"),
			WebsocketService: configv1.WebsocketUpstreamService_builder{
				Address: &addr,
				HealthCheck: configv1.WebsocketHealthCheck_builder{
					ExpectedResponseContains: lo.ToPtr("pong"),
				}.Build(),
			}.Build(),
		}.Build()

		checker := NewChecker(upstreamConfig)
		assert.NotNil(t, checker)
		assert.Equal(t, health.StatusDown, checker.Check(ctx).Status)
	})

	t.Run("ServerUnreachable", func(t *testing.T) {
		addr := "localhost:12345"
		upstreamConfig := configv1.UpstreamServiceConfig_builder{
			Name: lo.ToPtr("websocket-service"),
			WebsocketService: configv1.WebsocketUpstreamService_builder{
				Address: &addr,
				HealthCheck: configv1.WebsocketHealthCheck_builder{
					Timeout: durationpb.New(10 * time.Millisecond),
				}.Build(),
			}.Build(),
		}.Build()

		checker := NewChecker(upstreamConfig)
		assert.NotNil(t, checker)
		assert.Equal(t, health.StatusDown, checker.Check(ctx).Status)
	})

	t.Run("NilHealthCheck", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, err := websocket.Accept(w, r, nil)
			if err != nil {
				return
			}
			_ = conn.Close(websocket.StatusNormalClosure, "")
		}))
		defer server.Close()

		addr := strings.TrimPrefix(server.URL, "http://")
		upstreamConfig := configv1.UpstreamServiceConfig_builder{
			Name:             lo.ToPtr("websocket-service"),
			WebsocketService: configv1.WebsocketUpstreamService_builder{Address: &addr}.Build(),
		}.Build()

		checker := NewChecker(upstreamConfig)
		assert.NotNil(t, checker)
		assert.Equal(t, health.StatusUp, checker.Check(ctx).Status)
	})
}

func TestWebSocketHealthCheckBasic(t *testing.T) {
	ctx := context.Background()

	// Mock WebSocket server
	mockServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := websocket.Accept(w, r, nil)
			if err != nil {
				return
			}
			defer func() { _ = c.Close(websocket.StatusInternalError, "the sky is falling") }()
		}),
	)
	defer mockServer.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(mockServer.URL, "http")

	t.Run("WebSocket Service Health Check Success", func(t *testing.T) {
		config := (&configv1.UpstreamServiceConfig_builder{
			Name: lo.ToPtr("websocket-service-healthy"),
			WebsocketService: (&configv1.WebsocketUpstreamService_builder{
				Address: lo.ToPtr(mockServer.Listener.Addr().String()),
				HealthCheck: (&configv1.WebsocketHealthCheck_builder{
					Url: lo.ToPtr(wsURL),
				}).Build(),
			}).Build(),
		}).Build()
		checker := NewChecker(config)
		assert.NotNil(t, checker)
		assert.Equal(t, health.StatusUp, checker.Check(ctx).Status)
	})

	t.Run("WebSocket Service Health Check Failure", func(t *testing.T) {
		config := (&configv1.UpstreamServiceConfig_builder{
			Name: lo.ToPtr("websocket-service-unhealthy"),
			WebsocketService: (&configv1.WebsocketUpstreamService_builder{
				Address: lo.ToPtr("localhost:12345"),
				HealthCheck: (&configv1.WebsocketHealthCheck_builder{
					Url: lo.ToPtr("ws://localhost:12345"),
				}).Build(),
			}).Build(),
		}).Build()
		checker := NewChecker(config)
		assert.NotNil(t, checker)
		assert.Equal(t, health.StatusDown, checker.Check(ctx).Status)
	})
}

func TestWebRTCHealthCheck(t *testing.T) {
	ctx := context.Background()

	// Mock HTTP server for signaling
	mockSignalingServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)
	defer mockSignalingServer.Close()

	t.Run("WebRTC Service Health Check Success", func(t *testing.T) {
		config := (&configv1.UpstreamServiceConfig_builder{
			Name: lo.ToPtr("webrtc-service-healthy"),
			WebrtcService: (&configv1.WebrtcUpstreamService_builder{
				Address: lo.ToPtr(mockSignalingServer.Listener.Addr().String()),
				HealthCheck: (&configv1.WebRTCHealthCheck_builder{
					Http: (&configv1.HttpHealthCheck_builder{
						Url:          lo.ToPtr(mockSignalingServer.URL),
						ExpectedCode: lo.ToPtr(int32(http.StatusOK)),
					}).Build(),
				}).Build(),
			}).Build(),
		}).Build()
		checker := NewChecker(config)
		assert.NotNil(t, checker)
		assert.Equal(t, health.StatusUp, checker.Check(ctx).Status)
	})

	t.Run("WebRTC Service Health Check Failure", func(t *testing.T) {
		config := (&configv1.UpstreamServiceConfig_builder{
			Name: lo.ToPtr("webrtc-service-unhealthy"),
			WebrtcService: (&configv1.WebrtcUpstreamService_builder{
				Address: lo.ToPtr("localhost:12345"),
				HealthCheck: (&configv1.WebRTCHealthCheck_builder{
					Http: (&configv1.HttpHealthCheck_builder{
						Url: lo.ToPtr("http://localhost:12345"),
					}).Build(),
				}).Build(),
			}).Build(),
		}).Build()
		checker := NewChecker(config)
		assert.NotNil(t, checker)
		assert.Equal(t, health.StatusDown, checker.Check(ctx).Status)
	})
}
