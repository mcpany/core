// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mcpany/core/server/pkg/logging"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(_ *http.Request) bool {
		return true // Allow all origins for now, should be configurable
	},
}

// handleLogsWS handles WebSocket connections for log streaming.
func (a *Application) handleLogsWS() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logging.GetLogger().Error("failed to upgrade to websocket", "error", err)
			return
		}
		defer func() {
			if err := conn.Close(); err != nil {
				logging.GetLogger().Error("failed to close websocket connection", "error", err)
			}
		}()

		// Subscribe to logs with history
		logCh, history := logging.GlobalBroadcaster.SubscribeWithHistory()
		defer logging.GlobalBroadcaster.Unsubscribe(logCh)

		// Set write deadline
		if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
			logging.GetLogger().Error("failed to set write deadline", "error", err)
			return
		}
		conn.SetPongHandler(func(string) error {
			return conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		})

		// Send history
		for _, msg := range history {
			if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				logging.GetLogger().Error("failed to set write deadline", "error", err)
				return
			}
			// ⚡ BOLT: Write struct directly to WebSocket (marshals internally)
			if err := conn.WriteJSON(msg); err != nil {
				logging.GetLogger().Error("failed to write history log message to websocket", "error", err)
				return
			}
		}

		// ⚡ BOLT: Prevent concurrent read/write to the websocket
		// A websocket connection cannot be read and written to concurrently by multiple goroutines.
		// The read loop is needed to process incoming messages like pong responses, which
		// prevents the connection from timing out from the client's perspective and from our end.

		// Setup a channel to act as a close signal
		closeCh := make(chan struct{})

		// Start a read loop to process pongs and detect client disconnects
		go func() {
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					close(closeCh)
					return
				}
			}
		}()

		// Create a separate ping ticker
		pingTicker := time.NewTicker(5 * time.Second)
		defer pingTicker.Stop()

		for {
			select {
			case <-closeCh:
				return // Client disconnected
			case <-pingTicker.C:
				if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second)); err != nil {
					return
				}
			case msg, ok := <-logCh:
				if !ok {
					return
				}
				// ⚡ BOLT: Write struct directly to WebSocket (marshals internally)
				if err := conn.WriteJSON(msg); err != nil {
					logging.GetLogger().Error("failed to write log message to websocket", "error", err)
					return
				}
				if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
					logging.GetLogger().Error("failed to set write deadline", "error", err)
					return
				}
			}
		}
	}
}
