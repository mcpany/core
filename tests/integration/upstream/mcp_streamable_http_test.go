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
	"github.com/mcpany/core/tests/framework"
	"github.com/mcpany/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_MCP_StreamableHTTP(t *testing.T) {
	testCase := &framework.E2ETestCase{
		Name:                "everything server (Streamable HTTP)",
		UpstreamServiceType: "streamablehttp",
		BuildUpstream:       framework.BuildStreamableHTTPServer,
		RegisterUpstream:    framework.RegisterStreamableHTTPService,
		InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
			ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeMedium)
			defer cancel()

			testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
			cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpanyEndpoint}, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = cs.Close() }()

			listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
			if err != nil {
				t.Fatalf("failed to list tools from MCP service: %v", err)
			}
			t.Logf("Tools available from MCPANY server: %v", listToolsResult.Tools)

			serviceID, _ := util.SanitizeServiceName("e2e_everything_server_streamable")

			testCases := []struct {
				name       string
				tool       string
				jsonArgs   string
				expectFail bool
			}{
				{
					name:     "Tool_add",
					tool:     "add",
					jsonArgs: `{"a": 10, "b": 15}`,
				},
				{
					name:     "Tool_echo",
					tool:     "echo",
					jsonArgs: `{"message": "Hello, world!"}`,
				},
				{
					name:     "Tool_printEnv",
					tool:     "printEnv",
					jsonArgs: `{}`,
				},
				{
					name:       "Tool_nonexistent",
					tool:       "nonexistent_tool",
					jsonArgs:   `{}`,
					expectFail: true,
				},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					sanitizedToolName, _ := util.SanitizeToolName(tc.tool)
					toolName := serviceID + "." + sanitizedToolName
					if tc.expectFail {
						_, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(tc.jsonArgs)})
						require.Error(t, err, "Expected error when calling nonexistent tool '%s', but got none", toolName)
						t.Logf("SUCCESS: Expected failure when calling nonexistent tool '%s': %v", toolName, err)
					} else {
						res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(tc.jsonArgs)})
						require.NoError(t, err, "Error calling tool '%s': %v", toolName, err)
						require.NotNil(t, res, "Nil response from tool '%s'", toolName)
						require.Len(t, res.Content, 1, "Expected exactly one content item from tool '%s'", toolName)
						t.Logf("SUCCESS: Call to tool '%s' via MCPANY completed. Result: %s", toolName, res.Content[0])
					}
				})
			}
		},
	}

	framework.RunE2ETest(t, testCase)
}
