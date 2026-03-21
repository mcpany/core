// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestHTTPClientWrapper(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	configJSON := `{"http_service": {"address": "` + server.URL[7:] + `"}}`
	config := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), config))

	client := NewHTTPClientWrapper(&http.Client{}, config, nil)
	assert.NotNil(t, client)

	t.Run("IsHealthy", func(t *testing.T) {
		assert.True(t, client.IsHealthy(context.Background()))
	})

	t.Run("Close", func(t *testing.T) {
		err := client.Close()
		assert.NoError(t, err)
	})
	t.Run("IsHealthy_WithCheck", func(t *testing.T) {
		// Valid health check
		configJSON := `{"http_service": {"address": "` + server.URL[7:] + `", "health_check": {"url": "` + server.URL + `"}}, "name": "foo"}`
		config := &configv1.UpstreamServiceConfig{}
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), config))

		// IMPORTANT: NewChecker defaults expected code to 0 if not set? No, proto default.
		// Use builder to be sure or set ExpectedCode in JSON if needed?
		// But in JSON "expected_code" defaults to 0?
		// HttpHealthCheck expected_code field is int32.
		// If 0, my previous check failed because of 0 != 200.
		// In JSON I can set it.
		configJSON = `{"http_service": {"address": "` + server.URL[7:] + `", "health_check": {"url": "` + server.URL + `", "expected_code": 200}}, "name": "foo"}`
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), config))

		client := NewHTTPClientWrapper(&http.Client{}, config, nil)
		assert.True(t, client.IsHealthy(context.Background()))
	})

	t.Run("IsHealthy_Failure", func(t *testing.T) {
		// Create a config with a health check that fails (unreachable URL)
		badConfigJSON := `{"http_service": {"address": "` + server.URL[7:] + `", "health_check": {"url": "http://127.0.0.1:12345/health"}}, "name": "bar"}`
		badConfig := &configv1.UpstreamServiceConfig{}
		require.NoError(t, protojson.Unmarshal([]byte(badConfigJSON), badConfig))

		badClient := NewHTTPClientWrapper(&http.Client{}, badConfig, nil)
		assert.False(t, badClient.IsHealthy(context.Background()))
	})
	t.Run("IsHealthy_NilConfig", func(t *testing.T) {
		client := NewHTTPClientWrapper(&http.Client{}, nil, nil)
		assert.True(t, client.IsHealthy(context.Background()))
	})

	t.Run("IsHealthy_NoServiceConfig", func(t *testing.T) {
		client := NewHTTPClientWrapper(&http.Client{}, &configv1.UpstreamServiceConfig{}, nil)
		assert.True(t, client.IsHealthy(context.Background()))
	})
}

