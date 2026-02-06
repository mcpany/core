// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package upstream

import (
	"context"
	"encoding/json"
	"testing"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/mcpany/core/server/tests/framework"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_MCP_Playwright_Stdio(t *testing.T) {
	// t.Skip("Skipping failing Playwright test: tool returns 0 tools in test env (investigated: stdout pollution fixed in docker_transport.go)")

	testCase := &framework.E2ETestCase{
		Name:                "playwright server (Stdio)",
		UpstreamServiceType: "stdio",
		BuildUpstream:       func(_ *testing.T) *integration.ManagedProcess { return nil },
		RegisterUpstream: func(t *testing.T, registrationClient apiv1.RegistrationServiceClient, _ string) {
			const serviceID = "playwright"
			env := map[string]string{
				"HOME":                             "/tmp",
				"NPM_CONFIG_LOGLEVEL":              "warn",
				"PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD": "1",
				"NPM_CONFIG_YES":                   "true",
			}
			cmd := "node"
			args := []string{"./node_modules/.bin/mcp-server-playwright", "--", "--console-level", "debug"}
			setupCommands := []string{
				"npm install --no-optional @playwright/mcp",
			}
			integration.RegisterStdioServiceWithSetup(t, registrationClient, serviceID, cmd, true, "/tmp", "mcr.microsoft.com/playwright:v1.58.0-jammy", setupCommands, env, args...)



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
