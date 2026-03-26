// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
)

// WebsocketClientWrapper wraps a *websocket.Conn to adapt it for use in a
// connection pool, implementing the pool.ClosableClient interface.
//
// Summary: Wrapper for WebSocket connections to support pooling.
type WebsocketClientWrapper struct {
	Conn *websocket.Conn
}

// IsHealthy checks if the underlying WebSocket connection is still active.
//
// Summary: Evaluates the connection status by sending a WebSocket ping.
//
// It sends a ping message with a short deadline (2 seconds) to verify
// the connection's liveness.
//
// Parameters:
//   - _: context.Context. Unused.
//
// Returns:
//   - bool: True if the ping was successful, false otherwise.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (w *WebsocketClientWrapper) IsHealthy(_ context.Context) bool {
	// Send a ping to check the connection.
	// A short deadline is used to prevent blocking.
	err := w.Conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second*2))
	return err == nil
}

// Close terminates the underlying WebSocket connection.
//
// Summary: Releases resources associated with the WebSocket connection.
//
// Returns:
//   - error: An error if the closure fails.
//
// Parameters:
//   - None.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (w *WebsocketClientWrapper) Close() error {
	return w.Conn.Close()
}
