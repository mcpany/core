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

//go:build e2e

package wikipedia_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mcpany/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_Wikipedia(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Wikipedia Server...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EWikipediaServerTest", "--config-path", "../../../../examples/popular_services/wikipedia")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 2. Call Tool via MCPANY ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpAnyTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)
	require.Len(t, listToolsResult.Tools, 1, "Expected exactly one tool to be registered")
	registeredToolName := listToolsResult.Tools[0].Name
	t.Logf("Discovered tool from MCPANY: %s", registeredToolName)

	// --- 3. Test Cases ---
	testCases := []struct {
		name          string
		title         string
		expectedTitle string
		expectedPageID int
	}{
		{
			name:          "Pet Door",
			title:         "Pet_door",
			expectedTitle: "Pet door",
			expectedPageID: 3276454,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// --- 4. Call Tool ---
			args := json.RawMessage(`{"title": "` + tc.title + `"}`)
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: registeredToolName, Arguments: args})
			require.NoError(t, err)
			require.NotNil(t, res)

			// --- 5. Assert Response ---
			require.Len(t, res.Content, 1, "Expected exactly one content item")
			textContent, ok := res.Content[0].(*mcp.TextContent)
			require.True(t, ok, "Expected text content")

			var wikipediaResponse map[string]interface{}
			err = json.Unmarshal([]byte(textContent.Text), &wikipediaResponse)
			require.NoError(t, err, "Failed to unmarshal JSON response")

			require.Contains(t, wikipediaResponse, "parse", "The response should contain a parse object")

			parse, ok := wikipediaResponse["parse"].(map[string]interface{})
			require.True(t, ok, "Expected parse to be a map")

			require.Contains(t, parse, "title", "The response should contain a title")
			require.Contains(t, parse, "pageid", "The response should contain a pageid")
			require.Contains(t, parse, "text", "The response should contain text")

			require.Equal(t, tc.expectedTitle, parse["title"], "The title should match the expected value")
			require.Equal(t, float64(tc.expectedPageID), parse["pageid"], "The pageid should match the expected value")
		})
	}

	t.Log("INFO: E2E Test Scenario for Wikipedia Server Completed Successfully!")
}
