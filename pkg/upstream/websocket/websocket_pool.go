/*
 * Copyright 2025 Author(s) of MCP-XY
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package websocket

import (
	"context"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mcpxy/core/pkg/client"
	"github.com/mcpxy/core/pkg/pool"
)

// WebsocketPool is a type alias for a pool of WebSocket client connections.
// It simplifies the type signature for WebSocket connection pools.
type WebsocketPool = pool.Pool[*client.WebsocketClientWrapper]

// NewWebsocketPool creates a new connection pool for WebSocket clients. It
// configures the pool with a factory function that establishes new WebSocket
// connections to the specified address.
//
// maxSize is the maximum number of connections the pool can hold.
// idleTimeout is the duration after which an idle connection may be closed.
// address is the target URL of the WebSocket server.
// It returns a new WebSocket client pool or an error if the pool cannot be
// created.
func NewWebsocketPool(maxSize int, idleTimeout time.Duration, address string) (WebsocketPool, error) {
	factory := func(ctx context.Context) (*client.WebsocketClientWrapper, error) {
		conn, resp, err := websocket.DefaultDialer.Dial(address, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to websocket server %s: %w", address, err)
		}
		defer resp.Body.Close() //nolint:errcheck
		return &client.WebsocketClientWrapper{Conn: conn}, nil
	}

	// The generic pool expects idleTimeout as an int (seconds).
	// We'll use a minSize of 0 for this pool.
	p, err := pool.New(factory, 0, maxSize, int(idleTimeout.Seconds()), false)
	if err != nil {
		return nil, err
	}

	return p, nil
}
