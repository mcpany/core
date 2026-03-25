// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package public_api //nolint:revive

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockMCPServerBinary returns the path to the mock_mcp_server binary.
// Under Bazel it resolves from runfiles; otherwise it falls back to building from source.
func mockMCPServerBinary(t *testing.T) string {
	t.Helper()
	workspace := os.Getenv("TEST_WORKSPACE")
	if workspace == "" {
		workspace = "_main"
	}
	for _, base := range []string{os.Getenv("TEST_SRCDIR"), os.Getenv("RUNFILES_DIR")} {
		if base == "" {
			continue
		}
		for _, suffix := range []string{"mock_mcp_server_/mock_mcp_server", "mock_mcp_server"} {
			candidate := filepath.Join(base, workspace, "server", "cmd", "mock_mcp_server", suffix)
			if _, err := os.Stat(candidate); err == nil {
				return candidate
			}
		}
	}
	// Fall back to building from source
	root := integration.ProjectRoot(t)
	outBin := filepath.Join(t.TempDir(), "mock_mcp_server")
	cmd := exec.Command("go", "build", "-o", outBin, "./cmd/mock_mcp_server") //nolint:gosec
	cmd.Dir = root
	require.NoError(t, cmd.Run(), "Failed to build mock_mcp_server")
	return outBin
}

func TestCallPolicy_Enforcement(t *testing.T) {
	mockBin := mockMCPServerBinary(t)

	// Case 1: Deny All
	// We configure a policy that DENIES everything by default.
	// The mock MCP server exposes list_directory and read_file tools.
	configDenyAll := fmt.Sprintf(`
upstream_services:
  - id: "deny-service"
    name: "deny-service"
    mcp_service:
      stdio_connection:
        command: %q
    call_policies:
      - default_action: DENY
        rules: []
    auto_discover_tool: true
`, mockBin)

	t.Run("DenyAll", func(t *testing.T) {
		serverInfo := integration.StartMCPANYServerWithConfig(t, "PolicyDenyAll", configDenyAll)
		defer serverInfo.CleanupFunc()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
		}, 20*time.Second, 250*time.Millisecond, "Timed out waiting for tools")

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
        command: %q
    call_policies:
      - default_action: ALLOW
        rules:
          - action: DENY
            name_regex: "read_file"
    auto_discover_tool: true
`, mockBin)
	t.Run("MockMCP_Policy", func(t *testing.T) {
		serverInfo := integration.StartMCPANYServerWithConfig(t, "PolicyTestMock", configFs)
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
		}, 30*time.Second, 1*time.Second, "Timed out waiting for mock MCP tools")

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
				"path": "go.mod",
			},
		})
		assert.Error(t, err, "read_file should be denied")
		assert.Contains(t, err.Error(), "denied by policy", "Error message should mention policy denial")
	})
}
