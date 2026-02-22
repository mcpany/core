//go:build e2e

// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package e2e_sequential

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

// TestCUJ_Protocols covers CUJs 6-10: HTTP(SSE), External integrations, Errors, etc.
func TestCUJ_Protocols(t *testing.T) {
	// Prioritize local execution
	useLocal := true

	rootDir, err := os.Getwd()
	require.NoError(t, err)
	if strings.HasSuffix(rootDir, "tests/e2e_sequential") {
		rootDir = filepath.Join(rootDir, "../../..")
	} else if strings.HasSuffix(rootDir, "server") {
		rootDir = filepath.Join(rootDir, "..")
	}
	rootDir, err = filepath.Abs(rootDir)
	require.NoError(t, err)

	configDir := filepath.Join(rootDir, "build", "e2e_config_protocols")
	err = os.MkdirAll(configDir, 0755)
	require.NoError(t, err)
	defer os.RemoveAll(configDir)

	// Build Server
	binPath := buildServer(t, rootDir)

	// 1. Start Upstream Server (Backend)
	portUp := findFreePort(t)
	upstreamConfigDir := filepath.Join(configDir, "upstream")
	err = os.MkdirAll(upstreamConfigDir, 0755)
	require.NoError(t, err)

	upstreamConfig := fmt.Sprintf(`
global_settings:
  mcp_listen_address: "127.0.0.1:%d"
upstream_services:
  - id: "backend-fs"
    name: "Backend FS"
    disable: false
    filesystem_service:
      root_paths:
        "/data": "%s"
      os: {}
`, portUp, upstreamConfigDir) // Use absolute path for filesystem root
	configPathUp := filepath.Join(upstreamConfigDir, "config.yaml")
	err = os.WriteFile(configPathUp, []byte(upstreamConfig), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(upstreamConfigDir, "backend_file.txt"), []byte("I am backend"), 0644)
	require.NoError(t, err)

	var upstreamCmd *exec.Cmd
	// var upstreamBaseURL string // unused but returned by startServer

	if useLocal {
		upstreamCmd, _ = startServer(t, binPath, configPathUp, portUp)
	} else {
		t.Skip("Local execution required for this test refactor")
	}
	defer func() {
		if upstreamCmd != nil && upstreamCmd.Process != nil {
			upstreamCmd.Process.Kill()
		}
	}()

	// 2. Start Gateway Server (Frontend)
	portGw := findFreePort(t)
	gatewayConfigDir := filepath.Join(configDir, "gateway")
	err = os.MkdirAll(gatewayConfigDir, 0755)
	require.NoError(t, err)

	gatewayConfig := fmt.Sprintf(`
global_settings:
  mcp_listen_address: "127.0.0.1:%d"
upstream_services:
  - id: "proxy-service"
    name: "Proxy Service"
    disable: false
    upstream_auth:
      api_key:
        in: QUERY
        param_name: "api_key"
        value:
          plain_text: "test-key"
    mcp_service:
      tool_auto_discovery: true
      http_connection:
        http_address: "http://127.0.0.1:%d/sse"
`, portGw, portUp)
	configPathGw := filepath.Join(gatewayConfigDir, "config.yaml")
	err = os.WriteFile(configPathGw, []byte(gatewayConfig), 0644)
	require.NoError(t, err)

	var gatewayCmd *exec.Cmd
	var gatewayBaseURL string

	if useLocal {
		gatewayCmd, gatewayBaseURL = startServer(t, binPath, configPathGw, portGw)
	}
	defer func() {
		if gatewayCmd != nil && gatewayCmd.Process != nil {
			gatewayCmd.Process.Kill()
		}
	}()

	// Verify Gateway Health
	verifyEndpoint(t, fmt.Sprintf("%s/healthz", gatewayBaseURL), 200, 60*time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := mcp.NewClient(&mcp.Implementation{Name: "cuj-client", Version: "1.0"}, nil)
	transport := &mcp.StreamableClientTransport{Endpoint: gatewayBaseURL + "/mcp?api_key=test-key"}
	session, err := client.Connect(ctx, transport, nil)
	require.NoError(t, err)
	defer session.Close()

	// Verify Proxied Tools
	require.Eventually(t, func() bool {
		list, err := session.ListTools(ctx, nil)
		if err != nil {
			return false
		}
		for _, tool := range list.Tools {
			if strings.Contains(tool.Name, "list_directory") {
				return true
			}
		}
		return false
	}, 60*time.Second, 1*time.Second, "Failed to find proxied list_directory tool")

	// Call tool
	list, err := session.ListTools(ctx, nil)
	require.NoError(t, err)
	var listToolName string
	for _, tool := range list.Tools {
		if strings.Contains(tool.Name, "list_directory") {
			listToolName = tool.Name
			break
		}
	}
	require.NotEmpty(t, listToolName)

	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: listToolName,
		Arguments: map[string]interface{}{
			"path": "/data",
		},
	})
	require.NoError(t, err)

	foundFile := false
	for _, c := range res.Content {
		if txt, ok := c.(*mcp.TextContent); ok {
			if strings.Contains(txt.Text, "backend_file.txt") {
				foundFile = true
			}
		}
	}
	require.True(t, foundFile, "Result did not contain backend_file.txt in %v", res.Content)
}
