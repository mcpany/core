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
	"net/url"
	"testing"

	configv1 "github.com/mcpxy/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// mockHttpService is a mock implementation of HTTPServiceWithHealthCheck for testing.
type mockHttpService struct {
	address     string
	healthCheck *configv1.HttpHealthCheck
}

func (m *mockHttpService) GetAddress() string {
	return m.address
}

func (m *mockHttpService) GetHealthCheck() *configv1.HttpHealthCheck {
	return m.healthCheck
}

// TestCheck is the main test function for the Check function.
func TestCheck(t *testing.T) {
	ctx := context.Background()

	t.Run("nil config", func(t *testing.T) {
		assert.False(t, Check(ctx, nil))
	})

	t.Run("http service", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)

		cfg := &configv1.UpstreamServiceConfig{}
		httpService := &configv1.HttpUpstreamService{}
		httpService.SetAddress(u.Host)
		healthCheck := &configv1.HttpHealthCheck{}
		healthCheck.SetUrl(server.URL)
		healthCheck.SetExpectedCode(http.StatusOK)
		httpService.SetHealthCheck(healthCheck)
		cfg.SetHttpService(httpService)
		assert.True(t, Check(ctx, cfg))
	})

	t.Run("grpc service", func(t *testing.T) {
		addr, stop := createMockGRPCServer(t, healthpb.HealthCheckResponse_SERVING)
		defer stop()

		cfg := &configv1.UpstreamServiceConfig{}
		grpcService := &configv1.GrpcUpstreamService{}
		grpcService.SetAddress(addr)
		healthCheck := &configv1.GrpcHealthCheck{}
		healthCheck.SetService("test")
		grpcService.SetHealthCheck(healthCheck)
		cfg.SetGrpcService(grpcService)
		assert.True(t, Check(ctx, cfg))
	})

	t.Run("openapi service", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)

		cfg := &configv1.UpstreamServiceConfig{}
		openapiService := &configv1.OpenapiUpstreamService{}
		openapiService.SetAddress(u.Host)
		healthCheck := &configv1.HttpHealthCheck{}
		healthCheck.SetUrl(server.URL)
		healthCheck.SetExpectedCode(http.StatusOK)
		openapiService.SetHealthCheck(healthCheck)
		cfg.SetOpenapiService(openapiService)

		assert.True(t, Check(ctx, cfg))
	})

	t.Run("command line service", func(t *testing.T) {
		cfg := &configv1.UpstreamServiceConfig{}
		cfg.SetCommandLineService(&configv1.CommandLineUpstreamService{})
		assert.True(t, Check(ctx, cfg))
	})

	t.Run("websocket service", func(t *testing.T) {
		addr, stop := createMockTCPServer(t)
		defer stop()

		cfg := &configv1.UpstreamServiceConfig{}
		wsService := &configv1.WebsocketUpstreamService{}
		wsService.SetAddress(addr)
		cfg.SetWebsocketService(wsService)
		assert.True(t, Check(ctx, cfg))
	})

	t.Run("webrtc service", func(t *testing.T) {
		addr, stop := createMockTCPServer(t)
		defer stop()

		cfg := &configv1.UpstreamServiceConfig{}
		webrtcService := &configv1.WebrtcUpstreamService{}
		webrtcService.SetAddress(addr)
		cfg.SetWebrtcService(webrtcService)
		assert.True(t, Check(ctx, cfg))
	})

	t.Run("mcp service http", func(t *testing.T) {
		addr, stop := createMockTCPServer(t)
		defer stop()
		cfg := &configv1.UpstreamServiceConfig{}
		mcpService := &configv1.McpUpstreamService{}
		httpConn := &configv1.McpStreamableHttpConnection{}
		httpConn.SetHttpAddress(addr)
		mcpService.SetHttpConnection(httpConn)
		cfg.SetMcpService(mcpService)
		assert.True(t, Check(ctx, cfg))
	})

	t.Run("mcp service stdio", func(t *testing.T) {
		cfg := &configv1.UpstreamServiceConfig{}
		mcpService := &configv1.McpUpstreamService{}
		mcpService.SetStdioConnection(&configv1.McpStdioConnection{})
		cfg.SetMcpService(mcpService)
		assert.True(t, Check(ctx, cfg))
	})

	t.Run("no service", func(t *testing.T) {
		cfg := &configv1.UpstreamServiceConfig{}
		assert.False(t, Check(ctx, cfg))
	})
}

