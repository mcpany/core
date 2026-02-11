// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
)

// Summary: Wraps a *websocket.Conn to adapt it for use in a.
type WebsocketClientWrapper struct {
	Conn *websocket.Conn
}

// Summary: Checks if the underlying WebSocket connection is still active. It.
func (w *WebsocketClientWrapper) IsHealthy(_ context.Context) bool {
	// Send a ping to check the connection.
	// A short deadline is used to prevent blocking.
	err := w.Conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second*2))
	return err == nil
}

// Close terminates the underlying WebSocket connection.
//
// Returns an error if the operation fails.
// Summary: Terminates the underlying WebSocket connection.
func (w *WebsocketClientWrapper) Close() error {
	return w.Conn.Close()
}
