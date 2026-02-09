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

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

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
				// Ignore close errors as connection might already be closed
				logging.GetLogger().Debug("websocket connection closed", "error", err)
			}
		}()

		// Subscribe to logs with history
		logCh, history := logging.GlobalBroadcaster.SubscribeWithHistory()
		defer logging.GlobalBroadcaster.Unsubscribe(logCh)

		// ⚡ BOLT: Start a read loop to process Pongs and Close frames.
		// Without this, Pongs pile up in the buffer and the connection may become unhealthy.
		// Randomized Selection from Top 5 High-Impact Targets
		conn.SetReadLimit(512) // Limit incoming control/text frame size
		if err := conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			return
		}
		conn.SetPongHandler(func(string) error {
			return conn.SetReadDeadline(time.Now().Add(pongWait))
		})

		go func() {
			for {
				if _, _, err := conn.NextReader(); err != nil {
					conn.Close()
					break
				}
			}
		}()

		// Set initial write deadline
		if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
			logging.GetLogger().Error("failed to set write deadline", "error", err)
			return
		}

		// Send history
		for _, msg := range history {
			if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				logging.GetLogger().Error("failed to set write deadline", "error", err)
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				logging.GetLogger().Error("failed to write history log message to websocket", "error", err)
				return
			}
		}

		// Send ping periodically
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()

		// Loop for writing messages and pings
		for {
			select {
			case msg, ok := <-logCh:
				if !ok {
					// Channel closed (application shutdown?)
					conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}

				if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
					return
				}
				if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					logging.GetLogger().Error("failed to write log message to websocket", "error", err)
					return
				}

			case <-ticker.C:
				if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
					return
				}
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}
	}
}
