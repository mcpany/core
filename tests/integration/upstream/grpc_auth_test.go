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
	"fmt"
	"testing"

	"github.com/mcpxy/core/pkg/util"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	"github.com/mcpxy/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestUpstreamService_GRPC_WithBearerAuth(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Authenticated gRPC Calculator Server...")
	t.Parallel()

	// --- 1. Start Authenticated gRPC Calculator Server ---
	grpcServerPort := integration.FindFreePort(t)
	grpcServerProc := integration.NewManagedProcess(t, "grpc_authed_calculator_server", "/tmp/build/test/bin/grpc_authed_calculator_server", []string{fmt.Sprintf("--port=%d", grpcServerPort)}, nil)
	err := grpcServerProc.Start()
	require.NoError(t, err, "Failed to start authenticated gRPC Calculator server")
	t.Cleanup(grpcServerProc.Stop)
	integration.WaitForTCPPort(t, grpcServerPort, integration.ServiceStartupTimeout)

	// --- 2. Start MCPXY Server ---
	mcpxTestServerInfo := integration.StartMCPXYServer(t, "E2EGrpcAuthedCalculatorServerTest")
	defer mcpxTestServerInfo.CleanupFunc()

	// --- 3. Register Authenticated gRPC Calculator Server with MCPXY ---
	const calcServiceID = "e2e_grpc_authed_calculator"
	grpcServiceEndpoint := fmt.Sprintf("localhost:%d", grpcServerPort)
	t.Logf("INFO: Registering '%s' with MCPXY at endpoint %s...", calcServiceID, grpcServiceEndpoint)
	registrationGRPCClient := mcpxTestServerInfo.RegistrationClient

	authConfig := configv1.UpstreamAuthentication_builder{
		BearerToken: configv1.UpstreamBearerTokenAuth_builder{
			Token: proto.String("test-bearer-token"),
		}.Build(),
	}.Build()

	integration.RegisterGRPCService(t, registrationGRPCClient, calcServiceID, grpcServiceEndpoint, authConfig)
	t.Logf("INFO: '%s' registered.", calcServiceID)

	// --- 4. Call Tool via MCPXY ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpxTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	serviceKey, _ := util.GenerateID(calcServiceID)
	toolName, _ := util.GenerateToolID(serviceKey, "CalculatorAdd")
	addArgs := `{"a": 10, "b": 20}`
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(addArgs)})
	require.NoError(t, err, "Error calling Add tool with correct auth")
	require.NotNil(t, res, "Nil response from Add tool with correct auth")
	switch content := res.Content[0].(type) {
	case *mcp.TextContent:
		require.JSONEq(t, `{"result": 30}`, content.Text, "The sum is incorrect")
	default:
		t.Fatalf("Unexpected content type: %T", content)
	}

	t.Log("INFO: E2E Test Scenario for Authenticated gRPC Calculator Server Completed Successfully!")
}
