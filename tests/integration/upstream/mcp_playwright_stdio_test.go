
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
	testCase := &framework.E2ETestCase{
		Name:                "playwright server (Stdio)",
		UpstreamServiceType: "stdio",
		BuildUpstream:       func(t *testing.T) *integration.ManagedProcess { return nil },
		RegisterUpstream: func(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
			const serviceID = "playwright"
			integration.RegisterStdioMCPService(t, registrationClient, serviceID, "npx @playwright/mcp@latest", true)
		},
		InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
			ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
			defer cancel()

			testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
			cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpanyEndpoint}, nil)
			require.NoError(t, err)
			defer cs.Close()

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
