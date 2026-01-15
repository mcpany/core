// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	httpupstream "github.com/mcpany/core/server/pkg/upstream/http"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestHTTPPool_SSRF_Protection(t *testing.T) {
	// Create a local test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	configJSON := `{"http_service": {"address": "` + strings.TrimPrefix(server.URL, "http://") + `"}}`
	config := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), config))
	// Add config for connection pool to match expectations
	config.ConnectionPool = &configv1.ConnectionPoolConfig{
		MaxConnections:     proto.Int32(10),
		MaxIdleConnections: proto.Int32(10),
		IdleTimeout:        durationpb.New(time.Second),
	}

	t.Run("blocks_loopback_usage_by_default", func(t *testing.T) {
		// Ensure env var is unset/false
		t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "false")

		p, err := httpupstream.NewHTTPPool(1, 1, 10, config)
		require.NoError(t, err)
		defer func() { _ = p.Close() }()

		// Get should succeed because no health check is configured, so IsHealthy returns true
		client, err := p.Get(context.Background())
		require.NoError(t, err)

		// But using the client to connect to loopback should fail
		_, err = client.Get(server.URL)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ssrf attempt blocked")
	})

	t.Run("blocks_loopback_health_check", func(t *testing.T) {
		t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "false")

		// Config with health check
		configWithHC := &configv1.UpstreamServiceConfig{}
		configJSONWithHC := `{"http_service": {"address": "` + strings.TrimPrefix(server.URL, "http://") + `", "health_check": {"url": "` + server.URL + `"}}}`
		require.NoError(t, protojson.Unmarshal([]byte(configJSONWithHC), configWithHC))
		configWithHC.ConnectionPool = config.ConnectionPool

		p, err := httpupstream.NewHTTPPool(1, 1, 10, configWithHC)
		require.NoError(t, err)
		defer func() { _ = p.Close() }()

		// Now Get should fail because IsHealthy will fail (health check blocked)
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		_, err = p.Get(ctx)
		assert.Error(t, err)
	})

	t.Run("allows_loopback_when_configured", func(t *testing.T) {
		t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")

		p, err := httpupstream.NewHTTPPool(1, 1, 10, config)
		require.NoError(t, err)
		defer func() { _ = p.Close() }()

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		client, err := p.Get(ctx)
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.True(t, client.IsHealthy(ctx))
	})
}
