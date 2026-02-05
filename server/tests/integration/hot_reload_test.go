// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package integration_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/mcpany/core/server/tests/integration"
	"github.com/stretchr/testify/require"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestHotReload(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// 1. Initial Config
	initialConfig := `
global_settings:
  mcp_listen_address: "127.0.0.1:0"
  log_level: "debug"
upstream_services:
  - name: "service-a"
    http_service:
      address: "http://example.com"
      tools:
        - name: "tool-a"
          description: "Tool A"
          call_id: "call-a"
      calls:
        call-a:
          method: HTTP_METHOD_GET
          endpoint_path: "/a"
`
	serverInfo := integration.StartMCPANYServerWithConfig(t, "HotReloadTest", initialConfig)
	defer serverInfo.CleanupFunc()

	// Wait for start
	time.Sleep(5 * time.Second)

	// Helper to check tool existence
	hasTool := func(toolName string) bool {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		client := mcp.NewClient(&mcp.Implementation{Name: "hot-reload-client", Version: "1.0.0"}, nil)
		transport := &mcp.StreamableClientTransport{Endpoint: serverInfo.HTTPEndpoint}
		session, err := client.Connect(ctx, transport, nil)
		if err != nil {
			return false
		}
		defer func() { _ = session.Close() }()

		res, err := session.ListTools(ctx, &mcp.ListToolsParams{})
		if err != nil {
			return false
		}
		for _, tool := range res.Tools {
			// Name is serviceName.toolName usually, or just toolName depending on merge strategy
			// Default is serviceName.toolName
			if tool.Name == "service-a."+toolName || tool.Name == "service-b."+toolName {
				return true
			}
		}
		return false
	}

	require.True(t, hasTool("tool-a"), "Tool A should exist initially")
	require.False(t, hasTool("tool-b"), "Tool B should not exist initially")

	// 2. Find Config Path
	var configPath string
	for i, arg := range serverInfo.Process.Cmd().Args {
		if arg == "--config-path" && i+1 < len(serverInfo.Process.Cmd().Args) {
			configPath = serverInfo.Process.Cmd().Args[i+1]
			break
		}
	}
	require.NotEmpty(t, configPath, "Config path not found in args")

	// 3. Update Config (Add Service B)
	updatedConfig := `
global_settings:
  mcp_listen_address: "127.0.0.1:0"
  log_level: "debug"
upstream_services:
  - name: "service-a"
    http_service:
      address: "http://example.com"
      tools:
        - name: "tool-a"
          description: "Tool A"
          call_id: "call-a"
      calls:
        call-a:
          method: HTTP_METHOD_GET
          endpoint_path: "/a"
  - name: "service-b"
    http_service:
      address: "http://example.com"
      tools:
        - name: "tool-b"
          description: "Tool B"
          call_id: "call-b"
      calls:
        call-b:
          method: HTTP_METHOD_GET
          endpoint_path: "/b"
`
	err := os.WriteFile(configPath, []byte(updatedConfig), 0644)
	require.NoError(t, err)

	// 4. Wait for Reload (Debounce is likely few seconds)
	// server/docs/features/hot_reload.md says "The server debounces the events".
	require.Eventually(t, func() bool {
		return hasTool("tool-b")
	}, 15*time.Second, 1*time.Second, "Tool B should appear after hot reload")

	require.True(t, hasTool("tool-a"), "Tool A should still exist")
}
