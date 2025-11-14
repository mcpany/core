/*
 * Copyright 2025 Author(s) of MCP Any
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

package app

import (
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHealthCheck_ConnectionTimeout is designed to explicitly test the timeout
// behavior of the http.Client used in the HealthCheck function. The test sets
// up a TCP listener that accepts a connection but never reads from it,
// simulating a hanging server or network issue.
//
// Without a timeout configured on the http.Client's Transport, the client
// would wait indefinitely for the server's response, causing the health check
// to hang. This test ensures that the client has a timeout that prevents such
// hangs.
func TestHealthCheck_ConnectionTimeout(t *testing.T) {
	// 1. Set up a listener that will accept one connection and then hang.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err, "Failed to create a listener for the test.")
	defer listener.Close()

	// 2. In a separate goroutine, accept one connection and then do nothing.
	// This simulates a server that is unresponsive after connection.
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			// Keep the connection open but don't read or write to it.
			// The client will be stuck in the "waiting for response" state.
			time.Sleep(20 * time.Second) // Keep it open longer than the check timeout.
			conn.Close()
		}
	}()

	// 3. Get the address of our hanging listener.
	addr := listener.Addr().String()

	// 4. Perform the health check against the hanging server.
	// We expect this to time out.
	err = HealthCheck(io.Discard, addr, 15*time.Second)

	// 5. Assert that the health check failed with a timeout error.
	// The error message should indicate a timeout.
	assert.Error(t, err, "HealthCheck should have returned an error due to the timeout.")
	if err != nil {
		assert.Contains(t, err.Error(), "context deadline exceeded", "The error should be a context deadline exceeded error.")
	}
}
