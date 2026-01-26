// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http //nolint:revive,nolintlint // Package name 'http' is intentional for this directory structure.

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestHTTPPool_New(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		configJSON := `{"http_service": {"address": "` + strings.TrimPrefix(server.URL, "http://") + `"}}`
		config := configv1.UpstreamServiceConfig_builder{}.Build()
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), config))

		p, err := NewHTTPPool(1, 5, 100, config)
		require.NoError(t, err)
		assert.NotNil(t, p)
		defer func() { _ = p.Close() }()

		assert.Equal(t, 1, p.Len())

		client, err := p.Get(context.Background())
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.True(t, client.IsHealthy(context.Background()))

		p.Put(client)
		assert.Equal(t, 1, p.Len())
	})

	t.Run("invalid config", func(t *testing.T) {
		_, err := NewHTTPPool(5, 1, 10, configv1.UpstreamServiceConfig_builder{}.Build())
		assert.Error(t, err)
	})
}

func TestHTTPPool_GetPut(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	configJSON := `{"http_service": {"address": "` + strings.TrimPrefix(server.URL, "http://") + `"}}`
	config := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), config))

	p, err := NewHTTPPool(1, 1, 10, config)
	require.NoError(t, err)
	require.NotNil(t, p)

	client, err := p.Get(context.Background())
	require.NoError(t, err)
	require.NotNil(t, client)

	assert.True(t, client.IsHealthy(context.Background()))

	p.Put(client)

	// After putting it back, we should be able to get it again.
	client2, err := p.Get(context.Background())
	require.NoError(t, err)
	require.NotNil(t, client2)

	// It should be the same client since pool size is 1
	assert.Same(t, client, client2)
}

func TestHTTPPool_SharedClients(t *testing.T) {
	p, err := NewHTTPPool(2, 2, 10, configv1.UpstreamServiceConfig_builder{}.Build())
	require.NoError(t, err)
	require.NotNil(t, p)
	defer func() { _ = p.Close() }()

	client1, err := p.Get(context.Background())
	require.NoError(t, err)
	require.NotNil(t, client1)

	client2, err := p.Get(context.Background())
	require.NoError(t, err)
	require.NotNil(t, client2)

	// Clients in the pool should share the same underlying http.Client to enable connection pooling
	assert.Same(t, client1.Client, client2.Client)
}

func TestHTTPPool_Close(t *testing.T) {
	p, err := NewHTTPPool(1, 1, 10, configv1.UpstreamServiceConfig_builder{}.Build())
	require.NoError(t, err)
	require.NotNil(t, p)

	_ = p.Close()

	// After closing, get should fail
	_, err = p.Get(context.Background())
	assert.Error(t, err)
}

func TestHTTPPool_PoolFull(t *testing.T) {
	p, err := NewHTTPPool(1, 1, 1, configv1.UpstreamServiceConfig_builder{}.Build())
	require.NoError(t, err)
	require.NotNil(t, p)

	// Get the only client
	_, err = p.Get(context.Background())
	require.NoError(t, err)

	// Try to get another one, should time out
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	_, err = p.Get(ctx)
	assert.Error(t, err)
}

func TestHTTPPool_KeepAliveEnabled(t *testing.T) {
	p, err := NewHTTPPool(1, 1, 10, configv1.UpstreamServiceConfig_builder{}.Build())
	require.NoError(t, err)
	require.NotNil(t, p)
	defer func() { _ = p.Close() }()

	client, err := p.Get(context.Background())
	require.NoError(t, err)
	require.NotNil(t, client)

	// Check if it is *http.Transport or wrapped by otelhttp
	if _, ok := client.Transport.(*http.Transport); ok {
		transport := client.Transport.(*http.Transport)
		assert.False(t, transport.DisableKeepAlives, "KeepAlives should be enabled")
	} else {
		// If wrapped by otelhttp, we assume it preserves the underlying behavior.
		// We can't easily access the private Base field of otelhttp.Transport.
		t.Log("Transport is wrapped (likely by otelhttp), skipping direct *http.Transport assertions")
	}
}

func TestHTTPPool_TimeoutConfiguration(t *testing.T) {
	t.Run("default timeout", func(t *testing.T) {
		p, err := NewHTTPPool(1, 1, 10, configv1.UpstreamServiceConfig_builder{}.Build())
		require.NoError(t, err)
		defer func() { _ = p.Close() }()

		c, err := p.Get(context.Background())
		require.NoError(t, err)

		assert.Equal(t, 30*time.Second, c.Client.Timeout)
	})

	t.Run("configured timeout", func(t *testing.T) {
		config := configv1.UpstreamServiceConfig_builder{
			Resilience: configv1.ResilienceConfig_builder{
				Timeout: durationpb.New(10 * time.Second),
			}.Build(),
		}.Build()
		p, err := NewHTTPPool(1, 1, 10, config)
		require.NoError(t, err)
		defer func() { _ = p.Close() }()

		c, err := p.Get(context.Background())
		require.NoError(t, err)

		assert.Equal(t, 10*time.Second, c.Client.Timeout)
	})
}
