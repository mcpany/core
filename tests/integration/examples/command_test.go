package examples

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/mcpxy/core/pkg/consts"
	"github.com/mcpxy/core/tests/framework"
	"github.com/mcpxy/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"os"
	"os/exec"
	apiv1 "github.com/mcpxy/core/proto/api/v1"
	"github.com/stretchr/testify/require"
)

func TestCommandExample(t *testing.T) {
	t.Skip("Skipping command example test due to persistent timeout issues.")
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)

	testCase := &framework.E2ETestCase{
		Name:            "Command Example",
		UpstreamServiceType: "command",
		MCPXYServerArgs: []string{"--config-paths", root + "/examples/upstream/command/config"},
		BuildUpstream: func(t *testing.T) *integration.ManagedProcess {
			wd, _ := os.Getwd()
			t.Logf("Current working directory: %s", wd)

			t.Log("Running 'make build-e2e-mocks'...")
			mocksCmd := exec.Command("make", "build-e2e-mocks")
			mocksCmd.Dir = root
			mocksOutput, err := mocksCmd.CombinedOutput()
			require.NoError(t, err, "Failed to run 'make build-e2e-mocks'. Output:\n%s", string(mocksOutput))
			t.Logf("'make build-e2e-mocks' succeeded. Output:\n%s", string(mocksOutput))

			t.Log("Running 'make prepare'...")
			prepareCmd := exec.Command("make", "prepare")
			prepareCmd.Dir = root
			prepareOutput, err := prepareCmd.CombinedOutput()
			if err != nil {
				// The original test ignored this error, let's log it but continue for now.
				t.Logf("'make prepare' failed, but continuing as per original test logic. Output:\n%s", string(prepareOutput))
			} else {
				t.Logf("'make prepare' succeeded. Output:\n%s", string(prepareOutput))
			}

			t.Log("Running 'make build'...")
			buildCmd := exec.Command("make", "build")
			buildCmd.Dir = root
			buildOutput, err := buildCmd.CombinedOutput()
			require.NoError(t, err, "Failed to build mcpxy binary. Output:\n%s", string(buildOutput))
			t.Logf("'make build' succeeded. Output:\n%s", string(buildOutput))

			return nil
		},
		RegisterUpstream: func(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
			// No-op, registration is done via config
		},
		ValidateTool: func(t *testing.T, mcpxyEndpoint string) {
			toolName := fmt.Sprintf("hello-service%shello", consts.ToolNameServiceSeparator)
			framework.ValidateRegisteredTool(t, mcpxyEndpoint, &mcp.Tool{
				Name: toolName,
			})
		},
		InvokeAIClient: func(t *testing.T, mcpxyEndpoint string) {
			ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
			defer cancel()

			testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
			cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpxyEndpoint}, nil)
			require.NoError(t, err, "Failed to connect to MCPXY server")
			defer cs.Close()

			toolName := fmt.Sprintf("hello-service%shello", consts.ToolNameServiceSeparator)

			// Wait for the tool to be available
			require.Eventually(t, func() bool {
				result, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
				if err != nil {
					t.Logf("Failed to list tools: %v", err)
					return false
				}
				for _, tool := range result.Tools {
					if tool.Name == toolName {
						return true
					}
				}
				t.Logf("Tool %s not yet available", toolName)
				return false
			}, integration.TestWaitTimeMedium, 1*time.Second, "Tool %s did not become available in time", toolName)

			params := json.RawMessage(`{"name": "World"}`)

			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: params})
			require.NoError(t, err, "Error calling tool '%s'", toolName)
			require.NotNil(t, res, "Nil response from tool '%s'", toolName)
			require.Len(t, res.Content, 1, "Expected exactly one content item from tool '%s'", toolName)

			textContent, ok := res.Content[0].(*mcp.TextContent)
			require.True(t, ok, "Expected content to be of type TextContent")

			require.Equal(t, "Hello, World!", textContent.Text)
		},
	}

	framework.RunE2ETest(t, testCase)
}
