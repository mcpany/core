// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package integration_test

import (
	"context"

	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheMiddleware_CacheHit(t *testing.T) {
	var requestCount int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"status": "ok"}`)
	}))
	defer upstream.Close()

	// Use integration.TestWaitTimeLong instead of Short to prevent intermittent timeouts
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
	defer cancel()

	configTemplate := `
global_settings:
  mcp_listen_address: "127.0.0.1:0"
upstream_services:
  - name: cache-hit-service
    http_service:
      address: "%s"
      tools:
        - name: hit_tool
          call_id: hit_call
      calls:
        hit_call:
          endpoint_path: /hit
          method: HTTP_METHOD_GET
      middlewares:
        - type: cache
          cache:
            ttl_seconds: 60
`
	config := fmt.Sprintf(configTemplate, upstream.URL)
	serverInfo := integration.StartMCPANYServerWithConfig(t, "TestCacheMiddleware_CacheHit", config)
	defer serverInfo.CleanupFunc()

	client, session, cleanup := integration.ConnectToMCPServer(t, ctx, serverInfo.MCPAddress, serverInfo.APIKey)
	defer cleanup()
	defer session.Close()

	// Wait for tools to be registered
	require.Eventually(t, func() bool {
		res, err := client.ListTools(ctx, mcp.ListToolsRequest{})
		if err != nil {
			return false
		}
		for _, tool := range res.Tools {
			if tool.Name == "cache-hit-service.hit_tool" {
				return true
			}
		}
		return false
	}, 10*time.Second, 1*time.Second, "Tool was not registered in time")

	callTool := func() string {
		res, err := client.CallTool(ctx, mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
			}{
				Name:      "cache-hit-service.hit_tool",
				Arguments: map[string]any{},
			},
		})
		require.NoError(t, err)
		require.False(t, res.IsError)
		return res.Content[0].(mcp.TextContent).Text
	}

	// First call
	res1 := callTool()
	assert.Contains(t, res1, `"status": "ok"`)

	// Second call
	res2 := callTool()
	assert.Contains(t, res2, `"status": "ok"`)

	assert.Equal(t, int32(1), atomic.LoadInt32(&requestCount), "Upstream service should have been called only once")
}

func TestCacheMiddleware_CacheExpires(t *testing.T) {
	var requestCount int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"status": "ok"}`)
	}))
	defer upstream.Close()

	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
	defer cancel()

	configTemplate := `
global_settings:
  mcp_listen_address: "127.0.0.1:0"
upstream_services:
  - name: cache-expire-service
    http_service:
      address: "%s"
      tools:
        - name: expire_tool
          call_id: expire_call
      calls:
        expire_call:
          endpoint_path: /expire
          method: HTTP_METHOD_GET
      middlewares:
        - type: cache
          cache:
            ttl_seconds: 1
`
	config := fmt.Sprintf(configTemplate, upstream.URL)
	serverInfo := integration.StartMCPANYServerWithConfig(t, "TestCacheMiddleware_CacheExpires", config)
	defer serverInfo.CleanupFunc()

	client, session, cleanup := integration.ConnectToMCPServer(t, ctx, serverInfo.MCPAddress, serverInfo.APIKey)
	defer cleanup()
	defer session.Close()

	require.Eventually(t, func() bool {
		res, err := client.ListTools(ctx, mcp.ListToolsRequest{})
		if err != nil {
			return false
		}
		for _, tool := range res.Tools {
			if tool.Name == "cache-expire-service.expire_tool" {
				return true
			}
		}
		return false
	}, 10*time.Second, 1*time.Second, "Tool was not registered in time")

	callTool := func() {
		_, err := client.CallTool(ctx, mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
			}{
				Name:      "cache-expire-service.expire_tool",
				Arguments: map[string]any{},
			},
		})
		require.NoError(t, err)
	}

	callTool()
	callTool() // should hit cache

	assert.Equal(t, int32(1), atomic.LoadInt32(&requestCount))

	time.Sleep(1500 * time.Millisecond)

	callTool() // should miss cache

	assert.Equal(t, int32(2), atomic.LoadInt32(&requestCount))
}
