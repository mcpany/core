package upstream

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mcpany/core/pkg/util"
	apiv1 "github.com/mcpany/core/proto/api/v1"
	"github.com/mcpany/core/tests/framework"
	"github.com/mcpany/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_MCP_Playwright_Stdio(t *testing.T) {
	t.Skip("Skipping failing Playwright test: server hangs on startup with npm exec")

	testCase := &framework.E2ETestCase{
		Name:                "playwright server (Stdio)",
		UpstreamServiceType: "stdio",
		BuildUpstream:       func(_ *testing.T) *integration.ManagedProcess { return nil },
		RegisterUpstream: func(t *testing.T, registrationClient apiv1.RegistrationServiceClient, _ string) {
			const serviceID = "playwright"
			env := map[string]string{
				"HOME":                             "/tmp",
				"NPM_CONFIG_LOGLEVEL":              "silent",
				"PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD": "1",
				"NPM_CONFIG_YES":                   "true",
			}
			cmd := "npm" // Use npm to find the binary
			args := []string{"exec", "mcp-server-playwright", "--", "--console-level", "debug"}
			setupCommands := []string{
				"timeout -s 9 20s npm install @playwright/mcp", // Install local package
			}
			integration.RegisterStdioServiceWithSetup(t, registrationClient, serviceID, cmd, true, "/tmp", "mcr.microsoft.com/playwright:v1.50.0-jammy", setupCommands, env, args...)



		},
		InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
			ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
			defer cancel()

			testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
			cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpanyEndpoint}, nil)
			require.NoError(t, err)
			defer func() { _ = cs.Close() }()

			// Verify tools are listed (confirms connection and registration)
			listToolsRes, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
			require.NoError(t, err)
			require.NotEmpty(t, listToolsRes.Tools)

			foundNavigate := false
			for _, tool := range listToolsRes.Tools {
				if tool.Name == "playwright.browser_navigate" {
					foundNavigate = true
					break
				}
			}
			require.True(t, foundNavigate, "Expected to find playwright.browser_navigate tool")

			t.Logf("Playwright service registered and tools listed successfully.")

			serviceID, _ := util.SanitizeServiceName("playwright")
			toolName, _ := util.SanitizeToolName("browser_navigate")
			fullToolName := serviceID + "." + toolName

			args := json.RawMessage(`{"url": "https://www.google.com"}`)

			_, err = cs.CallTool(ctx, &mcp.CallToolParams{
				Name:      fullToolName,
				Arguments: args,
			})
			require.NoError(t, err)
		},
		RegistrationMethods: []framework.RegistrationMethod{framework.GRPCRegistration},
	}

	framework.RunE2ETest(t, testCase)
}
