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

func TestUpstreamService_GRPC(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for gRPC Calculator Server...")
	t.Parallel()

	// --- 1. Start gRPC Calculator Server ---
	grpcServerPort := integration.FindFreePort(t)
	grpcServerProc := integration.NewManagedProcess(t, "grpc_calculator_server", "../../../build/test/bin/grpc_calculator_server", []string{fmt.Sprintf("--port=%d", grpcServerPort)}, nil)
	err := grpcServerProc.Start()
	require.NoError(t, err, "Failed to start gRPC Calculator server")
	t.Cleanup(grpcServerProc.Stop)
	require.Eventually(t, func() bool {
		return integration.IsTCPPortAvailable(grpcServerPort)
	}, integration.ServiceStartupTimeout, integration.RetryInterval, "gRPC Calculator server did not become ready in time")

	// --- 2. Start MCPXY Server ---
	mcpxTestServerInfo := integration.StartMCPXYServer(t, "E2EGrpcCalculatorServerTest")
	defer mcpxTestServerInfo.CleanupFunc()

	// --- 3. Register gRPC Calculator Server with MCPXY ---
	const calcServiceID = "e2e_grpc_calculator"
	grpcServiceEndpoint := fmt.Sprintf("localhost:%d", grpcServerPort)
	t.Logf("INFO: Registering '%s' with MCPXY at endpoint %s...", calcServiceID, grpcServiceEndpoint)
	registrationGRPCClient := mcpxTestServerInfo.RegistrationClient
	integration.RegisterGRPCService(t, registrationGRPCClient, calcServiceID, grpcServiceEndpoint, nil)
	t.Logf("INFO: '%s' registered.", calcServiceID)

	// --- 4. Call Tool via MCPXY ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpxTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)
	for _, tool := range listToolsResult.Tools {
		t.Logf("Discovered tool from MCPXY: %s", tool.Name)
	}

	toolName := fmt.Sprintf("%s%sCalculatorAdd", calcServiceID, consts.ToolNameServiceSeparator)
	addArgs := `{"a": 10, "b": 20}`
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(addArgs)})
	require.NoError(t, err, "Error calling Add tool")
	require.NotNil(t, res, "Nil response from Add tool")
	switch content := res.Content[0].(type) {
	case *mcp.TextContent:
		require.JSONEq(t, `{"result": 30}`, content.Text, "The sum is incorrect")
	default:
		t.Fatalf("Unexpected content type: %T", content)
	}

	t.Log("INFO: E2E Test Scenario for gRPC Calculator Server Completed Successfully!")
}
