// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package websocket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestWSServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		// The connection will be closed by the client or when the server is closed.
		// We can read and discard messages to keep it alive.
		go func() {
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					break
				}
			}
		}()
	}))
}

func TestNewPool(t *testing.T) {
	t.Run("successful pool creation", func(t *testing.T) {
		server := newTestWSServer()
		defer server.Close()
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		pool, err := NewPool(5, 10*time.Second, wsURL)
		require.NoError(t, err)
		assert.NotNil(t, pool)
		defer func() { _ = pool.Close() }()

		client, err := pool.Get(context.Background())
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.True(t, client.IsHealthy(context.Background()))

		pool.Put(client)
	})

	t.Run("invalid address", func(t *testing.T) {
		pool, err := NewPool(5, 10*time.Second, "invalid-address")
		require.NoError(t, err)
		assert.NotNil(t, pool)
		defer func() { _ = pool.Close() }()

		_, err = pool.Get(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to connect to websocket server")
	})

	t.Run("connection failure", func(t *testing.T) {
		// Use a port that is not listening
		wsURL := "ws://127.0.0.1:9999"
		pool, err := NewPool(5, 10*time.Second, wsURL)
		require.NoError(t, err)
		assert.NotNil(t, pool)
		defer func() { _ = pool.Close() }()

		_, err = pool.Get(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to connect to websocket server")
	})

	t.Run("invalid config", func(t *testing.T) {
		server := newTestWSServer()
		defer server.Close()
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		_, err := NewPool(0, 10*time.Second, wsURL)
		require.Error(t, err)
	})
}
