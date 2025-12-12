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

package http

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
)

func TestHttpPool_New(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		configJSON := `{"http_service": {"address": "` + strings.TrimPrefix(server.URL, "http://") + `"}}`
		config := &configv1.UpstreamServiceConfig{}
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), config))

		p, err := NewHttpPool(1, 5, 100, config)
		require.NoError(t, err)
		assert.NotNil(t, p)
		defer p.Close()

		assert.Equal(t, 1, p.Len())

		client, err := p.Get(context.Background())
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.True(t, client.IsHealthy(context.Background()))

		p.Put(client)
		assert.Equal(t, 1, p.Len())
	})

	t.Run("invalid config", func(t *testing.T) {
		_, err := NewHttpPool(5, 1, 10, &configv1.UpstreamServiceConfig{})
		assert.Error(t, err)
	})
}

func TestHttpPool_GetPut(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	configJSON := `{"http_service": {"address": "` + strings.TrimPrefix(server.URL, "http://") + `"}}`
	config := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), config))

	p, err := NewHttpPool(1, 1, 10, config)
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

func TestHttpPool_UniqueClients(t *testing.T) {
	p, err := NewHttpPool(2, 2, 10, &configv1.UpstreamServiceConfig{})
	require.NoError(t, err)
	require.NotNil(t, p)
	defer p.Close()

	client1, err := p.Get(context.Background())
	require.NoError(t, err)
	require.NotNil(t, client1)

	client2, err := p.Get(context.Background())
	require.NoError(t, err)
	require.NotNil(t, client2)

	assert.NotSame(t, client1.Client, client2.Client)
}

func TestHttpPool_Close(t *testing.T) {
	p, err := NewHttpPool(1, 1, 10, &configv1.UpstreamServiceConfig{})
	require.NoError(t, err)
	require.NotNil(t, p)

	p.Close()

	// After closing, get should fail
	_, err = p.Get(context.Background())
	assert.Error(t, err)
}

func TestHttpPool_PoolFull(t *testing.T) {
	p, err := NewHttpPool(1, 1, 1, &configv1.UpstreamServiceConfig{})
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

func TestHttpPool_KeepAliveEnabled(t *testing.T) {
	p, err := NewHttpPool(1, 1, 10, &configv1.UpstreamServiceConfig{})
	require.NoError(t, err)
	require.NotNil(t, p)
	defer p.Close()

	client, err := p.Get(context.Background())
	require.NoError(t, err)
	require.NotNil(t, client)

	transport, ok := client.Transport.(*http.Transport)
	require.True(t, ok, "Transport is not an *http.Transport")

	assert.False(t, transport.DisableKeepAlives, "KeepAlives should be enabled")
}
