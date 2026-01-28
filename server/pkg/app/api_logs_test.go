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
	ws.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, msg, err := ws.ReadMessage()
	require.NoError(t, err)
	assert.Equal(t, historyMsg1, msg)

	ws.SetReadDeadline(time.Now().Add(5 * time.Second))
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

	// Wait for connection to be established by reading potential history (which should be empty)
	// OR sending a sync message.
	// Since we know the server implementation, it sends history immediately upon connection.
	// If history is empty, it just subscribes.
	// To reliably test streaming, we can use a "sync" message mechanism if we could inject it,
	// but here we are testing the endpoint.
	// Instead of sleep, we can retry broadcast until received or timeout.
	// OR we can assume that if we dial successfully, we are close to ready.
	// But `handleLogsWS` subscribes AFTER upgrade.
	// We can't easily hook into "subscribed" event.
	// Best approach: Broadcast a "PING" message repeatedly until client sees it, then broadcast real message.

	syncMsg := []byte("SYNC")
	newMsg := []byte("new message")

	// Helper to wait for sync
	ready := make(chan bool)
	go func() {
		for {
			ws.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
			_, msg, err := ws.ReadMessage()
			if err == nil && string(msg) == string(syncMsg) {
				ready <- true
				return
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
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

Loop:
	for {
		select {
		case <-ready:
			break Loop
		case <-timeout:
			t.Fatal("Timeout waiting for sync message")
		case <-ticker.C:
			logging.GlobalBroadcaster.Broadcast(syncMsg)
		}
	}

	// Now we are synced
	logging.GlobalBroadcaster.Broadcast(newMsg)

	// Read message
	ws.SetReadDeadline(time.Now().Add(5 * time.Second))
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

	// Use Sync pattern for all clients
	syncMsg := []byte("SYNC")
	msgContent := []byte("broadcast to all")

	// Wait for all to be ready
	readyCh := make(chan int, clientCount)
	for i, ws := range clients {
		go func(idx int, c *websocket.Conn) {
			for {
				c.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
				_, msg, err := c.ReadMessage()
				if err == nil && string(msg) == string(syncMsg) {
					readyCh <- idx
					return
				}
				if err != nil && !strings.Contains(err.Error(), "timeout") {
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
	// We just want to ensure no panic happens on next broadcast
	// Retry broadcast a few times to ensure logic paths are hit
	for i := 0; i < 5; i++ {
		logging.GlobalBroadcaster.Broadcast([]byte("test"))
		time.Sleep(10 * time.Millisecond)
	}
}
