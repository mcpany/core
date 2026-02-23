//go:build e2e

// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package e2e_sequential

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

// TestCUJ_Protocols covers CUJs 6-10: HTTP(SSE), External integrations, Errors, etc.
func TestCUJ_Protocols(t *testing.T) {
	binPath := BuildServer(t)
	rootDir := findRootDir(t)

	configDir := filepath.Join(rootDir, "build", "e2e_config_protocols")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)
	defer os.RemoveAll(configDir)

	// 1. Start Upstream Server (Backend)
	upstreamConfigDir := filepath.Join(configDir, "upstream")
	err = os.MkdirAll(upstreamConfigDir, 0755)
	require.NoError(t, err)

	// Note: using %s for root_path to inject absolute path
	upstreamConfig := fmt.Sprintf(`
global_settings:
  # port omitted
upstream_services:
  - id: "backend-fs"
    name: "Backend FS"
    disable: false
    filesystem_service:
      root_paths:
        "/data": "%s"
      os: {}
`, upstreamConfigDir)
	err = os.WriteFile(filepath.Join(upstreamConfigDir, "config.yaml"), []byte(upstreamConfig), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(upstreamConfigDir, "backend_file.txt"), []byte("I am backend"), 0644)
	require.NoError(t, err)

	upstreamSP := StartServerProcess(t, binPath, "--config-path", filepath.Join(upstreamConfigDir, "config.yaml"), "--debug", "--api-key", "test-key")
	defer upstreamSP.Stop()

	// 2. Start Gateway Server (Frontend)
	gatewayConfigDir := filepath.Join(configDir, "gateway")
	err = os.MkdirAll(gatewayConfigDir, 0755)
	require.NoError(t, err)

	gatewayConfig := fmt.Sprintf(`
global_settings:
  # port omitted
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
        http_address: "%s/mcp/sse"
`, upstreamSP.BaseURL) // Inject upstream URL
	err = os.WriteFile(filepath.Join(gatewayConfigDir, "config.yaml"), []byte(gatewayConfig), 0644)
	require.NoError(t, err)

	gatewaySP := StartServerProcess(t, binPath, "--config-path", filepath.Join(gatewayConfigDir, "config.yaml"), "--debug", "--api-key", "test-key")
	defer gatewaySP.Stop()

	baseURL := gatewaySP.BaseURL

	// Wait for Gateway to be healthy (already checked in StartServerProcess, but good to be sure)
	verifyEndpoint(t, fmt.Sprintf("%s/healthz", baseURL), 200, 60*time.Second)

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
