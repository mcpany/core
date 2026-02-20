// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
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
	// ⚡ BOLT: Updated to use LogEntry struct or string
	historyMsg1 := "history message 1"
	historyMsg2 := "history message 2"
	logging.GlobalBroadcaster.Broadcast(historyMsg1)
	logging.GlobalBroadcaster.Broadcast(historyMsg2)

	// Mock app structure if needed, or just test handler function
	// handleLogsWS is a method on Application, so we need an instance
	app := &Application{}
	handler := app.handleLogsWS()
	ts := httptest.NewServer(handler)
	defer ts.Close()

	u := "ws" + strings.TrimPrefix(ts.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	require.NoError(t, err)
	defer ws.Close()

	// Read history messages
	// We expect JSON strings now

	ws.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, msgBytes, err := ws.ReadMessage()
	require.NoError(t, err)
	var received1 string
	err = json.Unmarshal(msgBytes, &received1)
	require.NoError(t, err)
	assert.Equal(t, historyMsg1, received1)

	ws.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, msgBytes, err = ws.ReadMessage()
	require.NoError(t, err)
	var received2 string
	err = json.Unmarshal(msgBytes, &received2)
	require.NoError(t, err)
	assert.Equal(t, historyMsg2, received2)
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

	// Wait for connection to be established
	// We use the sync pattern
	syncMsg := "SYNC"
	newMsg := "new message"

	// Helper to wait for sync
	ready := make(chan bool)
	go func() {
		for {
			ws.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
			_, msgBytes, err := ws.ReadMessage()
			if err == nil {
				var received string
				if json.Unmarshal(msgBytes, &received) == nil && received == syncMsg {
					ready <- true
					return
				}
			}
			if err != nil {
				// if timeout, continue. if closed, return
				if !strings.Contains(err.Error(), "timeout") {
					return
				}
			}
		}
	}()

	// Broadcast sync until received
	// We need to run broadcast in loop
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				logging.GlobalBroadcaster.Broadcast(syncMsg)
			}
		}
	}()

	select {
	case <-ready:
		close(done)
	case <-time.After(10 * time.Second):
		close(done)
		t.Fatal("Timeout waiting for sync message")
	}

	// Now we are synced
	logging.GlobalBroadcaster.Broadcast(newMsg)

	// Read message
	ws.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, msgBytes, err := ws.ReadMessage()
	require.NoError(t, err)

	var receivedNew string
	err = json.Unmarshal(msgBytes, &receivedNew)
	require.NoError(t, err)
	assert.Equal(t, newMsg, receivedNew)
}
