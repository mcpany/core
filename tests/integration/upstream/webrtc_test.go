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

func TestUpstreamService_Webrtc(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for WebRTC Echo Server...")
	t.Parallel()

	// --- 1. Start WebRTC Echo Server ---
	echoServerPort := integration.FindFreePort(t)
	echoServerProc := integration.NewManagedProcess(t, "webrtc_echo_server", "../../../build/test/bin/webrtc_echo_server", []string{fmt.Sprintf("--port=%d", echoServerPort)}, nil)
	err := echoServerProc.Start()
	require.NoError(t, err, "Failed to start WebRTC Echo server")
	t.Cleanup(echoServerProc.Stop)
	require.Eventually(t, func() bool {
		return integration.IsTCPPortAvailable(echoServerPort)
	}, integration.ServiceStartupTimeout, integration.RetryInterval, "WebRTC Echo server did not become ready in time")

	// --- 2. Start MCPXY Server ---
	mcpxTestServerInfo := integration.StartMCPXYServer(t, "E2EWebrtcEchoServerTest")
	defer mcpxTestServerInfo.CleanupFunc()

	// --- 3. Register Webrtc Echo Server with MCPXY ---
	const echoServiceID = "e2e_webrtc_echo"
	echoServiceEndpoint := fmt.Sprintf("http://localhost:%d/signal", echoServerPort)
	t.Logf("INFO: Registering '%s' with MCPXY at endpoint %s...", echoServiceID, echoServiceEndpoint)
	registrationGRPCClient := mcpxTestServerInfo.RegistrationClient
	integration.RegisterWebrtcService(t, registrationGRPCClient, echoServiceID, echoServiceEndpoint, "echo", nil)
	t.Logf("INFO: '%s' registered.", echoServiceID)

	// --- 4. Call Tool via MCPXY ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpxTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)
	var foundTool bool
	for _, tool := range listToolsResult.Tools {
		t.Logf("Discovered tool from MCPXY: %s", tool.Name)
		if tool.Name == fmt.Sprintf("%s%secho", echoServiceID, consts.ToolNameServiceSeparator) {
			foundTool = true
		}
	}
	require.True(t, foundTool, "The webrtc echo tool was not discovered")

	toolName := fmt.Sprintf("%s%secho", echoServiceID, consts.ToolNameServiceSeparator)
	echoMessage := `{"message": "hello world from webrtc"}`
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(echoMessage)})
	require.NoError(t, err, "Error calling echo tool")
	require.NotNil(t, res, "Nil response from echo tool")
	switch content := res.Content[0].(type) {
	case *mcp.TextContent:
		require.JSONEq(t, echoMessage, content.Text, "The echoed message does not match the original")
	default:
		t.Fatalf("Unexpected content type: %T", content)
	}

	t.Log("INFO: E2E Test Scenario for WebRTC Echo Server Completed Successfully!")
}
