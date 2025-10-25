//go:build e2e

package examples

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mcpxy/core/tests/framework"
	"github.com/mcpxy/core/pkg/consts"
	"github.com/mcpxy/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommandExample(t *testing.T) {
	testCase := &framework.E2ETestCase{
		Name:                "Command Example",
		UpstreamServiceType: "command",
		BuildUpstream: func(t *testing.T) *integration.ManagedProcess {
			root, err := integration.GetProjectRoot()
			require.NoError(t, err)

			buildCmd := exec.Command("make", "build-e2e-mocks")
			buildCmd.Dir = root
			err = buildCmd.Run()
			require.NoError(t, err, "Failed to build command-tester binary")

			return &integration.ManagedProcess{}
		},
		GenerateUpstreamConfig: func(upstreamEndpoint string) string {
			root, err := integration.GetProjectRoot()
			require.NoError(t, err)

			commandTesterPath := filepath.Join(root, "build/test/bin/command-tester")
			configContent := fmt.Sprintf(`
upstream_services:
- name: command-service
  command_line_service:
    command: "%s"
    calls:
    - schema:
        name: "test-command"
        description: "A test command"
`, commandTesterPath)
			configPath := filepath.Join(t.TempDir(), "config.yaml")
			err = os.WriteFile(configPath, []byte(configContent), 0644)
			require.NoError(t, err)
			return configPath
		},
		InvokeAIClient: func(t *testing.T, mcpxyEndpoint string) {
			ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
			defer cancel()

			testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
			cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpxyEndpoint}, nil)
			require.NoError(t, err, "Failed to connect to MCPXY server")
			defer cs.Close()

			toolName := ""
			require.Eventually(t, func() bool {
				result, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
				if err != nil {
					t.Logf("Failed to list tools: %v", err)
					return false
				}
				for _, tool := range result.Tools {
					if strings.HasPrefix(tool.Name, "command-service") {
						toolName = tool.Name
						return true
					}
				}
				t.Logf("Tool not yet available")
				return false
			}, integration.TestWaitTimeMedium, 1*time.Second, "Tool did not become available in time")

			t.Run("success", func(t *testing.T) {
				params := json.RawMessage(`{"args": ["--stdout", "hello", "--exit-code", "0"]}`)
				res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: params})
				require.NoError(t, err, "Error calling tool '%s'", toolName)
				require.NotNil(t, res, "Nil response from tool '%s'", toolName)
				require.Len(t, res.Content, 1, "Expected exactly one content item from tool '%s'", toolName)
				textContent, ok := res.Content[0].(*mcp.TextContent)
				require.True(t, ok, "Expected content to be of type TextContent")

		var result map[string]interface{}
		err = json.Unmarshal([]byte(textContent.Text), &result)
		require.NoError(t, err, "Failed to unmarshal tool output")

		assert.Equal(t, "hello", result["stdout"])
		assert.Equal(t, "", result["stderr"])
				assert.Equal(t, consts.CommandStatusSuccess, result["status"])
		assert.Equal(t, float64(0), result["return_code"])
			})

			t.Run("error", func(t *testing.T) {
				params := json.RawMessage(`{"args": ["--stderr", "error", "--exit-code", "1"]}`)
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: params})
		require.NoError(t, err, "Error calling tool '%s'", toolName)
		require.NotNil(t, res, "Nil response from tool '%s'", toolName)
		require.Len(t, res.Content, 1, "Expected exactly one content item from tool '%s'", toolName)
		textContent, ok := res.Content[0].(*mcp.TextContent)
		require.True(t, ok, "Expected content to be of type TextContent")

		var result map[string]interface{}
		err = json.Unmarshal([]byte(textContent.Text), &result)
		require.NoError(t, err, "Failed to unmarshal tool output")

		assert.Equal(t, "", result["stdout"])
		assert.Equal(t, "error", result["stderr"])
				assert.Equal(t, consts.CommandStatusError, result["status"])
		assert.Equal(t, float64(1), result["return_code"])
			})

			t.Run("timeout", func(t *testing.T) {
		params := json.RawMessage(`{"args": ["--sleep", "2s"]}`)
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
				defer cancel()
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: params})
		require.NoError(t, err, "Error calling tool '%s'", toolName)
		require.NotNil(t, res, "Nil response from tool '%s'", toolName)
		require.Len(t, res.Content, 1, "Expected exactly one content item from tool '%s'", toolName)
		textContent, ok := res.Content[0].(*mcp.TextContent)
		require.True(t, ok, "Expected content to be of type TextContent")

		var result map[string]interface{}
		err = json.Unmarshal([]byte(textContent.Text), &result)
		require.NoError(t, err, "Failed to unmarshal tool output")

				assert.Equal(t, consts.CommandStatusTimeout, result["status"])
			})
		},
	}

	framework.RunE2ETest(t, testCase)
}
