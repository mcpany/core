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
// Summary: Represents a WebsocketClientWrapper.
type WebsocketClientWrapper struct {
	Conn *websocket.Conn
}

// IsHealthy isHealthy is healthy.
//
// Summary: IsHealthy is healthy.
//
// Parameters: - None.
//   - _ (context.Context): Unused parameter.
//
// Returns: - None.
//   - bool: The result.
func (w *WebsocketClientWrapper) IsHealthy(_ context.Context) bool {
	// Send a ping to check the connection.
	// A short deadline is used to prevent blocking.
	err := w.Conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second*2))
	return err == nil
}

// Close close close.
//
// Summary: Close close.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - error: An error if the operation fails.
func (w *WebsocketClientWrapper) Close() error {
	return w.Conn.Close()
}
