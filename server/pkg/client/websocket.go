// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
)

// WebsocketClientWrapper wraps a *websocket.Conn to adapt it for use in a.
//
// Summary: wraps a *websocket.Conn to adapt it for use in a.
type WebsocketClientWrapper struct {
	Conn *websocket.Conn
}

// IsHealthy checks if the underlying WebSocket connection is still active. It.
//
// Summary: checks if the underlying WebSocket connection is still active. It.
//
// Parameters:
//   - _: context.Context. The _.
//
// Returns:
//   - bool: The bool.
func (w *WebsocketClientWrapper) IsHealthy(_ context.Context) bool {
	// Send a ping to check the connection.
	// A short deadline is used to prevent blocking.
	err := w.Conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second*2))
	return err == nil
}

// Close terminates the underlying WebSocket connection.
//
// Summary: terminates the underlying WebSocket connection.
//
// Parameters:
//   None.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (w *WebsocketClientWrapper) Close() error {
	return w.Conn.Close()
}
