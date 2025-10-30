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
	"net/http"
	"testing"

	"github.com/mcpxy/core/pkg/util"
	apiv1 "github.com/mcpxy/core/proto/api/v1"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	"github.com/mcpxy/core/tests/framework"
	"github.com/mcpxy/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestUpstreamService_HTTP_WithAPIKeyAuth(t *testing.T) {
	testCase := &framework.E2ETestCase{
		Name:                "Authenticated HTTP Echo Server",
		UpstreamServiceType: "http",
		BuildUpstream:       framework.BuildHTTPAuthedEchoServer,
		RegisterUpstream:    framework.RegisterHTTPAuthedEchoService,
		InvokeAIClient: func(t *testing.T, mcpxyEndpoint string) {
			ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
			defer cancel()

			testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
			cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpxyEndpoint}, nil)
			require.NoError(t, err)
			defer cs.Close()

			const echoServiceID = "e2e_http_authed_echo"
			serviceKey, _ := util.SanitizeServiceName(echoServiceID)
			toolName, _ := util.SanitizeToolName(serviceKey, "echo")
			echoMessage := `{"message": "hello world from authed http"}`
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

func TestUpstreamService_HTTP_WithIncorrectAPIKeyAuth(t *testing.T) {
	testCase := &framework.E2ETestCase{
		Name:                "Authenticated HTTP Echo Server with Incorrect API Key",
		UpstreamServiceType: "http",
		BuildUpstream:       framework.BuildHTTPAuthedEchoServer,
		RegisterUpstream: func(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
			const wrongAuthServiceID = "e2e_http_wrong_auth_echo"
			wrongAuthConfig := configv1.UpstreamAuthentication_builder{
				ApiKey: configv1.UpstreamAPIKeyAuth_builder{
					HeaderName: proto.String("X-Api-Key"),
					ApiKey:     proto.String("wrong-key"),
				}.Build(),
			}.Build()
			integration.RegisterHTTPService(t, registrationClient, wrongAuthServiceID, upstreamEndpoint, "echo", "/echo", http.MethodPost, wrongAuthConfig)
		},
		InvokeAIClient: func(t *testing.T, mcpxyEndpoint string) {
			ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
			defer cancel()

			testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
			cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpxyEndpoint}, nil)
			require.NoError(t, err)
			defer cs.Close()

			const wrongAuthServiceID = "e2e_http_wrong_auth_echo"
			wrongServiceKey, _ := util.SanitizeServiceName(wrongAuthServiceID)
			wrongToolName, _ := util.SanitizeToolName(wrongServiceKey, "echo")
			echoMessage := `{"message": "hello world from authed http"}`
			_, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: wrongToolName, Arguments: json.RawMessage(echoMessage)})
			require.Error(t, err, "Expected error when calling echo tool with incorrect auth")
			require.Contains(t, err.Error(), "Unauthorized", "Expected error message to contain 'Unauthorized'")
		},
	}

	framework.RunE2ETest(t, testCase)
}
