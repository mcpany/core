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

func TestUpstreamService_OpenAPI_WithAPIKeyAuth(t *testing.T) {
	testCase := &framework.E2ETestCase{
		Name:                "Authenticated OpenAPI Echo Server",
		UpstreamServiceType: "openapi",
		BuildUpstream:       framework.BuildOpenAPIAuthedServer,
		RegisterUpstream:    framework.RegisterOpenAPIAuthedService,
		InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
			ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
			defer cancel()

			testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
			cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpanyEndpoint}, nil)
			require.NoError(t, err)
			defer cs.Close()

			serviceID, _ := util.SanitizeServiceName("e2e_openapi_authed_echo")
			sanitizedToolName, _ := util.SanitizeToolName("echo")
			toolName := serviceID + "." + sanitizedToolName
			echoMessage := `{"message": "hello world from authed openapi"}`
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(echoMessage)})
			require.NoError(t, err, "Error calling echo tool with correct auth")
			require.NotNil(t, res, "Nil response from echo tool with correct auth")
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
