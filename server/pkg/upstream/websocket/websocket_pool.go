// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package websocket

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/util"
)

// Pool is a type alias for a pool of WebSocket client connections.
// It simplifies the type signature for WebSocket connection pools.
type Pool = pool.Pool[*client.WebsocketClientWrapper]

// NewPool creates a new connection pool for WebSocket clients. It
// configures the pool with a factory function that establishes new WebSocket
// connections to the specified address.
//
// maxSize is the maximum number of connections the pool can hold.
// idleTimeout is the duration after which an idle connection may be closed.
// address is the target URL of the WebSocket server.
// It returns a new WebSocket client pool or an error if the pool cannot be
// created.
func NewPool(maxSize int, idleTimeout time.Duration, address string) (Pool, error) {
	factory := func(ctx context.Context) (*client.WebsocketClientWrapper, error) {
		safeDialer := util.NewSafeDialer()
		// Allow overriding safety checks via environment variables (consistent with SafeHTTPClient)
		if os.Getenv("MCPANY_ALLOW_LOOPBACK_RESOURCES") == "true" {
			safeDialer.AllowLoopback = true
		}
		if os.Getenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES") == "true" {
			safeDialer.AllowPrivate = true
		}

		dialer := &websocket.Dialer{
			Proxy:            http.ProxyFromEnvironment,
			HandshakeTimeout: 45 * time.Second,
			NetDialContext:   safeDialer.DialContext,
		}
		conn, resp, err := dialer.DialContext(ctx, address, nil)
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