func TestGrpcClientWrapper(t *testing.T) {
	// Set up a dummy gRPC server to connect to
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	server := grpc.NewServer()
	go func() {
		// This may return an error on server.Stop(), which is expected.
		_ = server.Serve(lis)
	}()
	defer server.Stop()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	configJSON := `{"grpc_service": {"address": "` + lis.Addr().String() + `"}}`
	config := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), config))

	wrapper := NewGrpcClientWrapper(conn, config, nil)
	assert.NotNil(t, wrapper)

	t.Run("IsHealthy_Initially", func(t *testing.T) {
		// Wait for the connection to be ready or idle. This is a more robust check.
		require.Eventually(t, func() bool {
			state := wrapper.GetState()
			return state == connectivity.Ready || state == connectivity.Idle
		}, time.Second*5, 10*time.Millisecond, "gRPC client should connect")
		assert.True(t, wrapper.IsHealthy(context.Background()))
	})

	t.Run("IsHealthy_Failure", func(t *testing.T) {
		// Create a config with a gRPC health check that fails (service name mismatch or unreachable)
		// We use NewGrpcClientWrapper with a new config but SAME connection (which is working).
		// But we configure a health check that checks a non-existent service.
		badConfigJSON := `{"grpc_service": {"address": "` + lis.Addr().String() + `", "health_check": {"service": "non-existent"}}}`
		badConfig := &configv1.UpstreamServiceConfig{}
		require.NoError(t, protojson.Unmarshal([]byte(badConfigJSON), badConfig))

		badWrapper := NewGrpcClientWrapper(conn, badConfig, nil) // Reuse connection
		// gRPC health check should fail because "non-existent" service is not registered in health server?
		// Wait, our dummy server does NOT register health server at all!
		// So ANY health check check call will fail with Unimplemented?
		// NewChecker uses grpc_health_v1.NewHealthClient.
		// If server doesn't implement it, it fails.
		// So IsHealthy should return false.

		assert.False(t, badWrapper.IsHealthy(context.Background()))
	})

	t.Run("Close and IsHealthy", func(t *testing.T) {
		err := wrapper.Close()
		require.NoError(t, err)

		// The state should eventually become Shutdown.
		require.Eventually(t, func() bool {
			return wrapper.GetState() == connectivity.Shutdown
		}, time.Second, 10*time.Millisecond, "gRPC client state should be Shutdown")

		assert.False(t, wrapper.IsHealthy(context.Background()))
	})

	t.Run("IsHealthy_Bufnet", func(t *testing.T) {
		// Set up a new dummy gRPC server to avoid using the closed connection from the previous test
		lis, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		server := grpc.NewServer()
		go func() {
			_ = server.Serve(lis)
		}()
		defer server.Stop()

		conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)
		defer func() { _ = conn.Close() }()

		configJSON := `{"grpc_service": {"address": "bufnet"}}`
		config := &configv1.UpstreamServiceConfig{}
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), config))

		wrapper := NewGrpcClientWrapper(conn, config, nil)
		require.Eventually(t, func() bool {
			state := wrapper.GetState()
			return state == connectivity.Ready || state == connectivity.Idle
		}, time.Second*5, 10*time.Millisecond, "gRPC client should connect")
		assert.True(t, wrapper.IsHealthy(context.Background()))
	})
}

func TestWebsocketClientWrapper(t *testing.T) {
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		// Just handle control messages for the health check
		_ = conn.SetReadDeadline(time.Now().Add(time.Second * 5))
		for {
			if _, _, err := conn.NextReader(); err != nil {
				break
			}
		}
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[len("http"):]
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer func() {
		err := resp.Body.Close()
		require.NoError(t, err)
	}()

	wrapper := &WebsocketClientWrapper{Conn: conn}

	t.Run("IsHealthy_Connected", func(t *testing.T) {
		assert.True(t, wrapper.IsHealthy(context.Background()))
	})

	t.Run("Close and IsHealthy", func(t *testing.T) {
		err := wrapper.Close()
		require.NoError(t, err)
		// After closing, IsHealthy should fail
		assert.False(t, wrapper.IsHealthy(context.Background()))
	})
}

func TestGrpcClientWrapper_WithHealthCheck(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	healthServer := health.NewServer()
	server := grpc.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)

	go func() { _ = server.Serve(lis) }()
	defer server.Stop()

	// construct config
	config := configv1.UpstreamServiceConfig_builder{
		Name: lo.ToPtr("grpc-service"),
		GrpcService: configv1.GrpcUpstreamService_builder{
			Address: lo.ToPtr(lis.Addr().String()),
			HealthCheck: configv1.GrpcHealthCheck_builder{
				Service: lo.ToPtr(""), // Empty service name for default health check
			}.Build(),
		}.Build(),
	}.Build()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	wrapper := NewGrpcClientWrapper(conn, config, nil)

	// Set status to SERVING
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	require.Eventually(t, func() bool {
		return wrapper.IsHealthy(context.Background())
	}, 5*time.Second, 100*time.Millisecond)

	// Set status to NOT_SERVING
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)

	require.Eventually(t, func() bool {
		return !wrapper.IsHealthy(context.Background())
	}, 5*time.Second, 100*time.Millisecond)
}
