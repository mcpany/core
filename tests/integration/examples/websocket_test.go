package examples

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mcpxy/core/pkg/util"
	"github.com/mcpxy/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestWebsocketExample(t *testing.T) {
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)

	// Find a free port for the upstream server
	port := integration.FindFreePort(t)

	// 2. Run the Upstream Websocket Server
	upstreamServerProcess := integration.NewManagedProcess(t, "upstream-websocket-server", "go",
		[]string{"run", "main.go"},
		[]string{"WEBSOCKET_PORT=" + strconv.Itoa(port)},
	)
	upstreamServerProcess.Cmd().Dir = filepath.Join(root, "examples", "upstream", "websocket", "echo_server", "server")
	err = upstreamServerProcess.Start()
	require.NoError(t, err, "Failed to start upstream websocket server")
	defer upstreamServerProcess.Stop()

	// Wait for the upstream server to be ready
	integration.WaitForHTTPHealth(t, fmt.Sprintf("http://localhost:%d/health", port), 10*time.Second)

	// Create a temporary config file with the dynamic port
	originalConfigPath := filepath.Join(root, "examples", "upstream", "websocket", "config", "mcpxy_config.yaml")
	configData, err := os.ReadFile(originalConfigPath)
	require.NoError(t, err)

	newConfigData := strings.Replace(string(configData), "ws://localhost:8082", fmt.Sprintf("ws://localhost:%d", port), 1)

	tempConfigFile, err := os.CreateTemp("", "mcpxy_config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tempConfigFile.Name())

	_, err = tempConfigFile.Write([]byte(newConfigData))
	require.NoError(t, err)
	err = tempConfigFile.Close()
	require.NoError(t, err)

	// 3. Run the MCPXY Server
	serverInfo := integration.StartMCPXYServer(t, "websocket-example", "--config-paths", tempConfigFile.Name())
	defer serverInfo.CleanupFunc()

	// Wait for the MCPXY server to be ready
	require.Eventually(t, func() bool {
		conn, err := net.DialTimeout("tcp", strings.TrimPrefix(serverInfo.JSONRPCEndpoint, "http://"), 1*time.Second)
		if err != nil {
			return false
		}
		defer conn.Close()
		return true
	}, 10*time.Second, 100*time.Millisecond, "MCPXY server did not become available on port %s", serverInfo.JSONRPCEndpoint)

	// 4. Interact with the Tool using MCP SDK
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
	defer cancel()

	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: serverInfo.JSONRPCEndpoint}, nil)
	require.NoError(t, err, "Failed to connect to MCPXY server")
	defer cs.Close()

	serviceID := "echo-service"
	serviceKey, _ := util.GenerateID(serviceID)
	toolName, _ := util.GenerateToolID(serviceKey, "echo")

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