// TestCheckHTTPHealth tests the checkHTTPHealth function.
func TestCheckHTTPHealth(t *testing.T) {
	ctx := context.Background()

	t.Run("successful health check", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()
		u, _ := url.Parse(server.URL)

		healthCheck := &configv1.HttpHealthCheck{}
		healthCheck.SetUrl(server.URL)
		healthCheck.SetExpectedCode(http.StatusOK)
		service := &mockHttpService{
			address:     u.Host,
			healthCheck: healthCheck,
		}
		assert.True(t, checkHTTPHealth(ctx, service))
	})

	t.Run("failed health check with wrong status code", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()
		u, _ := url.Parse(server.URL)

		healthCheck := &configv1.HttpHealthCheck{}
		healthCheck.SetUrl(server.URL)
		healthCheck.SetExpectedCode(http.StatusOK)
		service := &mockHttpService{
			address:     u.Host,
			healthCheck: healthCheck,
		}
		assert.False(t, checkHTTPHealth(ctx, service))
	})

	t.Run("failed health check with server down", func(t *testing.T) {
		healthCheck := &configv1.HttpHealthCheck{}
		healthCheck.SetUrl("http://localhost:12345")
		healthCheck.SetExpectedCode(http.StatusOK)
		service := &mockHttpService{
			address:     "localhost:12345",
			healthCheck: healthCheck,
		}
		assert.False(t, checkHTTPHealth(ctx, service))
	})

	t.Run("no health check config", func(t *testing.T) {
		addr, stop := createMockTCPServer(t)
		defer stop()
		service := &mockHttpService{address: addr}
		assert.True(t, checkHTTPHealth(ctx, service))
	})

	t.Run("no health check config and server down", func(t *testing.T) {
		service := &mockHttpService{address: "localhost:12345"}
		assert.False(t, checkHTTPHealth(ctx, service))
	})

	t.Run("invalid url", func(t *testing.T) {
		healthCheck := &configv1.HttpHealthCheck{}
		healthCheck.SetUrl("invalid-url")
		service := &mockHttpService{
			address:     "localhost:12345",
			healthCheck: healthCheck,
		}
		assert.False(t, checkHTTPHealth(ctx, service))
	})
}

// TestCheckGRPCHealth tests the checkGRPCHealth function.
func TestCheckGRPCHealth(t *testing.T) {
	ctx := context.Background()

	t.Run("successful health check", func(t *testing.T) {
		addr, stop := createMockGRPCServer(t, healthpb.HealthCheckResponse_SERVING)
		defer stop()
		service := &configv1.GrpcUpstreamService{}
		service.SetAddress(addr)
		healthCheck := &configv1.GrpcHealthCheck{}
		healthCheck.SetService("test")
		service.SetHealthCheck(healthCheck)
		assert.True(t, checkGRPCHealth(ctx, service))
	})

	t.Run("failed health check with not serving status", func(t *testing.T) {
		addr, stop := createMockGRPCServer(t, healthpb.HealthCheckResponse_NOT_SERVING)
		defer stop()
		service := &configv1.GrpcUpstreamService{}
		service.SetAddress(addr)
		healthCheck := &configv1.GrpcHealthCheck{}
		healthCheck.SetService("test")
		service.SetHealthCheck(healthCheck)
		assert.False(t, checkGRPCHealth(ctx, service))
	})

	t.Run("failed health check with server down", func(t *testing.T) {
		service := &configv1.GrpcUpstreamService{}
		service.SetAddress("localhost:12345")
		healthCheck := &configv1.GrpcHealthCheck{}
		healthCheck.SetService("test")
		service.SetHealthCheck(healthCheck)
		assert.False(t, checkGRPCHealth(ctx, service))
	})

	t.Run("no health check config", func(t *testing.T) {
		addr, stop := createMockTCPServer(t)
		defer stop()
		service := &configv1.GrpcUpstreamService{}
		service.SetAddress(addr)
		assert.True(t, checkGRPCHealth(ctx, service))
	})
}

// createMockGRPCServer creates a mock gRPC server for testing.
func createMockGRPCServer(t *testing.T, status healthpb.HealthCheckResponse_ServingStatus) (string, func()) {
	t.Helper()
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer()
	healthServer := health.NewServer()
	healthServer.SetServingStatus("test", status)
	healthpb.RegisterHealthServer(s, healthServer)
	go func() {
		if err := s.Serve(lis); err != nil {
			// not an error
		}
	}()
	return lis.Addr().String(), func() {
		s.Stop()
		lis.Close()
	}
}

// createMockTCPServer creates a mock TCP server for testing connection checks.
func createMockTCPServer(t *testing.T) (string, func()) {
	t.Helper()
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	go func() {
		for {
			conn, err := lis.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()
	return lis.Addr().String(), func() {
		lis.Close()
	}
}
