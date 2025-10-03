/*
 * Copyright 2025 Author(s) of MCPX
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
	"net"
	"testing"
	"time"

	"github.com/mcpxy/mcpx/pkg/consts"
	"github.com/mcpxy/mcpx/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_MCP_StreamableHTTP(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeMedium)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for 'everything' server (Streamable HTTP)...")
	t.Parallel()

	// --- 1. Start 'everything' server directly and get the random port ---
	hostPort := integration.FindFreePort(t)
	serviceHttpEndpoint := fmt.Sprintf("http://localhost:%d/mcp", hostPort)
	serviceMcpEndpoint := serviceHttpEndpoint

	args := []string{"@modelcontextprotocol/server-everything", "streamableHttp"}
	env := []string{fmt.Sprintf("PORT=%d", hostPort)}
	everythingProc := integration.NewManagedProcess(t, "everything_streamable_server", "npx", args, env)
	everythingProc.IgnoreExitStatusOne = true
	err := everythingProc.Start()
	require.NoError(t, err, "Failed to start 'everything' server. Stderr: %s", everythingProc.StderrString())
	t.Cleanup(everythingProc.Stop)

	// Wait for the 'everything' server to be ready
	require.Eventually(t, func() bool {
		conn, err := net.DialTimeout("tcp", net.JoinHostPort("localhost", fmt.Sprintf("%d", hostPort)), 1*time.Second)
		if err != nil {
			return false
		}
		conn.Close()
		return true
	}, integration.ServiceStartupTimeout, integration.RetryInterval, "everything server did not become ready in time")

	// --- 2. Start MCPX Server ---
	mcpxTestServerInfo := integration.StartMCPXServer(t, "E2EEverythingServerTest_Streamable")
	defer mcpxTestServerInfo.CleanupFunc()

	// --- 4. Register 'everything' server with MCPX ---
	const everythingServiceIDStream = "e2e_everything_server_streamable"
	t.Logf("INFO: Registering '%s' with MCPX...", everythingServiceIDStream)
	registrationGRPCClient := mcpxTestServerInfo.RegistrationClient
	integration.RegisterStreamableMCPService(t, registrationGRPCClient, everythingServiceIDStream, serviceMcpEndpoint, true, nil)
	t.Logf("INFO: '%s' registered with URL: %s", everythingServiceIDStream, serviceMcpEndpoint)

	// Create MCP client to MCPX server
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
	t.Logf("Tools available from MCPX server: %v", listToolsResult.Tools)

	testCases := []struct {
		name       string
		tool       string
		jsonArgs   string
		expectFail bool
	}{
		{
			name:     "Tool_add",
			tool:     fmt.Sprintf("%s%sadd", everythingServiceIDStream, consts.ToolNameServiceSeparator),
			jsonArgs: `{"a": 10, "b": 15}`,
		},
		{
			name:     "Tool_echo",
			tool:     fmt.Sprintf("%s%secho", everythingServiceIDStream, consts.ToolNameServiceSeparator),
			jsonArgs: `{"message": "Hello, world!"}`,
		},
		{
			name:     "Tool_printEnv",
			tool:     fmt.Sprintf("%s%sprintEnv", everythingServiceIDStream, consts.ToolNameServiceSeparator),
			jsonArgs: `{}`,
		},
		{
			name:       "Tool_nonexistent",
			tool:       fmt.Sprintf("%s%snonexistent_tool", everythingServiceIDStream, consts.ToolNameServiceSeparator),
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
				t.Logf("SUCCESS: Call to tool '%s' via MCPX completed. Result: %s", tc.tool, res.Content[0])
			}
		})
	}

	t.Log("INFO: E2E Test Scenario for 'everything' server (Streamable HTTP) Completed Successfully!")
}
