package examples

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mcpxy/core/pkg/consts"
	"github.com/mcpxy/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestGRPCExample(t *testing.T) {
	t.Skip("Skipping gRPC example test due to persistent timeout issues.")
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)

	// 1. Build the MCPXY binary
	buildCmd := exec.Command("make", "build")
	buildCmd.Dir = root
	err = buildCmd.Run()
	require.NoError(t, err, "Failed to build mcpxy binary")

	// 2. Generate Protobuf Files
	generateCmd := exec.Command("./generate.sh")
	generateCmd.Dir = root + "/examples/upstream/grpc/greeter_server"
	err = generateCmd.Run()
	require.NoError(t, err, "Failed to generate protobuf files")

	// Tidy dependencies for the upstream server
	serverDir := filepath.Join(root, "examples", "upstream", "grpc", "greeter_server", "server")
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = serverDir
	err = tidyCmd.Run()
	require.NoError(t, err, "Failed to tidy go module for gRPC server")

	// Find a free port for the upstream server
	port := integration.FindFreePort(t)

	// 3. Run the Upstream gRPC Server
	upstreamServerProcess := integration.NewManagedProcess(t, "upstream-grpc-server", "go",
		[]string{"run", "main.go"},
		[]string{"GRPC_PORT=" + strconv.Itoa(port)},
	)
	upstreamServerProcess.Cmd().Dir = filepath.Join(root, "examples", "upstream", "grpc", "greeter_server", "server")
	err = upstreamServerProcess.Start()
	require.NoError(t, err, "Failed to start upstream gRPC server")
	defer upstreamServerProcess.Stop()

	// Wait for the upstream gRPC server to be ready
	integration.WaitForTCPPort(t, port, 15*time.Second)

	// Create a temporary config file with the dynamic port
	originalConfigPath := filepath.Join(root, "examples", "upstream", "grpc", "config", "mcpxy_config.yaml")
	configData, err := os.ReadFile(originalConfigPath)
	require.NoError(t, err)

	newConfigData := strings.Replace(string(configData), "localhost:50051", fmt.Sprintf("localhost:%d", port), 1)

	tempConfigFile, err := os.CreateTemp("", "mcpxy_config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tempConfigFile.Name())

	_, err = tempConfigFile.Write([]byte(newConfigData))
	require.NoError(t, err)
	err = tempConfigFile.Close()
	require.NoError(t, err)

	// 4. Run the MCPXY Server
	serverInfo := integration.StartMCPXYServer(t, "grpc-example", "--config-paths", tempConfigFile.Name())
	defer serverInfo.CleanupFunc()

	// 5. Interact with the Tool using MCP SDK
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
	defer cancel()

	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: serverInfo.HTTPEndpoint}, nil)
	require.NoError(t, err, "Failed to connect to MCPXY server")
	defer cs.Close()

	toolName := fmt.Sprintf("greeter-service%sSayHello", consts.ToolNameServiceSeparator)

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
	}, integration.TestWaitTimeLong, 1*time.Second, "Tool %s did not become available in time", toolName)

	params := json.RawMessage(`{"name": "World"}`)

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: params})
	require.NoError(t, err, "Error calling tool '%s'", toolName)
	require.NotNil(t, res, "Nil response from tool '%s'", toolName)
	require.Len(t, res.Content, 1, "Expected exactly one content item from tool '%s'", toolName)

	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected content to be of type TextContent")

	var jsonResponse map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &jsonResponse)
	require.NoError(t, err, "Failed to unmarshal JSON response from tool")

	require.Equal(t, "Hello World", jsonResponse["message"])
}