//go:build e2e

// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package e2e_sequential

import (
	"context"
	"fmt"
	"net"
	"net/http"
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
	// t.Skip("Skipping E2E test as requested by user to unblock merge")

	rootDir, err := os.Getwd()
	require.NoError(t, err)
	if strings.HasSuffix(rootDir, "tests/e2e_sequential") {
		rootDir = filepath.Join(rootDir, "../../..")
	} else if strings.HasSuffix(rootDir, "server") {
		rootDir = filepath.Join(rootDir, "..")
	}
	rootDir, err = filepath.Abs(rootDir)
	require.NoError(t, err)

	serverBin := filepath.Join(rootDir, "build", "bin", "server")
	if _, err := os.Stat(serverBin); os.IsNotExist(err) {
		t.Skip("Server binary not found, skipping test. Run 'make build' first.")
	}

	configDir := filepath.Join(rootDir, "build", "e2e_config_protocols")
	err = os.MkdirAll(configDir, 0755)
	require.NoError(t, err)
	defer os.RemoveAll(configDir)

	// 1. Start Upstream Server (Backend)
	upstreamPort := findFreePort(t)
	upstreamConfigDir := filepath.Join(configDir, "upstream")
	err = os.MkdirAll(upstreamConfigDir, 0755)
	require.NoError(t, err)

	upstreamConfig := fmt.Sprintf(`
global_settings:
  mcp_listen_address: "127.0.0.1:%s"
  api_key: "test-key"
upstream_services:
  - id: "backend-fs"
    name: "Backend FS"
    disable: false
    filesystem_service:
      root_paths:
        "/data": "%s"
      os: {}
`, upstreamPort, upstreamConfigDir)

	err = os.WriteFile(filepath.Join(upstreamConfigDir, "config.yaml"), []byte(upstreamConfig), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(upstreamConfigDir, "backend_file.txt"), []byte("I am backend"), 0644)
	require.NoError(t, err)

	upstreamCmd := exec.Command(serverBin, "run", "--config-path", filepath.Join(upstreamConfigDir, "config.yaml"), "--debug", "--api-key", "test-key")
	upstreamCmd.Env = os.Environ() // Inherit env
	// Capture output
	upstreamOut, err := os.Create(filepath.Join(upstreamConfigDir, "server.log"))
	require.NoError(t, err)
	upstreamCmd.Stdout = upstreamOut
	upstreamCmd.Stderr = upstreamOut

	require.NoError(t, upstreamCmd.Start())
	defer func() {
		_ = upstreamCmd.Process.Kill()
		upstreamOut.Close()
	}()

	// Wait for Upstream
	verifyEndpoint(t, fmt.Sprintf("http://127.0.0.1:%s/healthz", upstreamPort), 200, 10*time.Second)

	// 2. Start Gateway Server (Frontend)
	gatewayPort := findFreePort(t)
	gatewayConfigDir := filepath.Join(configDir, "gateway")
	err = os.MkdirAll(gatewayConfigDir, 0755)
	require.NoError(t, err)

	gatewayConfig := fmt.Sprintf(`
global_settings:
  mcp_listen_address: "127.0.0.1:%s"
  api_key: "test-key"
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
        http_address: "http://127.0.0.1:%s/mcp/sse"
`, gatewayPort, upstreamPort)

	err = os.WriteFile(filepath.Join(gatewayConfigDir, "config.yaml"), []byte(gatewayConfig), 0644)
	require.NoError(t, err)

	gatewayCmd := exec.Command(serverBin, "run", "--config-path", filepath.Join(gatewayConfigDir, "config.yaml"), "--debug", "--api-key", "test-key")
	gatewayCmd.Env = os.Environ()
	gatewayOut, err := os.Create(filepath.Join(gatewayConfigDir, "server.log"))
	require.NoError(t, err)
	gatewayCmd.Stdout = gatewayOut
	gatewayCmd.Stderr = gatewayOut

	require.NoError(t, gatewayCmd.Start())
	defer func() {
		_ = gatewayCmd.Process.Kill()
		gatewayOut.Close()
	}()

	baseURL := fmt.Sprintf("http://127.0.0.1:%s", gatewayPort)

	require.Eventually(t, func() bool {
		resp, err := http.Get(fmt.Sprintf("%s/healthz", baseURL))
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == 200
	}, 20*time.Second, 1*time.Second, "Gateway did not become healthy")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := mcp.NewClient(&mcp.Implementation{Name: "cuj-client", Version: "1.0"}, nil)
	transport := &mcp.StreamableClientTransport{Endpoint: baseURL + "/mcp/sse?api_key=test-key"}
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
	}, 20*time.Second, 1*time.Second, "Failed to find proxied list_directory tool")

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

func findFreePort(t *testing.T) string {
    l, err := net.Listen("tcp", "127.0.0.1:0")
    require.NoError(t, err)
    defer l.Close()
    _, port, err := net.SplitHostPort(l.Addr().String())
    require.NoError(t, err)
    return port
}
