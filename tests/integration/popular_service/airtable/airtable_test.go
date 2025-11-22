/*
 * Copyright 2024 Author(s) of MCP Any
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

//go:build e2e

package airtable_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/mcpany/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"github.com/mcpany/core/proto/config/v1"
)

func TestUpstreamService_Airtable(t *testing.T) {
	airtableAPIToken := os.Getenv("AIRTABLE_API_TOKEN")
	if airtableAPIToken == "" {
		t.Skip("Skipping Airtable test because AIRTABLE_API_TOKEN is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Airtable Server...")
	t.Parallel()

	// --- 1. Create a temporary config file with the Airtable API token ---
	config := &configv1.UpstreamServiceConfig{
		Name: "airtable",
		HttpService: &configv1.HttpUpstreamService{
			Address: "https://api.airtable.com",
			Calls: []*configv1.HttpCallDefinition{
				{
					OperationId:  "list_bases",
					Description:  "List Airtable bases",
					Method:       configv1.HttpCallDefinition_HTTP_METHOD_GET,
					EndpointPath: "/v0/meta/bases",
				},
			},
		},
		UpstreamAuthentication: &configv1.UpstreamAuthentication{
			BearerToken: &configv1.BearerToken{
				Token: airtableAPIToken,
			},
		},
	}

	tempConfigFile := integration.CreateTempConfigFile(t, config)

	// --- 2. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EAirtableServerTest", "--config-path", tempConfigFile)
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 3. Call Tool via MCPANY ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpAnyTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)
	require.Len(t, listToolsResult.Tools, 1, "Expected exactly one tool to be registered")
	registeredToolName := listToolsResult.Tools[0].Name
	t.Logf("Discovered tool from MCPANY: %s", registeredToolName)

	// --- 4. Call Tool ---
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: registeredToolName})
	require.NoError(t, err)
	require.NotNil(t, res)

	// --- 5. Assert Response ---
	require.Len(t, res.Content, 1, "Expected exactly one content item")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content")

	var airtableResponse map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &airtableResponse)
	require.NoError(t, err, "Failed to unmarshal JSON response")

	require.Contains(t, airtableResponse, "bases", "The response should contain a list of bases")
}
