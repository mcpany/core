// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package public_api //nolint:revive

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCallPolicy_Enforcement(t *testing.T) {
	// We use the "chrome" config structure (command_line) but use @modelcontextprotocol/server-filesystem
	// This ensures we have a real MCP server that supports stdio.

	// Path to the filesystem server script
	fsServerPath, _ := filepath.Abs("../integration/upstream/node_modules/@modelcontextprotocol/server-filesystem/dist/index.js")

	// Case 1: Deny All
	// We configure a policy that DENIES everything by default.
	// We use the filesystem server, so it should expose tools like list_directory.
	configDenyAll := fmt.Sprintf(`
upstream_services:
  - id: "deny-service"
    name: "deny-service"
    mcp_service:
      stdio_connection:
        command: "node"
        args: ["%s", "."]
    call_policies:
      - default_action: DENY
        rules: []
    auto_discover_tool: true
`, fsServerPath)

	t.Run("DenyAll", func(t *testing.T) {
		serverInfo := integration.StartMCPANYServerWithConfig(t, "PolicyDenyAll", configDenyAll)
		defer serverInfo.CleanupFunc()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0"}, nil)
		cs, err := client.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: serverInfo.HTTPEndpoint}, nil)
		require.NoError(t, err)
		defer func() { _ = cs.Close() }()

		// List tools - should work, but EXECUTION is denied.
		// Wait for discovery
		var listDirTool string
		require.Eventually(t, func() bool {
			res, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
			if err != nil || len(res.Tools) == 0 {
				return false
			}
			for _, tool := range res.Tools {
				if tool.Name == "deny-service.list_directory" {
					listDirTool = tool.Name
					return true
				}
			}
			return false
		}, 10*time.Second, 100*time.Millisecond, "Timed out waiting for tools")

		// Try to call it - should be DENIED
		_, err = cs.CallTool(ctx, &mcp.CallToolParams{
			Name: listDirTool,
			Arguments: map[string]interface{}{
				"path": ".",
			},
		})
		assert.Error(t, err, "Should be denied by default policy")
		assert.Contains(t, err.Error(), "execution denied by policy")
	})

	// Case 2: Allow specific, deny specific
	// Default ALLOW, but DENY "read_file"
	configFs := fmt.Sprintf(`
upstream_services:
  - id: "fs-service"
    name: "fs-service"
    mcp_service:
      stdio_connection:
        command: "node"
        args: ["%s", "."]
    call_policies:
      - default_action: ALLOW
        rules:
          - action: DENY
            name_regex: "read_file"
    auto_discover_tool: true
`, fsServerPath)
	t.Run("Filesystem_Policy", func(t *testing.T) {
		serverInfo := integration.StartMCPANYServerWithConfig(t, "PolicyTestFs", configFs)
		defer serverInfo.CleanupFunc()

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0"}, nil)
		cs, err := client.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: serverInfo.HTTPEndpoint}, nil)
		require.NoError(t, err)
		defer func() { _ = cs.Close() }()

		// Wait for tools to be discovered
		var readFileTool, listDirTool string
		require.Eventually(t, func() bool {
			res, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
			if err != nil || len(res.Tools) == 0 {
				return false
			}
			for _, tool := range res.Tools {
				// Tool names are typically sanitized, usually "serviceId.toolName"
				if tool.Name == "fs-service.read_file" {
					readFileTool = tool.Name
				}
				if tool.Name == "fs-service.list_directory" {
					listDirTool = tool.Name
				}
			}
			return readFileTool != "" && listDirTool != ""
		}, 30*time.Second, 1*time.Second, "Timed out waiting for fs tools")

		t.Logf("Found tools: %s, %s", readFileTool, listDirTool)

		// 1. Call ALLOWED tool (list_directory)
		_, err = cs.CallTool(ctx, &mcp.CallToolParams{
			Name: listDirTool,
			Arguments: map[string]interface{}{
				"path": ".",
			},
		})
		assert.NoError(t, err, "list_directory should be allowed")

		// 2. Call DENIED tool (read_file)
		_, err = cs.CallTool(ctx, &mcp.CallToolParams{
			Name: readFileTool,
			Arguments: map[string]interface{}{
				"path": "go.mod", // valid file
			},
		})
		assert.Error(t, err, "read_file should be denied")
		assert.Contains(t, err.Error(), "denied by policy", "Error message should mention policy denial")
	})
}
