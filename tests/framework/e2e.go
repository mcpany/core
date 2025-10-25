/*
 * Copyright 2025 Author(s) of MCP-XY
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

package framework

import (
	"context"
	"fmt"
	"testing"

	apiv1 "github.com/mcpxy/core/proto/api/v1"

	"github.com/mcpxy/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

type RegistrationMethod string

const (
	FileRegistration   RegistrationMethod = "file"
	GRPCRegistration   RegistrationMethod = "grpc"
	JSONRPCRegistration RegistrationMethod = "jsonrpc"
)

type E2ETestCase struct {
	Name                   string
	UpstreamServiceType    string
	BuildUpstream          func(t *testing.T) *integration.ManagedProcess
	RegisterUpstream       func(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string)
	ValidateTool           func(t *testing.T, mcpxyEndpoint string)
	InvokeAIClient         func(t *testing.T, mcpxyEndpoint string)
	RegistrationMethods    []RegistrationMethod
	GenerateUpstreamConfig func(upstreamEndpoint string) string
}

func ValidateRegisteredTool(t *testing.T, mcpxyEndpoint string, expectedTool *mcp.Tool) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	client := mcp.NewClient(&mcp.Implementation{Name: "e2e-test-client"}, nil)

	transport := &mcp.StreamableClientTransport{
		Endpoint: mcpxyEndpoint,
	}

	session, err := client.Connect(ctx, transport, nil)
	require.NoError(t, err)
	defer session.Close()

	tools, err := session.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)

	var foundTool *mcp.Tool
	for _, tool := range tools.Tools {
		if tool.Name == expectedTool.Name {
			foundTool = tool
			break
		}
	}

	require.NotNil(t, foundTool, "tool %q not found", expectedTool.Name)
	require.Equal(t, expectedTool.Description, foundTool.Description)
	require.Equal(t, expectedTool.InputSchema, foundTool.InputSchema)
}

func RunE2ETest(t *testing.T, testCase *E2ETestCase) {
	for _, method := range testCase.RegistrationMethods {
		t.Run(fmt.Sprintf("%s_%s", testCase.Name, method), func(t *testing.T) {
			_, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
			defer cancel()

			t.Logf("INFO: Starting E2E Test Scenario for %s with %s registration...", testCase.Name, method)
			t.Parallel()

			// --- 1. Start Upstream Service ---
			var upstreamEndpoint string
			if testCase.UpstreamServiceType != "command" {
				upstreamServerProc := testCase.BuildUpstream(t)
				err := upstreamServerProc.Start()
				require.NoError(t, err, "Failed to start upstream server")
				t.Cleanup(upstreamServerProc.Stop)
				integration.WaitForTCPPort(t, upstreamServerProc.Port, integration.ServiceStartupTimeout)
				upstreamEndpoint = fmt.Sprintf("localhost:%d", upstreamServerProc.Port)
			}

			var mcpxyTestServerInfo *integration.MCPXYTestServerInfo
			if method == FileRegistration {
				configContent := testCase.GenerateUpstreamConfig(upstreamEndpoint)
				mcpxyTestServerInfo = integration.StartMCPXYServerWithConfig(t, testCase.Name, configContent)
			} else {
				mcpxyTestServerInfo = integration.StartMCPXYServer(t, testCase.Name)
			}
			defer mcpxyTestServerInfo.CleanupFunc()

			// --- 3. Register Upstream Service with MCPXY ---
			t.Logf("INFO: Registering upstream service with MCPXY at endpoint %s...", upstreamEndpoint)
			if method == GRPCRegistration {
				testCase.RegisterUpstream(t, mcpxyTestServerInfo.RegistrationClient, upstreamEndpoint)
			}
			//TODO: JSONRPC registration
			t.Logf("INFO: Upstream service registered.")

			// --- 4. Validate Registered Tool ---
			if testCase.ValidateTool != nil {
				t.Logf("INFO: Validating registered tool...")
				testCase.ValidateTool(t, mcpxyTestServerInfo.HTTPEndpoint)
				t.Logf("INFO: Tool validated.")
			}

			// --- 5. Invoke AI Client ---
			testCase.InvokeAIClient(t, mcpxyTestServerInfo.HTTPEndpoint)

			t.Logf("INFO: E2E Test Scenario for %s Completed Successfully!", testCase.Name)
		})
	}
}
