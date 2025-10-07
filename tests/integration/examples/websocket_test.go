package examples

import (
	"context"
	"encoding/json"
	"bytes"
	"fmt"
	"net"
	"os/exec"
	"testing"
	"time"

	"github.com/mcpxy/core/pkg/consts"
	"github.com/mcpxy/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestWebsocketExample(t *testing.T) {
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)

	// 1. Build the MCPXY binary
	buildCmd := exec.Command("make", "build")
	buildCmd.Dir = root
	err = buildCmd.Run()
	require.NoError(t, err, "Failed to build mcpxy binary")

	// 2. Run the Upstream Websocket Server
	upstreamServerCmd := exec.Command("go", "run", "main.go")
	upstreamServerCmd.Dir = root + "/examples/upstream/websocket/echo_server/server"
	err = upstreamServerCmd.Start()
	require.NoError(t, err, "Failed to start upstream websocket server")
	defer upstreamServerCmd.Process.Kill()

	// 3. Run the MCPXY Server
	mcpxyServerCmd := exec.Command("./start.sh")
	mcpxyServerCmd.Dir = root + "/examples/upstream/websocket"
	var stdout, stderr bytes.Buffer
	mcpxyServerCmd.Stdout = &stdout
	mcpxyServerCmd.Stderr = &stderr
	err = mcpxyServerCmd.Start()
	require.NoError(t, err, "Failed to start MCPXY server")
	defer func() {
		if t.Failed() {
			t.Logf("MCPXY Server stdout:\n%s", stdout.String())
			t.Logf("MCPXY Server stderr:\n%s", stderr.String())
		}
		mcpxyServerCmd.Process.Kill()
	}()

	// Wait for the MCPXY server to be ready
	require.Eventually(t, func() bool {
		conn, err := net.DialTimeout("tcp", "localhost:8080", 1*time.Second)
		if err != nil {
			return false
		}
		defer conn.Close()
		return true
	}, 10*time.Second, 100*time.Millisecond, "MCPXY server did not become available on port 8080")

	// 4. Interact with the Tool using MCP SDK
	ctx, cancel := context.WithTimeout(context.Background(), TestWaitTimeLong)
	defer cancel()

	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: "http://localhost:8080"}, nil)
	if err != nil {
		t.Logf("MCPXY Server stdout on connect error:\n%s", stdout.String())
		t.Logf("MCPXY Server stderr on connect error:\n%s", stderr.String())
	}
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
		t.Logf("Tool %s not yet available", toolName)
		return false
	}, TestWaitTimeLong, 1*time.Second, "Tool %s did not become available in time", toolName)

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
