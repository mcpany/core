// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
)

// WebsocketClientWrapper represents the public WebsocketClientWrapper entity.
//
// Summary: Defines the structured data model representing a client wrapper.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type WebsocketClientWrapper struct {
	Conn *websocket.Conn
}

// IsHealthy serves as a public interface for interacting with IsHealthy.
//
// Summary: Checks condition indicating whether the target is healthy.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (w *WebsocketClientWrapper) IsHealthy(_ context.Context) bool {
	// Send a ping to check the connection.
	// A short deadline is used to prevent blocking.
	err := w.Conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second*2))
	return err == nil
}

// Close serves as a public interface for interacting with Close.
//
// Summary: Close the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (w *WebsocketClientWrapper) Close() error {
	return w.Conn.Close()
}
