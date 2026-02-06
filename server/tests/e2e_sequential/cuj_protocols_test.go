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

	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

// TestCUJ_Protocols covers CUJs 6-10: HTTP(SSE), External integrations, Errors, etc.
func TestCUJ_Protocols(t *testing.T) {
	if !integration.IsDockerSocketAccessible() {
		t.Skip("Skipping E2E test. Docker socket not accessible.")
	}

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

	// Create Docker Network
	networkName := fmt.Sprintf("mcpany-net-%d", time.Now().UnixNano())
	out, err := exec.Command("docker", "network", "create", networkName).CombinedOutput()
	require.NoError(t, err, "Failed to create network: %s", string(out))
	defer exec.Command("docker", "network", "rm", networkName).Run()

	// 1. Start Upstream Server (Backend)
	upstreamName := fmt.Sprintf("mcpany-upstream-%d", time.Now().UnixNano())
	upstreamConfigDir := filepath.Join(configDir, "upstream")
	err = os.MkdirAll(upstreamConfigDir, 0755)
	require.NoError(t, err)

	upstreamConfig := `
global_settings:
  mcp_listen_address: ":50050"
upstream_services:
  - id: "backend-fs"
    name: "Backend FS"
    disable: false
    filesystem_service:
      root_paths:
        "/data": "/data"
      os: {}
`
	err = os.WriteFile(filepath.Join(upstreamConfigDir, "config.yaml"), []byte(upstreamConfig), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(upstreamConfigDir, "backend_file.txt"), []byte("I am backend"), 0644)
	require.NoError(t, err)

	upstreamCmd := exec.Command("docker", "run", "-d", "--name", upstreamName,
		"--network", networkName,
		"--network-alias", "upstream",
		"-p", "25010:50050",
		"-v", fmt.Sprintf("%s:/mcp_config", upstreamConfigDir),
		"-v", fmt.Sprintf("%s:/data", upstreamConfigDir),
		"mcpany/server:latest",
		"run", "--config-path", "/mcp_config/config.yaml", "--mcp-listen-address", ":50050", "--debug", "--api-key", "test-key",
	)
	out, err = upstreamCmd.CombinedOutput()
	require.NoError(t, err, "Failed to start upstream: %s", string(out))
	defer exec.Command("docker", "rm", "-f", upstreamName).Run()

	// 2. Start Gateway Server (Frontend)
	gatewayName := fmt.Sprintf("mcpany-gateway-%d", time.Now().UnixNano())
	gatewayConfigDir := filepath.Join(configDir, "gateway")
	err = os.MkdirAll(gatewayConfigDir, 0755)
	require.NoError(t, err)

	gatewayConfig := `
global_settings:
  mcp_listen_address: ":50050"
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
        http_address: "http://upstream:50050/sse"
`
	err = os.WriteFile(filepath.Join(gatewayConfigDir, "config.yaml"), []byte(gatewayConfig), 0644)
	require.NoError(t, err)

	gatewayCmd := exec.Command("docker", "run", "-d", "--name", gatewayName,
		"--network", networkName,
		"-p", "25011:50050",
		"-v", fmt.Sprintf("%s:/mcp_config", gatewayConfigDir),
		"mcpany/server:latest",
		"run", "--config-path", "/mcp_config/config.yaml", "--mcp-listen-address", ":50050", "--debug", "--api-key", "test-key",
	)
	out, err = gatewayCmd.CombinedOutput()
	require.NoError(t, err, "Failed to start gateway: %s", string(out))
	defer func() {
		// Always print logs for debugging
		logs, err := exec.Command("docker", "logs", gatewayName).CombinedOutput()
		if err == nil {
			t.Logf("Gateway Logs:\n%s", string(logs))
		} else {
			t.Logf("Failed to get Gateway logs: %v", err)
		}
		logsUp, err := exec.Command("docker", "logs", upstreamName).CombinedOutput()
		if err == nil {
			t.Logf("Upstream Logs:\n%s", string(logsUp))
		}
		exec.Command("docker", "rm", "-f", gatewayName).Run()
	}()

	// Discover Gateway Port
	var portStr string
	require.Eventually(t, func() bool {
		out, err := exec.Command("docker", "port", gatewayName, "50050/tcp").Output()
		if err != nil {
			return false
		}
		portBinding := strings.TrimSpace(string(out))
		if idx := strings.Index(portBinding, "\n"); idx != -1 {
			portBinding = portBinding[:idx]
		}
		_, p, err := net.SplitHostPort(portBinding)
		if err != nil {
			return false
		}
		portStr = p
		return true
	}, 10*time.Second, 500*time.Millisecond)

	baseURL := fmt.Sprintf("http://127.0.0.1:%s", portStr)

	require.Eventually(t, func() bool {
		resp, err := http.Get(fmt.Sprintf("%s/healthz", baseURL))
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == 200
	}, 60*time.Second, 1*time.Second, "Gateway did not become healthy")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := mcp.NewClient(&mcp.Implementation{Name: "cuj-client", Version: "1.0"}, nil)
	transport := &mcp.StreamableClientTransport{Endpoint: baseURL + "/mcp?api_key=test-key"}
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
