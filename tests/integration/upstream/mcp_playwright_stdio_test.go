/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
		BuildUpstream:       func(_ *testing.T) *integration.ManagedProcess { return nil },
		RegisterUpstream: func(t *testing.T, registrationClient apiv1.RegistrationServiceClient, _ string) {
			const serviceID = "playwright"
			integration.RegisterStdioMCPService(t, registrationClient, serviceID, "npx @playwright/mcp@latest", true)
		},
		InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
			ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
			defer cancel()

			testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
			cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpanyEndpoint}, nil)
			require.NoError(t, err)
			defer func() { _ = cs.Close() }()

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
