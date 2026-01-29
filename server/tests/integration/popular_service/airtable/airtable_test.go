//go:build e2e

package airtable_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_Airtable(t *testing.T) {
	if os.Getenv("AIRTABLE_API_TOKEN") == "" {
		// t.Skip("AIRTABLE_API_TOKEN is not set")
	}
	if os.Getenv("AIRTABLE_BASE_ID") == "" {
		// t.Skip("AIRTABLE_BASE_ID is not set")
	}
	if os.Getenv("AIRTABLE_TABLE_ID") == "" {
		// t.Skip("AIRTABLE_TABLE_ID is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Airtable Server...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EAirtableServerTest", "--config-path", "../../../../examples/popular_services/airtable")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 2. Call Tool via MCPANY ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpAnyTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)
	require.Len(t, listToolsResult.Tools, 1, "Expected exactly one tool to be registered")

	// --- 3. Test Cases ---
	t.Run("list_records", func(t *testing.T) {
		args := json.RawMessage(`{"baseId": "` + os.Getenv("AIRTABLE_BASE_ID") + `", "tableId": "` + os.Getenv("AIRTABLE_TABLE_ID") + `"}`)
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "airtable-list_records", Arguments: args})
		require.NoError(t, err)
		require.NotNil(t, res)

		require.Len(t, res.Content, 1, "Expected exactly one content item")
		textContent, ok := res.Content[0].(*mcp.TextContent)
		require.True(t, ok, "Expected text content")

		var response map[string]interface{}
		err = json.Unmarshal([]byte(textContent.Text), &response)
		require.NoError(t, err, "Failed to unmarshal JSON response")

		require.Contains(t, response, "records", "The response should contain records")
	})

	t.Log("INFO: E2E Test Scenario for Airtable Server Completed Successfully!")
}
