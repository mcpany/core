//go:build e2e

package gmail_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_Gmail(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Gmail Server...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EGmailServerTest", "--config-path", "../../../../../examples/popular_services/google/gmail")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 2. Call Tool via MCPANY ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpAnyTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)
	require.NotEmpty(t, listToolsResult.Tools, "Expected at least one tool to be registered")
	t.Logf("Discovered %d tools from MCPANY", len(listToolsResult.Tools))

	// --- 3. Find the gmail.users.messages.list tool ---
	var gmailListTool *mcp.Tool
	for _, tool := range listToolsResult.Tools {
		if tool.Name == "gmail/-/gmail.users.messages.list" {
			gmailListTool = tool
			break
		}
	}
	require.NotNil(t, gmailListTool, "Expected to find the gmail.users.messages.list tool")

	// --- 4. Call the gmail.users.messages.list tool ---
	args := json.RawMessage(`{"userId": "me"}`)
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: gmailListTool.Name, Arguments: args})
	require.NoError(t, err)
	require.NotNil(t, res)

	// --- 5. Assert Response ---
	require.Len(t, res.Content, 1, "Expected exactly one content item")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content")

	var gmailResponse map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &gmailResponse)
	require.NoError(t, err, "Failed to unmarshal JSON response")

	require.Contains(t, gmailResponse, "messages", "The response should contain a list of messages")
}
