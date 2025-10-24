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
	"github.com/stretchr/testify/require"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestCommandExample(t *testing.T) {
	testCase := &framework.E2ETestCase{
		Name:                "Command Example",
		UpstreamServiceType: "command",
		RegistrationMethods: []framework.RegistrationMethod{framework.FileRegistration, framework.GRPCRegistration},
		BuildUpstream: func(t *testing.T) *integration.ManagedProcess {
			// The command example doesn't run a separate upstream process
			return &integration.ManagedProcess{}
		},
		GenerateUpstreamConfig: func(upstreamEndpoint string) string {
			upstreamServiceConfig := &configv1.UpstreamService{
				Name: "hello-service",
				McpService: &configv1.McpUpstreamService{
					Connection: &configv1.McpUpstreamService_Stdio{
						Stdio: &configv1.McpStdioConnection{
							Command: "build/venv/bin/python",
							Args: []string{
								"-u",
								"examples/upstream/command/server/main.py",
								"--mcp-stdio",
							},
						},
					},
				},
			}
			config := &configv1.Config{
				Services: []*configv1.UpstreamService{upstreamServiceConfig},
			}

			jsonBytes, err := protojson.Marshal(config)
			require.NoError(t, err)
			return string(jsonBytes)
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
