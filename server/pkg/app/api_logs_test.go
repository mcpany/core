// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleLogsWS(t *testing.T) {
	// Setup Application
	app := &Application{}

	handler := app.handleLogsWS()
	ts := httptest.NewServer(handler)
	defer ts.Close()

	// Convert http URL to ws URL
	u := "ws" + strings.TrimPrefix(ts.URL, "http")

	// Connect to WebSocket
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	require.NoError(t, err)
	defer ws.Close()

	// Wait for connection to be established and subscription to happen
	time.Sleep(100 * time.Millisecond)

	// Broadcast a message
	testMsg := []byte("test log message")
	logging.GlobalBroadcaster.Broadcast(testMsg)

	// Read message
	_, msg, err := ws.ReadMessage()
	require.NoError(t, err)
	assert.Equal(t, testMsg, msg)

	// Test Ping/Pong (implicitly handled by goroutine in handler, but we can verify we don't get error)
	// We can try to write a control message
	err = ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second))
	require.NoError(t, err)
}
