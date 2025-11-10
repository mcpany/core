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

//go:build e2e_popular_service

package ipinfo_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mcpany/core/pkg/config"
	apiv1 "github.com/mcpany/core/proto/api/v1"
	"github.com/mcpany/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_IPInfo(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for IP Info Server...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpxTestServerInfo := integration.StartMCPANYServer(t, "E2EIPInfoServerTest")
	defer mcpxTestServerInfo.CleanupFunc()

	// --- 2. Register IP Info Server with MCPANY ---
	registrationGRPCClient := mcpxTestServerInfo.RegistrationClient

	// Create a new FileStore
	fs := afero.NewOsFs()
	store := config.NewFileStore(fs, []string{"../../../../examples/popular_services/ipinfo.io/config.yaml"})

	// Load the config
	cfg, err := config.LoadServices(store, "server")
	require.NoError(t, err)

	// Register the service
	for _, service := range cfg.GetUpstreamServices() {
		req := apiv1.RegisterServiceRequest_builder{
			Config: service,
		}.Build()
		integration.RegisterServiceViaAPI(t, registrationGRPCClient, req)
	}

	// --- 3. Call Tool via MCPANY ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpxTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)
	require.Len(t, listToolsResult.Tools, 1, "Expected exactly one tool to be registered")
	registeredToolName := listToolsResult.Tools[0].Name
	t.Logf("Discovered tool from MCPANY: %s", registeredToolName)

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: registeredToolName, Arguments: json.RawMessage(`{}`)})
	require.NoError(t, err)
	require.NotNil(t, res)

	// --- 4. Assert Response ---
	require.Len(t, res.Content, 1, "Expected exactly one content item")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content")

	var ipInfoResponse map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &ipInfoResponse)
	require.NoError(t, err, "Failed to unmarshal JSON response")

	require.Contains(t, ipInfoResponse, "ip", "The response should contain an IP address")
	t.Logf("SUCCESS: Received correct IP info: %s", textContent.Text)

	t.Log("INFO: E2E Test Scenario for IP Info Server Completed Successfully!")
}
