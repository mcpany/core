//go:build e2e

package poetrydb_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_PoetryDB(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for PoetryDB Server...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EPoetryDBServerTest", "--config-path", "../../../../examples/popular_services/poetrydb")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 2. Call Tool via MCPANY ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpAnyTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)
	require.Len(t, listToolsResult.Tools, 2, "Expected exactly two tools to be registered")

	// --- 3. Test Cases ---
	t.Run("getPoemsByAuthor", func(t *testing.T) {
		args := json.RawMessage(`{"author_name": "Emily Dickinson"}`)
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "getPoemsByAuthor", Arguments: args})
		require.NoError(t, err)
		require.NotNil(t, res)

		require.Len(t, res.Content, 1, "Expected exactly one content item")
		textContent, ok := res.Content[0].(*mcp.TextContent)
		require.True(t, ok, "Expected text content")

		var response []interface{}
		err = json.Unmarshal([]byte(textContent.Text), &response)
		require.NoError(t, err, "Failed to unmarshal JSON response")

		require.True(t, len(response) > 0)
		poemInfo := response[0].(map[string]interface{})
		require.Equal(t, "Emily Dickinson", poemInfo["author"])
	})

	t.Run("getPoemByTitle", func(t *testing.T) {
		args := json.RawMessage(`{"title": "The Raven"}`)
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "getPoemByTitle", Arguments: args})
		require.NoError(t, err)
		require.NotNil(t, res)

		require.Len(t, res.Content, 1, "Expected exactly one content item")
		textContent, ok := res.Content[0].(*mcp.TextContent)
		require.True(t, ok, "Expected text content")

		var response []interface{}
		err = json.Unmarshal([]byte(textContent.Text), &response)
		require.NoError(t, err, "Failed to unmarshal JSON response")

		require.True(t, len(response) > 0)
		poemInfo := response[0].(map[string]interface{})
		require.Equal(t, "The Raven", poemInfo["title"])
	})

	t.Log("INFO: E2E Test Scenario for PoetryDB Server Completed Successfully!")
}
