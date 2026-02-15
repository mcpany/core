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

	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheMiddleware_CacheHit(t *testing.T) {
	// // t.Skip("Skipping flaky test: tool registration times out intermittently.")
	var requestCount int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"status": "ok"}`)
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

	require.NoError(t, serverInfo.Initialize(context.Background()))

	require.Eventually(t, func() bool {
		listToolsResult, err := serverInfo.ListTools(context.Background())
		if err != nil {
			t.Logf("ListTools failed: %v", err)
			return false
		}
		var names []string
		for _, tool := range listToolsResult.Tools {
			names = append(names, tool.Name)
			if tool.Name == "test-service.test-tool" {
				return true
			}
		}
		t.Logf("Tool not found. Found: %v", names)
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
	// // t.Skip("Skipping flaky test: tool registration times out intermittently.")
	var requestCount int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"status": "ok"}`)
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

	require.NoError(t, serverInfo.Initialize(context.Background()))

	require.Eventually(t, func() bool {
		listToolsResult, err := serverInfo.ListTools(context.Background())
		if err != nil {
			t.Logf("ListTools error: %v", err)
			return false
		}
		var names []string
		for _, tool := range listToolsResult.Tools {
			names = append(names, tool.Name)
			if tool.Name == "test-service.test-tool" {
				return true
			}
		}
		t.Logf("Tool not found. Found: %v", names)
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
