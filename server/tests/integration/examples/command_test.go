//go:build e2e

package examples

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/consts"
	"github.com/mcpany/core/server/tests/framework"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommandExample(t *testing.T) {
	testCase := &framework.E2ETestCase{
		Name:                "Command Example",
		UpstreamServiceType: "command",
		BuildUpstream: func(t *testing.T) *integration.ManagedProcess {
			return &integration.ManagedProcess{}
		},
		GenerateUpstreamConfig: func(upstreamEndpoint string) string {
			configContent := `
upstream_services:
- name: datetime-service
  command_line_service:
    command: "date"
    calls:
    - schema:
        name: "get_current_date"
        description: "Get the current date"
    - schema:
        name: "get_current_date_iso"
        description: "Get the current date in ISO format"
      args:
        - "-I"
`
			configPath := filepath.Join(t.TempDir(), "config.yaml")
			err := os.WriteFile(configPath, []byte(configContent), 0o644)
			require.NoError(t, err)
			return configPath
		},
		InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
			ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
			defer cancel()

			testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
			cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpanyEndpoint}, nil)
			require.NoError(t, err, "Failed to connect to MCPANY server")
			defer cs.Close()

			var dateToolName, dateIsoToolName string
			require.Eventually(t, func() bool {
				result, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
				if err != nil {
					t.Logf("Failed to list tools: %v", err)
					return false
				}
				for _, tool := range result.Tools {
					if strings.HasSuffix(tool.Name, "get_current_date") {
						dateToolName = tool.Name
					}
					if strings.HasSuffix(tool.Name, "get_current_date_iso") {
						dateIsoToolName = tool.Name
					}
				}
				return dateToolName != "" && dateIsoToolName != ""
			}, integration.TestWaitTimeMedium, 1*time.Second, "Tools did not become available in time")

			t.Run("get_current_date", func(t *testing.T) {
				res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: dateToolName})
				require.NoError(t, err, "Error calling tool '%s'", dateToolName)
				require.NotNil(t, res, "Nil response from tool '%s'", dateToolName)
				require.Len(t, res.Content, 1, "Expected exactly one content item from tool '%s'", dateToolName)
				textContent, ok := res.Content[0].(*mcp.TextContent)
				require.True(t, ok, "Expected content to be of type TextContent")

				var result map[string]interface{}
				err = json.Unmarshal([]byte(textContent.Text), &result)
				require.NoError(t, err, "Failed to unmarshal tool output")

				assert.NotEmpty(t, result["stdout"])
				assert.Equal(t, consts.CommandStatusSuccess, result["status"])
				assert.Equal(t, 0, result["return_code"])
			})

			t.Run("get_current_date_iso", func(t *testing.T) {
				res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: dateIsoToolName})
				require.NoError(t, err, "Error calling tool '%s'", dateIsoToolName)
				require.NotNil(t, res, "Nil response from tool '%s'", dateIsoToolName)
				require.Len(t, res.Content, 1, "Expected exactly one content item from tool '%s'", dateIsoToolName)
				textContent, ok := res.Content[0].(*mcp.TextContent)
				require.True(t, ok, "Expected content to be of type TextContent")

				var result map[string]interface{}
				err = json.Unmarshal([]byte(textContent.Text), &result)
				require.NoError(t, err, "Failed to unmarshal tool output")

				assert.NotEmpty(t, result["stdout"])
				_, err = time.Parse("2006-01-02\n", result["stdout"].(string))
				assert.NoError(t, err, "stdout is not in the expected format")
				assert.Equal(t, consts.CommandStatusSuccess, result["status"])
				assert.Equal(t, 0, result["return_code"])
			})
		},
	}

	framework.RunE2ETest(t, testCase)
}

