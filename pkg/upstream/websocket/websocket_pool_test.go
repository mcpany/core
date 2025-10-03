/*
 * Copyright 2025 Author(s) of MCPX
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

func TestNewWebsocketPool(t *testing.T) {
	t.Run("successful pool creation", func(t *testing.T) {
		server := newTestWSServer()
		defer server.Close()
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		pool, err := NewWebsocketPool(5, 10*time.Second, wsURL)
		require.NoError(t, err)
		assert.NotNil(t, pool)
		defer pool.Close()

		client, err := pool.Get(context.Background())
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.True(t, client.IsHealthy())

		pool.Put(client)
	})

	t.Run("invalid address", func(t *testing.T) {
		pool, err := NewWebsocketPool(5, 10*time.Second, "invalid-address")
		require.NoError(t, err)
		assert.NotNil(t, pool)
		defer pool.Close()

		_, err = pool.Get(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to connect to websocket server")
	})

	t.Run("connection failure", func(t *testing.T) {
		// Use a port that is not listening
		wsURL := "ws://localhost:9999"
		pool, err := NewWebsocketPool(5, 10*time.Second, wsURL)
		require.NoError(t, err)
		assert.NotNil(t, pool)
		defer pool.Close()

		_, err = pool.Get(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to connect to websocket server")
	})

	t.Run("invalid config", func(t *testing.T) {
		server := newTestWSServer()
		defer server.Close()
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		_, err := NewWebsocketPool(0, 10*time.Second, wsURL)
		require.Error(t, err)
	})
}
