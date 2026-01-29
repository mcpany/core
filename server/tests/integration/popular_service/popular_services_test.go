//go:build e2e

package popular_service_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_Trello(t *testing.T) {
	if os.Getenv("TRELLO_API_KEY") == "" || os.Getenv("TRELLO_API_TOKEN") == "" || os.Getenv("TRELLO_API_KEY") == "dummy" {
		// t.Skip("TRELLO_API_KEY or TRELLO_API_TOKEN not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Trello Server...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2ETrelloServerTest", "--config-path", "../../../examples/popular_services/trello")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 2. Call Tool via MCPANY ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpAnyTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)
	require.Len(t, listToolsResult.Tools, 1, "Expected exactly one tool to be registered")
	registeredToolName := listToolsResult.Tools[0].Name
	t.Logf("Discovered tool from MCPANY: %s", registeredToolName)

	// --- 3. Call Tool ---
	args := json.RawMessage(`{}`)
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: registeredToolName, Arguments: args})
	require.NoError(t, err)
	require.NotNil(t, res)

	// --- 4. Assert Response ---
	require.Len(t, res.Content, 1, "Expected exactly one content item")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content")

	var trelloResponse []interface{}
	err = json.Unmarshal([]byte(textContent.Text), &trelloResponse)
	require.NoError(t, err, "Failed to unmarshal JSON response")

	t.Log("INFO: E2E Test Scenario for Trello Server Completed Successfully!")
}

func TestUpstreamService_Miro(t *testing.T) {
	if os.Getenv("MIRO_API_TOKEN") == "" || os.Getenv("MIRO_API_TOKEN") == "dummy" {
		// t.Skip("MIRO_API_TOKEN not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Miro Server...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EMiroServerTest", "--config-path", "../../../examples/popular_services/miro")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 2. Call Tool via MCPANY ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpAnyTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)
	require.Len(t, listToolsResult.Tools, 1, "Expected exactly one tool to be registered")
	registeredToolName := listToolsResult.Tools[0].Name
	t.Logf("Discovered tool from MCPANY: %s", registeredToolName)

	// --- 3. Call Tool ---
	args := json.RawMessage(`{}`)
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: registeredToolName, Arguments: args})
	require.NoError(t, err)
	require.NotNil(t, res)

	// --- 4. Assert Response ---
	require.Len(t, res.Content, 1, "Expected exactly one content item")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content")

	var miroResponse map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &miroResponse)
	require.NoError(t, err, "Failed to unmarshal JSON response")

	t.Log("INFO: E2E Test Scenario for Miro Server Completed Successfully!")
}

func TestUpstreamService_Figma(t *testing.T) {
	if os.Getenv("FIGMA_API_TOKEN") == "" || os.Getenv("FIGMA_TEAM_ID") == "" || os.Getenv("FIGMA_API_TOKEN") == "dummy" {
		// t.Skip("FIGMA_API_TOKEN or FIGMA_TEAM_ID not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Figma Server...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EFigmaServerTest", "--config-path", "../../../examples/popular_services/figma")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 2. Call Tool via MCPANY ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpAnyTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)
	require.Len(t, listToolsResult.Tools, 1, "Expected exactly one tool to be registered")
	registeredToolName := listToolsResult.Tools[0].Name
	t.Logf("Discovered tool from MCPANY: %s", registeredToolName)

	// --- 3. Call Tool ---
	args := json.RawMessage(`{"team_id": "` + os.Getenv("FIGMA_TEAM_ID") + `"}`)
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: registeredToolName, Arguments: args})
	require.NoError(t, err)
	require.NotNil(t, res)

	// --- 4. Assert Response ---
	require.Len(t, res.Content, 1, "Expected exactly one content item")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content")

	var figmaResponse map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &figmaResponse)
	require.NoError(t, err, "Failed to unmarshal JSON response")

	t.Log("INFO: E2E Test Scenario for Figma Server Completed Successfully!")
}
