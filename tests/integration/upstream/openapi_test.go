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
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/mcpxy/core/pkg/util"
	"github.com/mcpxy/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_OpenAPI(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for OpenAPI Calculator Server...")
	t.Parallel()

	// --- 1. Start OpenAPI Calculator Server ---
	openapiServerPort := integration.FindFreePort(t)
	openapiServerProc := integration.NewManagedProcess(t, "openapi_calculator_server", "/tmp/build/test/bin/openapi_calculator_server", []string{fmt.Sprintf("--port=%d", openapiServerPort)}, nil)
	err := openapiServerProc.Start()
	require.NoError(t, err, "Failed to start OpenAPI Calculator server")
	t.Cleanup(openapiServerProc.Stop)
	integration.WaitForTCPPort(t, openapiServerPort, integration.ServiceStartupTimeout*2)

	// --- 2. Start MCPXY Server ---
	mcpxTestServerInfo := integration.StartMCPXYServer(t, "E2EOpenAPICalculatorServerTest")
	defer mcpxTestServerInfo.CleanupFunc()

	// --- 3. Register OpenAPI Calculator Server with MCPXY ---
	const calcServiceID = "e2e_openapi_calculator"
	serverURL := fmt.Sprintf("http://localhost:%d", openapiServerPort)
	openapiSpecEndpoint := fmt.Sprintf("%s/openapi.json", serverURL)
	t.Logf("INFO: Fetching OpenAPI spec from %s...", openapiSpecEndpoint)

	resp, err := http.Get(openapiSpecEndpoint)
	require.NoError(t, err, "Failed to fetch OpenAPI spec from server")
	defer resp.Body.Close()
	specContent, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read OpenAPI spec content")

	tmpfile, err := os.CreateTemp("", "openapi-*.json")
	require.NoError(t, err, "Failed to create temp file for OpenAPI spec")
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(specContent)
	require.NoError(t, err, "Failed to write spec to temp file")
	err = tmpfile.Close()
	require.NoError(t, err, "Failed to close temp file")

	t.Logf("INFO: Registering '%s' with MCPXY using spec from temporary file %s...", calcServiceID, tmpfile.Name())
	registrationGRPCClient := mcpxTestServerInfo.RegistrationClient
	integration.RegisterOpenAPIService(t, registrationGRPCClient, calcServiceID, tmpfile.Name(), serverURL, nil)
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

	serviceKey, _ := util.GenerateID(calcServiceID)
	toolName, _ := util.GenerateToolID(serviceKey, "add")
	addArgs := `{"a": 5, "b": 7}`
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(addArgs)})
	require.NoError(t, err, "Error calling add tool")
	require.NotNil(t, res, "Nil response from add tool")
	switch content := res.Content[0].(type) {
	case *mcp.TextContent:
		require.JSONEq(t, `{"result": 12}`, content.Text, "The sum is incorrect")
	default:
		t.Fatalf("Unexpected content type: %T", content)
	}

	t.Log("INFO: E2E Test Scenario for OpenAPI Calculator Server Completed Successfully!")
}
