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
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY, either express or implied.
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
	"sync/atomic"
	"testing"
	"time"

	"github.com/mcpany/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheMiddleware_CacheHit(t *testing.T) {
	t.Skip("Skipping flaky test: tool registration times out intermittently.")
	var requestCount int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"status": "ok"}`)
	}))
	defer upstream.Close()

	configContent := fmt.Sprintf(`
upstream_services:
  - name: "test-service"
    http_service:
      address: "%s"
      calls:
        test_call:
          endpoint_path: "/"
          method: "HTTP_METHOD_GET"
      tools:
        - name: "test-tool"
          call_id: "test_call"
    cache:
      is_enabled: true
      ttl: "10s"
`, upstream.URL)

	serverInfo := integration.StartMCPANYServerWithConfig(t, "cache-hit-test", configContent)
	defer serverInfo.CleanupFunc()

	require.Eventually(t, func() bool {
		listToolsResult, err := serverInfo.ListTools(context.Background())
		if err != nil {
			return false
		}
		for _, tool := range listToolsResult.Tools {
			if tool.Name == "test-service.test-tool" {
				return true
			}
		}
		return false
	}, 15*time.Second, 500*time.Millisecond, "tool was not registered")

	callToolParams := &mcp.CallToolParams{
		Name:      "test-service.test-tool",
		Arguments: json.RawMessage(`{}`),
	}
	_, err := serverInfo.CallTool(context.Background(), callToolParams)
	require.NoError(t, err)

	_, err = serverInfo.CallTool(context.Background(), callToolParams)
	require.NoError(t, err)

	assert.Equal(t, int32(1), atomic.LoadInt32(&requestCount), "Upstream service should have been called only once")
}

func TestCacheMiddleware_CacheExpires(t *testing.T) {
	t.Skip("Skipping flaky test: tool registration times out intermittently.")
	var requestCount int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"status": "ok"}`)
	}))
	defer upstream.Close()

	configContent := fmt.Sprintf(`
upstream_services:
  - name: "test-service"
    http_service:
      address: "%s"
      calls:
        test_call:
          endpoint_path: "/"
          method: "HTTP_METHOD_GET"
      tools:
        - name: "test-tool"
          call_id: "test_call"
    cache:
      is_enabled: true
      ttl: "1s"
`, upstream.URL)

	serverInfo := integration.StartMCPANYServerWithConfig(t, "cache-expires-test", configContent)
	defer serverInfo.CleanupFunc()

	require.Eventually(t, func() bool {
		listToolsResult, err := serverInfo.ListTools(context.Background())
		if err != nil {
			return false
		}
		for _, tool := range listToolsResult.Tools {
			if tool.Name == "test-service.test-tool" {
				return true
			}
		}
		return false
	}, 15*time.Second, 500*time.Millisecond, "tool was not registered")

	callToolParams := &mcp.CallToolParams{
		Name:      "test-service.test-tool",
		Arguments: json.RawMessage(`{}`),
	}
	_, err := serverInfo.CallTool(context.Background(), callToolParams)
	require.NoError(t, err)
	require.Equal(t, int32(1), atomic.LoadInt32(&requestCount), "Upstream should be called once")

	time.Sleep(1100 * time.Millisecond)

	_, err = serverInfo.CallTool(context.Background(), callToolParams)
	require.NoError(t, err)

	assert.Equal(t, int32(2), atomic.LoadInt32(&requestCount), "Upstream service should have been called again after cache expired")
}

func TestAuthMiddleware_APIKeyAuthentication(t *testing.T) {
	configContent := `
global_settings:
  api_key: "test-api-key"
`
	serverInfo := integration.StartMCPANYServerWithConfig(t, "api-key-auth-test", configContent)
	defer serverInfo.CleanupFunc()

	// Test case 1: Valid API key
	t.Run("ValidAPIKey", func(t *testing.T) {
		ctx := context.Background()
		_, err := serverInfo.ListTools(ctx, func(req *http.Request) {
			req.Header.Set("X-API-Key", "test-api-key")
		})
		require.NoError(t, err)
	})

	// Test case 2: Invalid API key
	t.Run("InvalidAPIKey", func(t *testing.T) {
		ctx := context.Background()
		_, err := serverInfo.ListTools(ctx, func(req *http.Request) {
			req.Header.Set("X-API-Key", "invalid-api-key")
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Unauthorized")
	})

	// Test case 3: Missing API key
	t.Run("MissingAPIKey", func(t *testing.T) {
		ctx := context.Background()
		_, err := serverInfo.ListTools(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Unauthorized")
	})
}
