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

package google_calendar_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/mcpany/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_GoogleCalendar(t *testing.T) {
	if os.Getenv("GOOGLE_API_KEY") == "" {
		t.Skip("Skipping test because GOOGLE_API_KEY is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Google Calendar Server...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EGoogleCalendarServerTest", "--config-path", "../../../../examples/popular_services/google_calendar")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 2. Call Tool via MCPANY ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpAnyTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)
	require.Len(t, listToolsResult.Tools, 1, "Expected exactly one tool to be registered")

	// --- 3. Test Cases ---
	testCases := []struct {
		name          string
		calendarId    string
	}{
		{
			name:          "List events from a public calendar",
			calendarId:    "en.usa#holiday@group.v.calendar.google.com",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// --- 4. Call list_events Tool ---
			args := json.RawMessage(`{"calendarId": "` + tc.calendarId + `"}`)
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "google_calendar/-/list_events", Arguments: args})
			require.NoError(t, err)
			require.NotNil(t, res)

			// --- 5. Assert list_events Response ---
			require.Len(t, res.Content, 1, "Expected exactly one content item")
			textContent, ok := res.Content[0].(*mcp.TextContent)
			require.True(t, ok, "Expected text content")

			var response map[string]interface{}
			err = json.Unmarshal([]byte(textContent.Text), &response)
			require.NoError(t, err, "Failed to unmarshal JSON response")

			items, ok := response["items"].([]interface{})
			require.True(t, ok)
			require.NotEmpty(t, items, "Expected at least one event")
		})
	}

	t.Log("INFO: E2E Test Scenario for Google Calendar Server Completed Successfully!")
}
