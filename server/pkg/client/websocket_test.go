package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebsocketClientWrapper_IsHealthy(t *testing.T) {
	// Start a test server that accepts websocket connections
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Read loop to handle ping/pong and keep connection open
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	t.Run("Healthy Connection", func(t *testing.T) {
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		// We don't close conn here, wrapper will own it (conceptually, though wrapper doesn't own Close logic in original code, it just calls Close on it)
		// Wait, wrapper.Close() calls conn.Close(). So we should let wrapper close it.

		wrapper := &client.WebsocketClientWrapper{Conn: conn}
		assert.True(t, wrapper.IsHealthy(context.Background()))

		err = wrapper.Close()
		assert.NoError(t, err)
	})

	t.Run("Closed Connection", func(t *testing.T) {
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)

		wrapper := &client.WebsocketClientWrapper{Conn: conn}
		err = wrapper.Close()
		require.NoError(t, err)

		assert.False(t, wrapper.IsHealthy(context.Background()))
	})
}
