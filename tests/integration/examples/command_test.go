package examples

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/mcpxy/core/pkg/consts"
	"github.com/mcpxy/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestCommandExample(t *testing.T) {
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)

	// 1. Build the MCPXY binary
	buildCmd := exec.Command("make", "build")
	buildCmd.Dir = root
	err = buildCmd.Run()
	require.NoError(t, err, "Failed to build mcpxy binary")

	// 2. Make the script executable
	chmodCmd := exec.Command("chmod", "+x", "hello.sh")
	chmodCmd.Dir = filepath.Join(root, "examples", "upstream", "command", "server")
	err = chmodCmd.Run()
	require.NoError(t, err, "Failed to make hello.sh executable")

	// 3. Get free ports
	mcpxyHttpPort := getFreePort(t)
	mcpxyGrpcPort := getFreePort(t)

	// 4. Create a temporary MCPXY config
	mcpxyConfig := `
upstream_services:
- name: hello-service
  command_line_service:
    command: ./server/hello.sh
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "mcpxy_config.yaml")
	err = os.WriteFile(configPath, []byte(mcpxyConfig), 0644)
	require.NoError(t, err)

	// 5. Run the MCPXY Server
	mcpxyServerCmd := exec.Command(filepath.Join(root, "build", "bin", "server"),
		"--config-paths", configPath,
		"--grpc-port", strconv.Itoa(mcpxyGrpcPort),
		"--jsonrpc-port", strconv.Itoa(mcpxyHttpPort),
	)
	mcpxyServerCmd.Dir = filepath.Join(root, "examples", "upstream", "command")
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
	mcpxyAddr := fmt.Sprintf("localhost:%d", mcpxyHttpPort)
	require.Eventually(t, func() bool {
		conn, err := net.DialTimeout("tcp", mcpxyAddr, 1*time.Second)
		if err != nil {
			return false
		}
		defer conn.Close()
		return true
	}, 10*time.Second, 100*time.Millisecond, "MCPXY server did not become available on port %d", mcpxyHttpPort)

	// 6. Interact with the Tool using MCP SDK
	ctx, cancel := context.WithTimeout(context.Background(), TestWaitTimeLong)
	defer cancel()

	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: "http://" + mcpxyAddr}, nil)
	if err != nil {
		t.Logf("MCPXY Server stdout on connect error:\n%s", stdout.String())
		t.Logf("MCPXY Server stderr on connect error:\n%s", stderr.String())
	}
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
		t.Logf("Tool %s not yet available", toolName)
		return false
	}, TestWaitTimeLong, 1*time.Second, "Tool %s did not become available in time", toolName)

	params := json.RawMessage(`{}`)

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: params})
	require.NoError(t, err, "Error calling tool '%s'", toolName)
	require.NotNil(t, res, "Nil response from tool '%s'", toolName)
	require.Len(t, res.Content, 1, "Expected exactly one content item from tool '%s'", toolName)

	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected content to be of type TextContent")

	require.Equal(t, "Hello from command upstream!\n", textContent.Text)
}
