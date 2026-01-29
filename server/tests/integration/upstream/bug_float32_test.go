package upstream

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_Float32Bug(t *testing.T) {
	// 1. Start a mock upstream server
	var receivedPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer ts.Close()

	// 2. Generate config
	upstreamURL := ts.URL
	configContent := fmt.Sprintf(`
upstream_services:
  - name: bug_check_service
    http_service:
      address: %s
      tools:
        - name: get_item
          call_id: get_item_call
      calls:
        get_item_call:
          endpoint_path: /items/{{id}}
          method: HTTP_METHOD_GET
          parameters:
            - schema:
                name: id
          input_schema:
            type: object
            properties:
              id:
                type: number
            required:
              - id
`, upstreamURL)

	// 3. Start MCP Any server
	serverInfo := integration.StartMCPANYServerWithConfig(t, "Float32BugTest", configContent)
	defer serverInfo.CleanupFunc()

	// 4. Connect MCP Client
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: serverInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer func() { _ = cs.Close() }()

	// 5. Call tool with large number
	toolName := "bug_check_service.get_item"

	// Wait for tool to be available
	require.Eventually(t, func() bool {
		listRes, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
		if err != nil {
			return false
		}
		for _, tool := range listRes.Tools {
			if tool.Name == toolName {
				return true
			}
		}
		return false
	}, integration.TestWaitTimeShort, 100*time.Millisecond, "Tool not found")

	// Call with 3 billion
	// 3000000000
	args := json.RawMessage(`{"id": 3000000000}`)
	_, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      toolName,
		Arguments: args,
	})
	require.NoError(t, err)

	// 6. Verify received path
	// Before fix: /items/3e+09
	// After fix: /items/3000000000
	require.Equal(t, "/items/3000000000", receivedPath)
}
