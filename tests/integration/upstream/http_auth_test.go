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
	"net/http"
	"testing"

	"github.com/mcpxy/mcpx/pkg/consts"
	configv1 "github.com/mcpxy/mcpx/proto/config/v1"
	"github.com/mcpxy/mcpx/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestUpstreamService_HTTP_WithAPIKeyAuth(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Authenticated HTTP Echo Server...")
	t.Parallel()

	// --- 1. Start Authenticated HTTP Echo Server ---
	echoServerPort := integration.FindFreePort(t)
	echoServerProc := integration.NewManagedProcess(t, "http_authed_echo_server", "../../../build/test/bin/http_authed_echo_server", []string{fmt.Sprintf("--port=%d", echoServerPort)}, nil)
	err := echoServerProc.Start()
	require.NoError(t, err, "Failed to start authenticated HTTP Echo server")
	t.Cleanup(echoServerProc.Stop)
	require.Eventually(t, func() bool {
		return integration.IsTCPPortAvailable(echoServerPort)
	}, integration.ServiceStartupTimeout, integration.RetryInterval, "Authenticated HTTP Echo server did not become ready in time")

	// --- 2. Start MCPX Server ---
	mcpxTestServerInfo := integration.StartMCPXServer(t, "E2EHttpAuthedEchoServerTest")
	defer mcpxTestServerInfo.CleanupFunc()

	// --- 3. Register Authenticated HTTP Echo Server with MCPX ---
	const echoServiceID = "e2e_http_authed_echo"
	echoServiceEndpoint := fmt.Sprintf("http://localhost:%d", echoServerPort)
	t.Logf("INFO: Registering '%s' with MCPX at endpoint %s...", echoServiceID, echoServiceEndpoint)
	registrationGRPCClient := mcpxTestServerInfo.RegistrationClient

	authConfig := configv1.UpstreamAuthentication_builder{
		ApiKey: configv1.UpstreamAPIKeyAuth_builder{
			HeaderName: proto.String("X-Api-Key"),
			ApiKey:     proto.String("test-api-key"),
		}.Build(),
	}.Build()

	integration.RegisterHTTPService(t, registrationGRPCClient, echoServiceID, echoServiceEndpoint, "echo", "/echo", http.MethodPost, authConfig)
	t.Logf("INFO: '%s' registered.", echoServiceID)

	// --- 4. Call Tool via MCPX ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpxTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	toolName := fmt.Sprintf("%s%secho", echoServiceID, consts.ToolNameServiceSeparator)
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

	// --- 5. Test with incorrect auth ---
	t.Run("IncorrectAuth", func(t *testing.T) {
		const wrongAuthServiceID = "e2e_http_wrong_auth_echo"
		wrongAuthConfig := configv1.UpstreamAuthentication_builder{
			ApiKey: configv1.UpstreamAPIKeyAuth_builder{
				HeaderName: proto.String("X-Api-Key"),
				ApiKey:     proto.String("wrong-key"),
			}.Build(),
		}.Build()
		integration.RegisterHTTPService(t, registrationGRPCClient, wrongAuthServiceID, echoServiceEndpoint, "echo", "/echo", http.MethodPost, wrongAuthConfig)

		wrongToolName := fmt.Sprintf("%s%secho", wrongAuthServiceID, consts.ToolNameServiceSeparator)
		_, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: wrongToolName, Arguments: json.RawMessage(echoMessage)})
		require.Error(t, err, "Expected error when calling echo tool with incorrect auth")
		require.Contains(t, err.Error(), "Unauthorized", "Expected error message to contain 'Unauthorized'")
	})

	t.Log("INFO: E2E Test Scenario for Authenticated HTTP Echo Server Completed Successfully!")
}
