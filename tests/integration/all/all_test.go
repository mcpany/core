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

package all_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/mcpany/core/pkg/config"
	apiv1 "github.com/mcpany/core/proto/api/v1"
	"github.com/mcpany/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestLoadAllPopularServices(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for All Popular Services...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpxTestServerInfo := integration.StartMCPANYServer(t, "E2EAllPopularServicesTest")
	defer mcpxTestServerInfo.CleanupFunc()

	// --- 2. Register All Popular Services with MCPANY ---
	registrationGRPCClient := mcpxTestServerInfo.RegistrationClient

	// Find all the config files
	configs, err := filepath.Glob("../../../examples/popular_services/*/config.yaml")
	require.NoError(t, err)
	require.Greater(t, len(configs), 0, "No popular service configs found")

	// Create a new FileStore
	fs := afero.NewOsFs()
	store := config.NewFileStore(fs, configs)

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

	// --- 4. Assert Response ---
	require.Greater(t, len(listToolsResult.Tools), 0, "No tools were registered")

	for _, tool := range listToolsResult.Tools {
		t.Logf("Discovered tool from MCPANY: %s", tool.Name)
	}

	t.Log("INFO: E2E Test Scenario for All Popular Services Completed Successfully!")
}
