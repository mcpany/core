//go:build e2e

// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package e2e_sequential

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

// TestProcessStack replaces TestDockerComposeE2E for environments without Docker.
// It verifies the integration of an external HTTP service via the server.
func TestProcessStack(t *testing.T) {
	// 1. Start External HTTP Echo Server (Simulating the 3rd party service)
	portEcho := findFreePort(t)
	echoMux := http.NewServeMux()
	echoMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		fmt.Printf("Echo Server Received: %s\n", string(body))
		w.Header().Set("Content-Type", "application/json")
		// Echo back the body + some metadata
		// The tool configuration expects the response to be used.
		// For simplicity, we just return what we received wrapped or as is.
		w.Write(body)
	})
	echoServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", portEcho),
		Handler: echoMux,
	}
	go func() {
		_ = echoServer.ListenAndServe()
	}()
	defer echoServer.Close()

	// 2. Prepare MCP Any Server Config
	rootDir, err := os.Getwd()
	require.NoError(t, err)
	if strings.HasSuffix(rootDir, "tests/e2e_sequential") {
		rootDir = filepath.Join(rootDir, "../../..")
	} else if strings.HasSuffix(rootDir, "server") {
		rootDir = filepath.Join(rootDir, "..")
	}
	rootDir, err = filepath.Abs(rootDir)
	require.NoError(t, err)

	configDir := filepath.Join(rootDir, "build", "e2e_config_stack")
	err = os.MkdirAll(configDir, 0755)
	require.NoError(t, err)
	defer os.RemoveAll(configDir)

	portMcp := findFreePort(t)
	configContent := fmt.Sprintf(`
global_settings:
  mcp_listen_address: "127.0.0.1:%d"
upstream_services:
  - id: "http-echo"
    name: "HTTP Echo"
    disable: false
    http_service:
      address: "http://127.0.0.1:%d"
      tools:
        - name: "echo"
          description: "Echo tool"
          call_id: "echo_call"
          input_schema:
            type: "object"
            properties:
              message:
                type: "string"
      calls:
        echo_call:
          endpoint_path: "/"
          method: "HTTP_METHOD_POST"
          parameters:
            - schema:
                name: "message"
                type: "STRING"
`, portMcp, portEcho)

	configPath := filepath.Join(configDir, "config.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// 3. Build and Start MCP Any Server
	binPath := buildServer(t, rootDir)
	cmd, baseURL := startServer(t, binPath, configPath, portMcp)
	defer func() {
		if cmd != nil && cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	// 4. Verify Integration
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := mcp.NewClient(&mcp.Implementation{Name: "stack-client", Version: "1.0"}, nil)
	transport := &mcp.StreamableClientTransport{Endpoint: baseURL + "/mcp?api_key=test-key"}
	session, err := client.Connect(ctx, transport, nil)
	require.NoError(t, err)
	defer session.Close()

	// Verify Tool Exists
	require.Eventually(t, func() bool {
		list, err := session.ListTools(ctx, nil)
		if err != nil {
			return false
		}
		for _, tool := range list.Tools {
			if strings.Contains(tool.Name, "echo") {
				return true
			}
		}
		return false
	}, 10*time.Second, 500*time.Millisecond, "Echo tool not found")

	// Call Tool
	// We need to match the tool definition.
	// The definition expects "message" (string).
	// The HTTP service receives JSON body.
	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "echo", // Might be prefixed? "http-echo.echo" or just "echo" depending on config.
		// Usually name is prefixed by service ID if auto-discovery or explicit mapping?
		// Explicit config usually preserves name unless collision.
		// But in other tests we saw prefixes.
		// Let's assume we need to find the name first.
		Arguments: map[string]interface{}{
			"message": "Hello Process Stack",
		},
	})

	// Helper to find actual name
	list, _ := session.ListTools(ctx, nil)
	var toolName string
	for _, tool := range list.Tools {
		if strings.HasSuffix(tool.Name, "echo") {
			toolName = tool.Name
			break
		}
	}
	require.NotEmpty(t, toolName)

	res, err = session.CallTool(ctx, &mcp.CallToolParams{
		Name: toolName,
		Arguments: map[string]interface{}{
			"message": "Hello Process Stack",
		},
	})
	require.NoError(t, err)

	// Verify Output
	// The echo server returns the request body.
	// The tool execution returns that body as text (default).
	require.NotEmpty(t, res.Content)
	if txt, ok := res.Content[0].(*mcp.TextContent); ok {
		require.Contains(t, txt.Text, "Hello Process Stack")
	} else {
		require.Fail(t, "Expected TextContent")
	}
}
