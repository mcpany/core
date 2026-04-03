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

// Pool represents the public Pool entity.
//
// Summary: Defines the structured data model representing a .
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
type Pool = pool.Pool[*client.WebsocketClientWrapper]

// NewPool serves as a public interface for interacting with NewPool.
//
// Summary: Constructs and returns an initialized pool ready for consumption.
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
