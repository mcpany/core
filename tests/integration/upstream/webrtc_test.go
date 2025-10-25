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

package upstream

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mcpxy/core/pkg/util"
	"github.com/mcpxy/core/tests/framework"
	"github.com/mcpxy/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_Webrtc(t *testing.T) {
	testCase := &framework.E2ETestCase{
		Name:                "WebRTC Echo Server",
		UpstreamServiceType: "webrtc",
		BuildUpstream:       framework.BuildWebrtcServer,
		RegisterUpstream:    framework.RegisterWebrtcService,
		InvokeAIClient: func(t *testing.T, mcpxyEndpoint string) {
			ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
			defer cancel()

			testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
			cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpxyEndpoint}, nil)
			require.NoError(t, err)
			defer cs.Close()

			listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
			require.NoError(t, err)
			var foundTool bool
			serviceKey, _ := util.GenerateID("e2e_webrtc_echo")
			expectedToolName, _ := util.GenerateToolID(serviceKey, "echo")
			for _, tool := range listToolsResult.Tools {
				t.Logf("Discovered tool from MCPXY: %s", tool.Name)
				if tool.Name == expectedToolName {
					foundTool = true
				}
			}
			require.True(t, foundTool, "The webrtc echo tool was not discovered")

			echoMessage := `{"message": "hello world from webrtc"}`
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: expectedToolName, Arguments: json.RawMessage(echoMessage)})
			require.NoError(t, err, "Error calling echo tool")
			require.NotNil(t, res, "Nil response from echo tool")
			switch content := res.Content[0].(type) {
			case *mcp.TextContent:
				require.JSONEq(t, echoMessage, content.Text, "The echoed message does not match the original")
			default:
				t.Fatalf("Unexpected content type: %T", content)
			}
		},
	}

	framework.RunE2ETest(t, testCase)
}
