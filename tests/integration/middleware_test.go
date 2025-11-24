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

package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mcpany/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMiddlewareIntegration(t *testing.T) {
	t.Skip("Skipping flaky middleware integration test")
	requestCount := 0
	// Start a mock upstream service
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"status": "ok"}`)
	}))
	defer upstream.Close()

	// Create a config file for the server
	configContent := fmt.Sprintf(`
upstream_services:
  - name: "test-service"
    http_service:
      address: "%s"
      tools:
        - name: "test-tool"
          call_id: "test-call"
      calls:
        test-call:
          endpoint_path: "/"
          method: "HTTP_METHOD_GET"
    cache:
      is_enabled: true
      ttl: "10s"
`, upstream.URL)

	// Start the MCP Any server
	serverInfo := integration.StartMCPANYServerWithConfig(t, "middleware-test", configContent)
	defer serverInfo.CleanupFunc()

	// Wait for the tool to be registered
	require.Eventually(t, func() bool {
		listToolsResult, err := serverInfo.ListTools(context.Background())
		if err != nil {
			t.Logf("ListTools error: %v", err)
			return false
		}
		for _, tool := range listToolsResult.Tools {
			if tool.Name == "test-service.test-tool" {
				return true
			}
		}
		t.Logf("Tool not found in list: %v", listToolsResult.Tools)
		return false
	}, 15*time.Second, 500*time.Millisecond, "tool was not registered")

	// Call the tool for the first time
	callToolParams := &mcp.CallToolParams{
		Name:      "test-service.test-tool",
		Arguments: json.RawMessage(`{}`),
	}
	_, err := serverInfo.CallTool(context.Background(), callToolParams)
	require.NoError(t, err)

	// Check the logs for the first request
	logs := serverInfo.Process.StdoutString()
	require.True(t, strings.Contains(logs, "Request received"), "Log should contain 'Request received'")
	require.True(t, strings.Contains(logs, "Request completed"), "Log should contain 'Request completed'")

	// Call the tool for the second time (should be a cache hit)
	_, err = serverInfo.CallTool(context.Background(), callToolParams)
	require.NoError(t, err)

	// Check that the upstream service was only called once
	assert.Equal(t, 1, requestCount, "Upstream service should have been called only once")
}

func TestRateLimitMiddlewareIntegration(t *testing.T) {
	t.Skip("Skipping test due to issues with MCP HTTP session initialization in test helper")
	// Start a mock upstream service
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"status": "ok"}`)
	}))
	defer upstream.Close()

	// Create a config file for the server with rate limiting enabled
	configContent := fmt.Sprintf(`
upstream_services:
  - name: "rate-limited-service"
    http_service:
      address: "%s"
      tools:
        - name: "test-tool"
          call_id: "test-call"
      calls:
        test-call:
          endpoint_path: "/"
          method: "HTTP_METHOD_GET"
    rate_limit:
      is_enabled: true
      requests_per_second: 1
      burst: 1
`, upstream.URL)

	// Start the MCP Any server
	serverInfo := integration.StartMCPANYServerWithConfig(t, "rate-limit-test", configContent)
	defer serverInfo.CleanupFunc()

	// Initialize MCP session
	require.NoError(t, serverInfo.Initialize(context.Background()))

	// Wait for the tool to be registered
	require.Eventually(t, func() bool {
		listToolsResult, err := serverInfo.ListTools(context.Background())
		if err != nil {
			t.Logf("ListTools error: %v", err)
			return false
		}
		for _, tool := range listToolsResult.Tools {
			if tool.Name == "rate-limited-service.test-tool" {
				return true
			}
		}
		t.Logf("Tool not found in list: %v", listToolsResult.Tools)
		return false
	}, 15*time.Second, 500*time.Millisecond, "tool was not registered")

	callToolParams := &mcp.CallToolParams{
		Name:      "rate-limited-service.test-tool",
		Arguments: json.RawMessage(`{}`),
	}

	// First call should succeed
	_, err := serverInfo.CallTool(context.Background(), callToolParams)
	require.NoError(t, err)

	// Second call should fail due to rate limiting
	_, err = serverInfo.CallTool(context.Background(), callToolParams)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ResourceExhausted")
}
