//go:build e2e

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

package calendar_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mcpany/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_GoogleCalendar(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Google Calendar Server...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EGoogleCalendarServerTest", "--config-path", "../../../../../examples/popular_services/google/calendar")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 2. Call Tool via MCPANY ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpAnyTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)
	require.NotEmpty(t, listToolsResult.Tools, "Expected at least one tool to be registered")
	t.Logf("Discovered %d tools from MCPANY", len(listToolsResult.Tools))

	// --- 3. Find the calendar.events.list tool ---
	var calendarListTool *mcp.Tool
	for _, tool := range listToolsResult.Tools {
		if tool.Name == "google_calendar/-/calendar.events.list" {
			calendarListTool = tool
			break
		}
	}
	require.NotNil(t, calendarListTool, "Expected to find the calendar.events.list tool")

	// --- 4. Call the calendar.events.list tool ---
	args := json.RawMessage(`{"calendarId": "primary"}`)
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: calendarListTool.Name, Arguments: args})
	require.NoError(t, err)
	require.NotNil(t, res)

	// --- 5. Assert Response ---
	require.Len(t, res.Content, 1, "Expected exactly one content item")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content")

	var calendarResponse map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &calendarResponse)
	require.NoError(t, err, "Failed to unmarshal JSON response")

	require.Contains(t, calendarResponse, "items", "The response should contain a list of events")
}
