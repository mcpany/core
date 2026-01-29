//go:build e2e

package qurancloud_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_QuranCloud(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Quran Cloud Server...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EQuranCloudServerTest", "--config-path", "../../../../examples/popular_services/qurancloud")
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
	t.Run("getAyah", func(t *testing.T) {
		args := json.RawMessage(`{"ayah": "2:255"}`)
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "getAyah", Arguments: args})
		require.NoError(t, err)
		require.NotNil(t, res)

		require.Len(t, res.Content, 1, "Expected exactly one content item")
		textContent, ok := res.Content[0].(*mcp.TextContent)
		require.True(t, ok, "Expected text content")

		var response map[string]interface{}
		err = json.Unmarshal([]byte(textContent.Text), &response)
		require.NoError(t, err, "Failed to unmarshal JSON response")

		require.Contains(t, response, "data")
		data := response["data"].(map[string]interface{})
		require.Contains(t, data, "text")
	})

	t.Run("getSurah", func(t *testing.T) {
		args := json.RawMessage(`{"surah": "1"}`)
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "getSurah", Arguments: args})
		require.NoError(t, err)
		require.NotNil(t, res)

		require.Len(t, res.Content, 1, "Expected exactly one content item")
		textContent, ok := res.Content[0].(*mcp.TextContent)
		require.True(t, ok, "Expected text content")

		var response map[string]interface{}
		err = json.Unmarshal([]byte(textContent.Text), &response)
		require.NoError(t, err, "Failed to unmarshal JSON response")

		require.Contains(t, response, "data")
		data := response["data"].(map[string]interface{})
		require.Equal(t, "Al-Faatiha", data["englishName"])
	})

	t.Log("INFO: E2E Test Scenario for Quran Cloud Server Completed Successfully!")
}