func TestCommandExampleWithTimeout(t *testing.T) {
	testCase := &framework.E2ETestCase{
		Name:                "Command Example with Timeout",
		UpstreamServiceType: "command",
		BuildUpstream: func(t *testing.T) *integration.ManagedProcess {
			return &integration.ManagedProcess{}
		},
		GenerateUpstreamConfig: func(upstreamEndpoint string) string {
			configContent := `
upstream_services:
- name: sleep-service
  command_line_service:
    command: "sleep"
    timeout: 1s
    calls:
    - schema:
        name: "sleep"
        description: "Sleep for a given duration"
      args:
        - "2"
`
			configPath := filepath.Join(t.TempDir(), "config.yaml")
			err := os.WriteFile(configPath, []byte(configContent), 0o644)
			require.NoError(t, err)
			return configPath
		},
		InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
			ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
			defer cancel()

			testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
			cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpanyEndpoint}, nil)
			require.NoError(t, err, "Failed to connect to MCPANY server")
			defer cs.Close()

			var sleepToolName string
			require.Eventually(t, func() bool {
				result, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
				if err != nil {
					t.Logf("Failed to list tools: %v", err)
					return false
				}
				for _, tool := range result.Tools {
					if strings.HasSuffix(tool.Name, "sleep") {
						sleepToolName = tool.Name
					}
				}
				return sleepToolName != ""
			}, integration.TestWaitTimeMedium, 1*time.Second, "Tool did not become available in time")

			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: sleepToolName})
			require.NoError(t, err, "Error calling tool '%s'", sleepToolName)
			require.NotNil(t, res, "Nil response from tool '%s'", sleepToolName)
			require.Len(t, res.Content, 1, "Expected exactly one content item from tool '%s'", sleepToolName)
			textContent, ok := res.Content[0].(*mcp.TextContent)
			require.True(t, ok, "Expected content to be of type TextContent")

			var result map[string]interface{}
			err = json.Unmarshal([]byte(textContent.Text), &result)
			require.NoError(t, err, "Failed to unmarshal tool output")

			assert.Equal(t, consts.CommandStatusTimeout, result["status"])
			assert.Equal(t, -1, result["return_code"])
		},
	}

	framework.RunE2ETest(t, testCase)
}

func TestCommandExampleWithContainer(t *testing.T) {
	testCase := &framework.E2ETestCase{
		Name:                "Command Example with Container",
		UpstreamServiceType: "command",
		BuildUpstream: func(t *testing.T) *integration.ManagedProcess {
			return &integration.ManagedProcess{}
		},
		GenerateUpstreamConfig: func(upstreamEndpoint string) string {
			configContent := `
upstream_services:
- name: datetime-service
  command_line_service:
    command: "date"
    container_environment:
      image: "alpine:latest"
    calls:
    - schema:
        name: "get_current_date"
        description: "Get the current date"
`
			configPath := filepath.Join(t.TempDir(), "config.yaml")
			err := os.WriteFile(configPath, []byte(configContent), 0o644)
			require.NoError(t, err)
			return configPath
		},
		InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
			ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
			defer cancel()

			testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
			cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpanyEndpoint}, nil)
			require.NoError(t, err, "Failed to connect to MCPANY server")
			defer cs.Close()

			var dateToolName string
			require.Eventually(t, func() bool {
				result, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
				if err != nil {
					t.Logf("Failed to list tools: %v", err)
					return false
				}
				for _, tool := range result.Tools {
					if strings.HasSuffix(tool.Name, "get_current_date") {
						dateToolName = tool.Name
					}
				}
				return dateToolName != ""
			}, integration.TestWaitTimeMedium, 1*time.Second, "Tool did not become available in time")

			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: dateToolName})
			require.NoError(t, err, "Error calling tool '%s'", dateToolName)
			require.NotNil(t, res, "Nil response from tool '%s'", dateToolName)
			require.Len(t, res.Content, 1, "Expected exactly one content item from tool '%s'", dateToolName)
			textContent, ok := res.Content[0].(*mcp.TextContent)
			require.True(t, ok, "Expected content to be of type TextContent")

			var result map[string]interface{}
			err = json.Unmarshal([]byte(textContent.Text), &result)
			require.NoError(t, err, "Failed to unmarshal tool output")

			assert.NotEmpty(t, result["stdout"])
			assert.Equal(t, consts.CommandStatusSuccess, result["status"])
			assert.Equal(t, 0, result["return_code"])
		},
	}

	framework.RunE2ETest(t, testCase)
}
