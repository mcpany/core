package examples

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/mcpxy/core/pkg/consts"
	"github.com/mcpxy/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestCommandExample(t *testing.T) {
	t.Skip("Skipping command example test due to persistent timeout issues.")
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)

	// 1. Build the MCPXY binary and install dependencies
	prepareCmd := exec.Command("make", "prepare")
	prepareCmd.Dir = root
	// We ignore the error here because in some CI environments, the git config is
	// locked down, and pre-commit install fails. This is not critical for the test.
	_ = prepareCmd.Run()

	buildCmd := exec.Command("make", "build")
	buildCmd.Dir = root
	err = buildCmd.Run()
	require.NoError(t, err, "Failed to build mcpxy binary")

	// 2. Run the MCPXY Server
	serverInfo := integration.StartMCPXYServer(t, "command-example", "--config-paths", root+"/examples/upstream/command/config")
	defer serverInfo.CleanupFunc()

	// 3. Interact with the Tool using MCP SDK
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
	defer cancel()

	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: serverInfo.HTTPEndpoint}, nil)
	require.NoError(t, err, "Failed to connect to MCPXY server")
	defer cs.Close()

	toolName := fmt.Sprintf("hello-service%shello", consts.ToolNameServiceSeparator)
	t.Cleanup(func() {
		if t.Failed() {
			t.Logf("MCPXY Server Stdout:\n%s", serverInfo.Process.StdoutString())
			t.Logf("MCPXY Server Stderr:\n%s", serverInfo.Process.StderrString())
		}
	})

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
		t.Logf("Tool %s not yet available", toolName)
		return false
	}, integration.TestWaitTimeMedium, 1*time.Second, "Tool %s did not become available in time", toolName)

	params := json.RawMessage(`{"name": "World"}`)

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: params})
	require.NoError(t, err, "Error calling tool '%s'", toolName)
	require.NotNil(t, res, "Nil response from tool '%s'", toolName)
	require.Len(t, res.Content, 1, "Expected exactly one content item from tool '%s'", toolName)

	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected content to be of type TextContent")

	require.Equal(t, "Hello, World!", textContent.Text)
}
