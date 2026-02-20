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

	// Populate history with structs
	entry1 := logging.LogEntry{Message: "history message 1"}
	entry2 := logging.LogEntry{Message: "history message 2"}
	logging.GlobalBroadcaster.Broadcast(entry1)
	logging.GlobalBroadcaster.Broadcast(entry2)

	app := &Application{}
	handler := app.handleLogsWS()
	ts := httptest.NewServer(handler)
	defer ts.Close()

	u := "ws" + strings.TrimPrefix(ts.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	require.NoError(t, err)
	defer ws.Close()

	// Read history messages
	ws.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, msg, err := ws.ReadMessage()
	require.NoError(t, err)

	var received1 logging.LogEntry
	err = json.Unmarshal(msg, &received1)
	require.NoError(t, err)
	assert.Equal(t, entry1.Message, received1.Message)

	ws.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, msg, err = ws.ReadMessage()
	require.NoError(t, err)

	var received2 logging.LogEntry
	err = json.Unmarshal(msg, &received2)
	require.NoError(t, err)
	assert.Equal(t, entry2.Message, received2.Message)
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

	syncMsg := logging.LogEntry{Message: "SYNC"}
	newMsg := logging.LogEntry{Message: "new message"}

	// Wait for sync
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			ws.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
			_, msg, err := ws.ReadMessage()
			if err == nil {
				var received logging.LogEntry
				if json.Unmarshal(msg, &received) == nil && received.Message == "SYNC" {
					return
				}
			}
			if err != nil && !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "temporarily unavailable") {
				return
			}
		}
	}()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	timeout := time.After(5 * time.Second)

Loop:
	for {
		select {
		case <-done:
			break Loop
		case <-timeout:
			t.Fatal("Timeout waiting for sync")
		case <-ticker.C:
			logging.GlobalBroadcaster.Broadcast(syncMsg)
		}
	}

	// Now broadcast real message
	logging.GlobalBroadcaster.Broadcast(newMsg)

	// Read it
	ws.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, msg, err := ws.ReadMessage()
	require.NoError(t, err)

	var received logging.LogEntry
	err = json.Unmarshal(msg, &received)
	require.NoError(t, err)
	assert.Equal(t, newMsg.Message, received.Message)
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

	// Use Sync pattern for all clients
	syncMsg := logging.LogEntry{Message: "SYNC"}
	msgContent := logging.LogEntry{Message: "broadcast to all"}

	// Wait for all to be ready
	readyCh := make(chan int, clientCount)
	for i, ws := range clients {
		go func(idx int, c *websocket.Conn) {
			for {
				c.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
				_, msg, err := c.ReadMessage()
				if err == nil {
					var received logging.LogEntry
					if json.Unmarshal(msg, &received) == nil && received.Message == "SYNC" {
						readyCh <- idx
						return
					}
				}
				if err != nil && !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "temporarily unavailable") {
					return
				}
			}
		}(i, ws)
	}

	// Broadcast sync
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	syncedCount := 0
	// We need to track which clients synced to avoid double counting if they receive multiple SYNCs?
	// Actually, the goroutine returns after one SYNC receipt.
	// But `readyCh` receives one int per client.

SyncLoop:
	for {
		select {
		case <-readyCh:
			syncedCount++
			if syncedCount == clientCount {
				break SyncLoop
			}
		case <-timeout:
			t.Fatal("Timeout waiting for clients to sync")
		case <-ticker.C:
			logging.GlobalBroadcaster.Broadcast(syncMsg)
		}
	}

	// Broadcast real message
	logging.GlobalBroadcaster.Broadcast(msgContent)

	for i, ws := range clients {
		ws.SetReadDeadline(time.Now().Add(5 * time.Second))
		_, msg, err := ws.ReadMessage()
		require.NoError(t, err, "Client %d failed to read message", i)

		var received logging.LogEntry
		err = json.Unmarshal(msg, &received)
		require.NoError(t, err)
		assert.Equal(t, msgContent.Message, received.Message, "Client %d received wrong message", i)
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
	// We just want to ensure no panic happens on next broadcast
	// Retry broadcast a few times to ensure logic paths are hit
	for i := 0; i < 5; i++ {
		logging.GlobalBroadcaster.Broadcast(logging.LogEntry{Message: "test"})
		time.Sleep(10 * time.Millisecond)
	}
}
