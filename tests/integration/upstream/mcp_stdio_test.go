/*
 * Copyright 2025 Author(s) of MCPXY
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
	"fmt"
	"testing"

	"github.com/mcpxy/core/pkg/consts"
	"github.com/mcpxy/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_MCP_Stdio(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for 'everything' server (Stdio)...")
	t.Parallel()

	// --- 1. Start MCPXY Server ---
	mcpxTestServerInfo := integration.StartMCPXYServer(t, "E2EEverythingServerTest_Stdio")
	defer mcpxTestServerInfo.CleanupFunc()

	// --- 2. Register 'everything' server with MCPXY ---
	const everythingServiceIDStdio = "e2e_everything_server_stdio"
	serviceStdioEndpoint := "npx @modelcontextprotocol/server-everything stdio"
	t.Logf("INFO: Registering '%s' with MCPXY...", everythingServiceIDStdio)
	registrationGRPCClient := mcpxTestServerInfo.RegistrationClient
	integration.RegisterStdioMCPService(t, registrationGRPCClient, everythingServiceIDStdio, serviceStdioEndpoint, true)
	t.Logf("INFO: '%s' registered with command: %s", everythingServiceIDStdio, serviceStdioEndpoint)

	// Create MCP client to MCPXY server
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpxTestServerInfo.HTTPEndpoint}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cs.Close()

	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		t.Fatalf("failed to list tools from MCP service: %v", err)
	}
	t.Logf("Tools available from MCPXY server: %v", listToolsResult.Tools)

	testCases := []struct {
		name       string
		tool       string
		jsonArgs   string
		expectFail bool
	}{
		{
			name:     "Tool_add",
			tool:     fmt.Sprintf("%s%sadd", everythingServiceIDStdio, consts.ToolNameServiceSeparator),
			jsonArgs: `{"a": 10, "b": 15}`,
		},
		{
			name:     "Tool_echo",
			tool:     fmt.Sprintf("%s%secho", everythingServiceIDStdio, consts.ToolNameServiceSeparator),
			jsonArgs: `{"message": "Hello, world!"}`,
		},
		{
			name:     "Tool_printEnv",
			tool:     fmt.Sprintf("%s%sprintEnv", everythingServiceIDStdio, consts.ToolNameServiceSeparator),
			jsonArgs: `{}`,
		},
		{
			name:       "Tool_nonexistent",
			tool:       fmt.Sprintf("%s%snonexistent_tool", everythingServiceIDStdio, consts.ToolNameServiceSeparator),
			jsonArgs:   `{}`,
			expectFail: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectFail {
				_, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: tc.tool, Arguments: json.RawMessage(tc.jsonArgs)})
				require.Error(t, err, "Expected error when calling nonexistent tool '%s', but got none", tc.tool)
				t.Logf("SUCCESS: Expected failure when calling nonexistent tool '%s': %v", tc.tool, err)
			} else {
				res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: tc.tool, Arguments: json.RawMessage(tc.jsonArgs)})
				require.NoError(t, err, "Error calling tool '%s': %v", tc.tool, err)
				require.NotNil(t, res, "Nil response from tool '%s'", tc.tool)
				require.Len(t, res.Content, 1, "Expected exactly one content item from tool '%s'", tc.tool)
				t.Logf("SUCCESS: Call to tool '%s' via MCPXY completed. Result: %s", tc.tool, res.Content[0])
			}
		})
	}

	t.Log("INFO: E2E Test Scenario for 'everything' server (Stdio) Completed Successfully!")
}
