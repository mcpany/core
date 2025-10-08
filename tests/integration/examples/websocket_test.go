package examples

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/mcpxy/core/pkg/consts"
	"github.com/mcpxy/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestWebsocketExample(t *testing.T) {
	// 1. Start the upstream websocket server on a dynamic port
	upstreamPort := integration.FindFreePort(t)
	upstreamAddr := fmt.Sprintf("ws://localhost:%d/echo", upstreamPort)
	upstreamEnv := []string{fmt.Sprintf("WEBSOCKET_ECHO_SERVER_PORT=%d", upstreamPort)}

	root, err := integration.GetProjectRoot()
	require.NoError(t, err)

	upstreamServer := integration.NewManagedProcess(t, "websocket-echo-server", "go", []string{"run", "./websocket/echo_server/server/main.go"}, upstreamEnv, root+"/examples/upstream")
	err = upstreamServer.Start()
	require.NoError(t, err, "Failed to start upstream websocket server")
	defer upstreamServer.Stop()

	// 2. Start the MCP-XY Server on a dynamic port
	mcpxyServer := integration.StartMCPXYServer(t, "WebsocketExample")
	defer mcpxyServer.CleanupFunc()

	// 3. Register the upstream service with MCP-XY
	integration.RegisterWebsocketService(t, mcpxyServer.RegistrationClient, "echo-service", upstreamAddr, "/echo", nil)

	// 4. Interact with the Tool using MCP SDK
	ctx, cancel := context.WithTimeout(context.Background(), TestWaitTimeLong)
	defer cancel()

	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpxyServer.JSONRPCEndpoint}, nil)
	require.NoError(t, err, "Failed to connect to MCPXY server")
	defer cs.Close()

	toolName := fmt.Sprintf("echo-service%s/echo", consts.ToolNameServiceSeparator)

	// Wait for the tool to be available
	require.Eventually(t, func() bool {
		result, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
		if err != nil {
			t.Logf("Failed to list tools: %v", err)
			return false
		}
		for _, tool := range result.Tools {
			if tool.Name == toolName {
				return true
			}
		}
		return false
	}, 10*time.Second, 1*time.Second, "Tool %s did not become available in time", toolName)

	params := json.RawMessage(`{"message": "hello"}`)

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: params})
	require.NoError(t, err, "Error calling tool '%s'", toolName)
	require.NotNil(t, res, "Nil response from tool '%s'", toolName)
	require.Len(t, res.Content, 1, "Expected exactly one content item from tool '%s'", toolName)

	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected content to be of type TextContent")

	var jsonResponse map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &jsonResponse)
	require.NoError(t, err, "Failed to unmarshal JSON response from tool")

	require.Equal(t, "hello", jsonResponse["message"])
}