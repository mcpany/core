package examples

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mcpxy/core/pkg/consts"
	"github.com/mcpxy/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

const (
	TestWaitTimeLong = 5 * time.Minute
)

func TestCommandExample(t *testing.T) {
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)

	// 1. Make the script executable
	scriptPath := filepath.Join(root, "examples", "upstream", "command", "server", "hello.sh")
	err = os.Chmod(scriptPath, 0755)
	require.NoError(t, err, "Failed to make hello.sh executable")

	// 2. Start the MCP-XY Server on a dynamic port
	configDir := filepath.Join(root, "examples", "upstream", "command", "config")
	mcpxyServer := integration.StartMCPXYServer(t, "CommandExample", "--config-paths", configDir)
	defer mcpxyServer.CleanupFunc()

	// 3. Interact with the Tool using MCP SDK
	ctx, cancel := context.WithTimeout(context.Background(), TestWaitTimeLong)
	defer cancel()

	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpxyServer.JSONRPCEndpoint}, nil)
	require.NoError(t, err, "Failed to connect to MCPXY server")
	defer cs.Close()

	toolName := fmt.Sprintf("hello-service%shello.sh", consts.ToolNameServiceSeparator)

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

	params := json.RawMessage(`{}`)

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: params})
	require.NoError(t, err, "Error calling tool '%s'", toolName)
	require.NotNil(t, res, "Nil response from tool '%s'", toolName)
	require.Len(t, res.Content, 1, "Expected exactly one content item from tool '%s'", toolName)

	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected content to be of type TextContent")

	require.Equal(t, "Hello from command upstream!\n", textContent.Text)
}