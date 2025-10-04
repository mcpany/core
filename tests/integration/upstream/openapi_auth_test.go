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
	"os"
	"testing"

	"github.com/mcpxy/core/pkg/consts"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	"github.com/mcpxy/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestUpstreamService_OpenAPI_WithAPIKeyAuth(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Authenticated OpenAPI Echo Server...")
	t.Parallel()

	// --- 1. Start Authenticated HTTP Echo Server ---
	echoServerPort := integration.FindFreePort(t)
	echoServerProc := integration.NewManagedProcess(t, "http_authed_echo_server_openapi", "../../../build/test/bin/http_authed_echo_server", []string{fmt.Sprintf("--port=%d", echoServerPort)}, nil)
	err := echoServerProc.Start()
	require.NoError(t, err, "Failed to start authenticated HTTP Echo server for OpenAPI test")
	t.Cleanup(echoServerProc.Stop)
	require.Eventually(t, func() bool {
		return integration.IsTCPPortAvailable(echoServerPort)
	}, integration.ServiceStartupTimeout, integration.RetryInterval, "Authenticated HTTP Echo server for OpenAPI test did not become ready in time")

	// --- 2. Start MCPXY Server ---
	mcpxTestServerInfo := integration.StartMCPXYServer(t, "E2EOpenAPIAuthedEchoServerTest")
	defer mcpxTestServerInfo.CleanupFunc()

	// --- 3. Register Authenticated OpenAPI Echo Server with MCPXY ---
	const echoServiceID = "e2e_openapi_authed_echo"
	echoServiceEndpoint := fmt.Sprintf("http://localhost:%d", echoServerPort)
	t.Logf("INFO: Registering '%s' with MCPXY at endpoint %s...", echoServiceID, echoServiceEndpoint)
	registrationGRPCClient := mcpxTestServerInfo.RegistrationClient

	openapiSpec := fmt.Sprintf(`
openapi: 3.0.0
info:
  title: Authenticated Echo Service
  version: 1.0.0
servers:
  - url: %s
paths:
  /echo:
    post:
      operationId: echo
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                message:
                  type: string
      responses:
        '200':
          description: OK
`, echoServiceEndpoint)

	tmpfile, err := os.CreateTemp("", "openapi-auth-*.json")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())
	_, err = tmpfile.WriteString(openapiSpec)
	require.NoError(t, err)
	err = tmpfile.Close()
	require.NoError(t, err)

	authConfig := configv1.UpstreamAuthentication_builder{
		ApiKey: configv1.UpstreamAPIKeyAuth_builder{
			HeaderName: proto.String("X-Api-Key"),
			ApiKey:     proto.String("test-api-key"),
		}.Build(),
	}.Build()

	integration.RegisterOpenAPIService(t, registrationGRPCClient, echoServiceID, tmpfile.Name(), echoServiceEndpoint, authConfig)
	t.Logf("INFO: '%s' registered.", echoServiceID)

	// --- 4. Call Tool via MCPXY ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpxTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	toolName := fmt.Sprintf("%s%secho", echoServiceID, consts.ToolNameServiceSeparator)
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

	t.Log("INFO: E2E Test Scenario for Authenticated OpenAPI Echo Server Completed Successfully!")
}
