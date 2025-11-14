/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may not obtain a copy of the License at
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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/mcpserver"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_WithCaching(t *testing.T) {
	fs := afero.NewMemMapFs()

	// 1. Set up a mock upstream service
	var callCount int32
	mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.Header().Set("Content-Type", "application/json")
		// This is a simplified successful JSON-RPC response
		response := map[string]interface{}{
			"jsonrpc": "2.0",
			"result":  map[string]interface{}{"output": "mock response"},
			"id":      "1",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer mockUpstream.Close()

	// 2. Create a config file that points to the mock upstream
	configContent := fmt.Sprintf(`
upstream_services:
 - name: "caching-test-service"
   http_service:
     address: "%s"
     tools:
       - name: "test-tool"
         call_id: "test-call"
         cache_config:
           ttl: "10s" # Enable caching with a 10-second TTL
     calls:
        test-call:
          id: "test-call"
          endpoint_path: "/"
          method: "HTTP_METHOD_POST"
`, mockUpstream.URL)
	err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0o644)
	require.NoError(t, err)

	// 3. Run the application
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := NewApplication()
	errChan := make(chan error, 1)
	var serverAddr string
	go func() {
		// Use an ephemeral port for the JSON-RPC server
		l, err := net.Listen("tcp", "localhost:0")
		require.NoError(t, err)
		serverAddr = l.Addr().String()
		l.Close() // Close the listener, app.Run will re-listen
		errChan <- app.Run(ctx, fs, false, serverAddr, "", []string{"/config.yaml"}, 5*time.Second)
	}()

	// Wait for the server to be ready
	require.Eventually(t, func() bool {
		conn, err := net.Dial("tcp", serverAddr)
		if err == nil {
			conn.Close()
			return true
		}
		return false
	}, 3*time.Second, 100*time.Millisecond, "Server did not start in time")

	// 4. Send two identical requests
	httpClient := &http.Client{Timeout: 2 * time.Second}
	makeRequest := func() (*http.Response, error) {
		requestBody := map[string]interface{}{
			"jsonrpc": "2.0",
			"method":  "tools/call",
			"params": map[string]interface{}{
				"name":      "test-tool",
				"arguments": json.RawMessage(`{"input": "test"}`),
			},
			"id": "1",
		}
		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://%s", serverAddr), bytes.NewReader(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		return httpClient.Do(req)
	}

	// First request - should go to the upstream
	resp, err := makeRequest()
	require.NoError(t, err, "First tool call failed")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount), "Upstream should have been called once after the first request")

	// Second request - should be served from cache
	resp, err = makeRequest()
	require.NoError(t, err, "Second tool call failed")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount), "Upstream should not be called a second time; response should be cached")

	// 5. Cleanly shut down the server
	cancel()
	err = <-errChan
	assert.NoError(t, err, "app.Run should return nil on graceful shutdown")
}

func TestRun_GRPCRegistrationServerFailsToCreate(t *testing.T) {
	fs := afero.NewMemMapFs()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Inject an error into the NewRegistrationServer function
	mcpserver.NewRegistrationServerHook = func(bus *bus.BusProvider) (*mcpserver.RegistrationServer, error) {
		return nil, fmt.Errorf("injected registration server error")
	}
	defer func() { mcpserver.NewRegistrationServerHook = nil }()

	app := NewApplication()
	err := app.Run(ctx, fs, false, "localhost:0", "localhost:0", nil, 5*time.Second)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create API server: injected registration server error")
}

func TestRun_MCPServerFailsToCreate(t *testing.T) {
	fs := afero.NewMemMapFs()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Inject an error into the NewServer function
	mcpserver.NewServerHook = func() (*mcpserver.Server, error) {
		return nil, fmt.Errorf("injected mcp server error")
	}
	defer func() { mcpserver.NewServerHook = nil }()

	app := NewApplication()
	err := app.Run(ctx, fs, false, "localhost:0", "localhost:0", nil, 5*time.Second)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create mcp server: injected mcp server error")
}
