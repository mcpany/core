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

func TestHandleLogsWS_History(t *testing.T) {
	// Backup and replace GlobalBroadcaster
	originalBroadcaster := logging.GlobalBroadcaster
	logging.GlobalBroadcaster = logging.NewBroadcaster()
	defer func() { logging.GlobalBroadcaster = originalBroadcaster }()

	// Populate history
	historyMsg1 := []byte("history message 1")
	historyMsg2 := []byte("history message 2")
	logging.GlobalBroadcaster.Broadcast(historyMsg1)
	logging.GlobalBroadcaster.Broadcast(historyMsg2)

	app := &Application{}
	handler := app.handleLogsWS()
	ts := httptest.NewServer(handler)
	defer ts.Close()

	u := "ws" + strings.TrimPrefix(ts.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	require.NoError(t, err)
	defer ws.Close()

	// Read history messages
	// Note: We might receive them in quick succession
	_, msg, err := ws.ReadMessage()
	require.NoError(t, err)
	assert.Equal(t, historyMsg1, msg)

	_, msg, err = ws.ReadMessage()
	require.NoError(t, err)
	assert.Equal(t, historyMsg2, msg)
}

func TestHandleLogsWS_Streaming(t *testing.T) {
	// Backup and replace GlobalBroadcaster
	originalBroadcaster := logging.GlobalBroadcaster
	logging.GlobalBroadcaster = logging.NewBroadcaster()
	defer func() { logging.GlobalBroadcaster = originalBroadcaster }()

	app := &Application{}
	handler := app.handleLogsWS()
	ts := httptest.NewServer(handler)
	defer ts.Close()

	u := "ws" + strings.TrimPrefix(ts.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	require.NoError(t, err)
	defer ws.Close()

	// Give it a moment to connect and subscribe
	time.Sleep(50 * time.Millisecond)

	// Broadcast new message
	newMsg := []byte("new message")
	logging.GlobalBroadcaster.Broadcast(newMsg)

	// Read message
	_, msg, err := ws.ReadMessage()
	require.NoError(t, err)
	assert.Equal(t, newMsg, msg)
}

func TestHandleLogsWS_Concurrency(t *testing.T) {
	// Backup and replace GlobalBroadcaster
	originalBroadcaster := logging.GlobalBroadcaster
	logging.GlobalBroadcaster = logging.NewBroadcaster()
	defer func() { logging.GlobalBroadcaster = originalBroadcaster }()

	app := &Application{}
	handler := app.handleLogsWS()
	ts := httptest.NewServer(handler)
	defer ts.Close()

	u := "ws" + strings.TrimPrefix(ts.URL, "http")

	clientCount := 5
	clients := make([]*websocket.Conn, clientCount)

	for i := 0; i < clientCount; i++ {
		ws, _, err := websocket.DefaultDialer.Dial(u, nil)
		require.NoError(t, err)
		clients[i] = ws
		defer ws.Close()
	}

	// Give clients time to connect
	time.Sleep(100 * time.Millisecond)

	msgContent := []byte("broadcast to all")
	logging.GlobalBroadcaster.Broadcast(msgContent)

	for i, ws := range clients {
		ws.SetReadDeadline(time.Now().Add(1 * time.Second))
		_, msg, err := ws.ReadMessage()
		require.NoError(t, err, "Client %d failed to read message", i)
		assert.Equal(t, msgContent, msg, "Client %d received wrong message", i)
	}
}

func TestHandleLogsWS_Close(t *testing.T) {
	// Backup and replace GlobalBroadcaster
	originalBroadcaster := logging.GlobalBroadcaster
	logging.GlobalBroadcaster = logging.NewBroadcaster()
	defer func() { logging.GlobalBroadcaster = originalBroadcaster }()

	app := &Application{}
	handler := app.handleLogsWS()
	ts := httptest.NewServer(handler)
	defer ts.Close()

	u := "ws" + strings.TrimPrefix(ts.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	require.NoError(t, err)

	// Close immediately
	ws.Close()

	// Give server time to detect close (though broadcast is async/channel based)
	time.Sleep(50 * time.Millisecond)

	// Broadcast shouldn't panic and should handle the closed channel internally (via Unsubscribe)
	logging.GlobalBroadcaster.Broadcast([]byte("test"))
}
