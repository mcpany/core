// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	httpupstream "github.com/mcpany/core/server/pkg/upstream/http"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestHTTPPoolConnectionLeak(t *testing.T) {
	// Start a mock server that counts active connections
	connectionCount := 0
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// We use the server's ConnState hook to track connections
	server.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		switch state {
		case http.StateNew:
			connectionCount++
		case http.StateClosed: // This might not trigger immediately for Keep-Alive connections
			// connectionCount-- // We want to see if they accumulate, so we won't decrement here relies on client closing
		}
	}
	server.Start()
	defer server.Close()

	initialGoroutines := runtime.NumGoroutine()

	// Config for the pool
	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
		HttpService: configv1.HttpUpstreamService_builder{
			Address: proto.String(server.URL),
			TlsConfig: configv1.TLSConfig_builder{
				InsecureSkipVerify: proto.Bool(true),
			}.Build(),
		}.Build(),
		ConnectionPool: configv1.ConnectionPoolConfig_builder{
			MaxConnections:     proto.Int32(10),
			MaxIdleConnections: proto.Int32(10),
			IdleTimeout:        durationpb.New(time.Second),
		}.Build(),
	}.Build()

	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "true")

	for i := 0; i < 20; i++ {
		// Create a new pool
		pool, err := httpupstream.NewHTTPPool(1, 10, time.Second, config)
		assert.NoError(t, err)

		// Get a client and make a request
		ctx := context.Background()
		client, err := pool.Get(ctx)
		assert.NoError(t, err)

		resp, err := client.Get(server.URL)
		if assert.NoError(t, err) {
			resp.Body.Close()
		}

		// Return client to pool
		pool.Put(client)

		// Close the pool
		// This SHOULD close the idle connections
		err = pool.Close()
		assert.NoError(t, err)

		// Wait a bit for cleanup
		time.Sleep(10 * time.Millisecond)
	}

	// Force GC and wait to ensure connections are closed if finalizers are involved (unlikely here)
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	finalGoroutines := runtime.NumGoroutine()
	fmt.Printf("Initial Goroutines: %d, Final Goroutines: %d\n", initialGoroutines, finalGoroutines)

	// If connections are leaking, we expect goroutines (handling connections) or just open sockets.
	// httptest server spawns a goroutine per connection.
	// If the client doesn't close the connection, the server keeps it open (Keep-Alive).
	// So goroutine count should increase significantly if we leak connections.

	// We did 20 iterations. If we leak 1 connection per iteration, we expect +20 goroutines (roughly).
	// Allow some noise.
	assert.LessOrEqual(t, finalGoroutines, initialGoroutines+5, "Goroutine leak detected, likely due to unclosed connections")
}
