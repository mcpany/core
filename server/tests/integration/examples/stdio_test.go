//go:build e2e

package examples

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/mcpany/core/server/tests/framework"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStdioExample(t *testing.T) {
	testCase := &framework.E2ETestCase{
		Name:                "Stdio Example",
		UpstreamServiceType: "stdio",
		BuildUpstream: func(t *testing.T) *integration.ManagedProcess {
			return &integration.ManagedProcess{}
		},
		GenerateUpstreamConfig: func(upstreamEndpoint string) string {
			// The config path is relative to the root of the repo
			return "../../examples/demo/stdio/config.yaml"
		},
		InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
			ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
			defer cancel()

			testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
			cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpanyEndpoint}, nil)
			require.NoError(t, err, "Failed to connect to MCPANY server")
			defer cs.Close()

			var greetToolName string
			require.Eventually(t, func() bool {
				result, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
				if err != nil {
					t.Logf("Failed to list tools: %v", err)
					return false
				}
				for _, tool := range result.Tools {
					if strings.HasSuffix(tool.Name, "greet") {
						greetToolName = tool.Name
					}
				}
				return greetToolName != ""
			}, integration.TestWaitTimeMedium, 1*time.Second, "Tools did not become available in time")

			t.Run("greet", func(t *testing.T) {
				res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: greetToolName, Arguments: "{\"name\": \"world\"}"})
				require.NoError(t, err, "Error calling tool '%s'", greetToolName)
				require.NotNil(t, res, "Nil response from tool '%s'", greetToolName)
				require.Len(t, res.Content, 1, "Expected exactly one content item from tool '%s'", greetToolName)
				textContent, ok := res.Content[0].(*mcp.TextContent)
				require.True(t, ok, "Expected content to be of type TextContent")

				var result map[string]interface{}
				err = json.Unmarshal([]byte(textContent.Text), &result)
				require.NoError(t, err, "Failed to unmarshal tool output")

				assert.Equal(t, "Hello, world!", result["message"])
			})
		},
	}

	framework.RunE2ETest(t, testCase)
}
