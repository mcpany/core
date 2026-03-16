// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
)

// WebsocketClientWrapper wraps a *websocket.Conn to adapt it for use in a
//
// Summary: WebsocketClientWrapper wraps a *websocket.Conn to adapt it for use in a
// Summary: WebsocketClientWrapper wraps a *websocket.Conn to adapt it for use in a
type WebsocketClientWrapper struct {
	Conn *websocket.Conn
}
// IsHealthy checks if the underlying WebSocket connection is still active. It sends a ping message with a short deadline to verify the connection's liveness.
//
// Summary: IsHealthy checks if the underlying WebSocket connection is still active. It sends a ping message with a short deadline to verify the connection's liveness.
//
// Parameters:
//   - _ (context.Context): The provided _ data.
//
// Returns:
//   - bool: True if successful or valid, false otherwise.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (w *WebsocketClientWrapper) IsHealthy(_ context.Context) bool {
	// Send a ping to check the connection.
	// A short deadline is used to prevent blocking.
	err := w.Conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second*2))
// Close terminates the underlying WebSocket connection. Returns an error if the operation fails.
//
// Summary: Close terminates the underlying WebSocket connection. Returns an error if the operation fails.
//
// Parameters:
//   - None.
//
// Returns:
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (w *WebsocketClientWrapper) Close() error {
	return w.Conn.Close()
}
