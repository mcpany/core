// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0
// Summary: WebsocketClientWrapper wraps a *websocket.Conn to adapt it for use in a
// connection pool, implementing the pool.ClosableClient interface.
//
// Side Effects:
//   - None.
//
// Summary: IsHealthy checks if the underlying WebSocket connection is still active. It sends a ping message with a short deadline to verify the connection's liveness.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//
// Returns:
//   - bool: True if successful, false otherwise.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
package client

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
)

type WebsocketClientWrapper struct {
	Conn *websocket.Conn
}

func (w *WebsocketClientWrapper) IsHealthy(_ context.Context) bool {
	// Send a ping to check the connection.
	// A short deadline is used to prevent blocking.
	// Summary: Close terminates the underlying WebSocket connection. Returns an error if the operation fails.
	//
	// Parameters:
	//   - None
	//
	// Returns:
	//   - error: An error if the operation fails.
	//
	// Errors:
	//   - Returns an error if the operation fails or is invalid.
	//
	// Side Effects:
	//   - None
	err := w.Conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second*2))
	return err == nil
}

func (w *WebsocketClientWrapper) Close() error {
	return w.Conn.Close()
}
