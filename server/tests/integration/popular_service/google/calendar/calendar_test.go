// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package calendar_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mcpany/core/server/tests/integration"
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
