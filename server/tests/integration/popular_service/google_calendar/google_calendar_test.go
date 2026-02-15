//go:build e2e

package google_calendar_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_GoogleCalendar(t *testing.T) {
	if os.Getenv("GOOGLE_API_KEY") == "" {
		// t.Skip("Skipping test because GOOGLE_API_KEY is not set")
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
		name       string
		calendarId string
	}{
		{
			name:       "List events from a public calendar",
			calendarId: "en.usa#holiday@group.v.calendar.google.com",
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
