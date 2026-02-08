// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package websocket

import (
	"context"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/pool"
)

// Pool is a type alias for a pool of WebSocket client connections.
//
// Summary: is a type alias for a pool of WebSocket client connections.
type Pool = pool.Pool[*client.WebsocketClientWrapper]

// NewPool creates a new connection pool for WebSocket clients. It.
//
// Summary: creates a new connection pool for WebSocket clients. It.
//
// Parameters:
//   - maxSize: int. The maxSize.
//   - idleTimeout: time.Duration. The idleTimeout.
//   - address: string. The address.
//
// Returns:
//   - Pool: The Pool.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func NewPool(maxSize int, idleTimeout time.Duration, address string) (Pool, error) {
	factory := func(_ context.Context) (*client.WebsocketClientWrapper, error) {
		conn, resp, err := websocket.DefaultDialer.Dial(address, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to websocket server %s: %w", address, err)
		}
		defer func() { _ = resp.Body.Close() }()
		return &client.WebsocketClientWrapper{Conn: conn}, nil
	}

	// The generic pool expects idleTimeout as an int (seconds).
	// We'll use a minSize of 0 for this pool.
	p, err := pool.New(factory, 0, 0, maxSize, idleTimeout, false)
	if err != nil {
		return nil, err
	}

	return p, nil
}
