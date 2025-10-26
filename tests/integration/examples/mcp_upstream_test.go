//go:build e2e

package examples

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mcpxy/core/pkg/util"
	"github.com/mcpxy/core/tests/framework"
	"github.com/mcpxy/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestMCPUpstreamExample_Stdio(t *testing.T) {
	testCase := &framework.E2ETestCase{
		Name:                "MCP Upstream Stdio Example",
		UpstreamServiceType: "mcp-stdio",
		GenerateUpstreamConfig: func(upstreamEndpoint string) string {
			root, err := integration.GetProjectRoot()
			require.NoError(t, err)
			return filepath.Join(root, "examples/upstream/mcp/stdio/mcpxy_config.yaml")
		},
		InvokeAIClient: func(t *testing.T, mcpxyEndpoint string) {
			ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
			defer cancel()

			testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
			cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpxyEndpoint}, nil)
			require.NoError(t, err, "Failed to connect to MCPXY server")
			defer cs.Close()

			serviceKey, err := util.GenerateID("mcp-cat-service")
			require.NoError(t, err)
			toolName, err := util.GenerateToolID(serviceKey, "cat")
			require.NoError(t, err)

			// Wait for the tool to be available
			require.Eventually(t, func() bool {
				result, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
				if err != nil {
					t.Logf("Failed to list tools: %v", err)
					return false
				}
				for _, tool := range result.Tools {
					if strings.HasPrefix(tool.Name, "mcp-cat-service") {
						return true
					}
				}
				t.Logf("Tool %s not yet available", toolName)
				return false
			}, integration.TestWaitTimeMedium, 1*time.Second, "Tool %s did not become available in time", toolName)

			params := json.RawMessage(`{"content": "hello"}`)

			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: params})
			require.NoError(t, err, "Error calling tool '%s'", toolName)
			require.NotNil(t, res, "Nil response from tool '%s'", toolName)
			require.Len(t, res.Content, 1, "Expected exactly one content item from tool '%s'", toolName)

			textContent, ok := res.Content[0].(*mcp.TextContent)
			require.True(t, ok, "Expected content to be of type TextContent")
			t.Logf("Tool output: %s", textContent.Text)
			require.Contains(t, textContent.Text, "hello", "Expected response to contain 'hello'")
		},
	}

	framework.RunE2ETest(t, testCase)
}

func TestMCPUpstreamExample_HTTP(t *testing.T) {
	// Start the upstream http server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"content": [{"text": "woof"}]}`)
	}))
	defer server.Close()

	testCase := &framework.E2ETestCase{
		Name:                "MCP Upstream HTTP Example",
		UpstreamServiceType: "mcp-http",
		GenerateUpstreamConfig: func(upstreamEndpoint string) string {
			root, err := integration.GetProjectRoot()
			require.NoError(t, err)
			return filepath.Join(root, "examples/upstream/mcp/http/mcpxy_config.yaml")
		},
		InvokeAIClient: func(t *testing.T, mcpxyEndpoint string) {
			ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
			defer cancel()

			testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
			cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpxyEndpoint}, nil)
			require.NoError(t, err, "Failed to connect to MCPXY server")
			defer cs.Close()

			serviceKey, err := util.GenerateID("mcp-dog-service")
			require.NoError(t, err)
			toolName, err := util.GenerateToolID(serviceKey, "get_dog")
			require.NoError(t, err)

			// Wait for the tool to be available
			require.Eventually(t, func() bool {
				result, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
				if err != nil {
					t.Logf("Failed to list tools: %v", err)
					return false
				}
				for _, tool := range result.Tools {
					if strings.HasPrefix(tool.Name, "mcp-dog-service") {
						return true
					}
				}
				t.Logf("Tool %s not yet available", toolName)
				return false
			}, integration.TestWaitTimeMedium, 1*time.Second, "Tool %s did not become available in time", toolName)

			params := json.RawMessage(`{}`)

			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: params})
			require.NoError(t, err, "Error calling tool '%s'", toolName)
			require.NotNil(t, res, "Nil response from tool '%s'", toolName)
			require.Len(t, res.Content, 1, "Expected exactly one content item from tool '%s'", toolName)

			textContent, ok := res.Content[0].(*mcp.TextContent)
			require.True(t, ok, "Expected content to be of type TextContent")
			t.Logf("Tool output: %s", textContent.Text)
			require.Contains(t, textContent.Text, "woof", "Expected response to contain 'woof'")
		},
	}

	framework.RunE2ETest(t, testCase)
}
