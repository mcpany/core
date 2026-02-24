// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package google_calendar_test

import (
	"context"
	"encoding/json"
	"testing"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestUpstreamService_GoogleCalendar(t *testing.T) {
	// No longer skipping due to missing API Key
	// if os.Getenv("GOOGLE_API_KEY") == "" {
	// 	t.Skip("Skipping test because GOOGLE_API_KEY is not set")
	// }

	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Google Calendar Server...")
	t.Parallel()

	// --- 1. Start Mock Google API Server ---
	mockResponse := `{
		"kind": "calendar#events",
		"summary": "Holidays in United States",
		"items": [
			{
				"kind": "calendar#event",
				"id": "20231225_60o30c1g60o30c1g60o30c1g60",
				"status": "confirmed",
				"summary": "Christmas Day",
				"start": { "date": "2023-12-25" },
				"end": { "date": "2023-12-26" }
			}
		]
	}`
	// The path depends on how the tool constructs the URL.
	// Typically /calendars/{calendarId}/events
	// The calendar ID used in test is "en.usa#holiday@group.v.calendar.google.com"
	// URL encoded: "en.usa%23holiday%40group.v.calendar.google.com"
	// Note: mockHandler uses r.URL.Path which is decoded, so we use the unencoded path here.
	mockHandler := integration.DefaultMockHandler(t, map[string]string{
		"/calendars/en.usa#holiday@group.v.calendar.google.com/events": mockResponse,
	})
	mockServer := integration.StartMockServer(t, mockHandler)
	defer mockServer.Close()

	// --- 2. Start MCPANY Server (No Config File) ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EGoogleCalendarServerTest")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 3. Register Service Dynamically ---
	// We manually register the service to point to our mock server instead of Google API
	// Note: We need to define the tool/call definition here to match what the YAML would have provided.
	// Assuming the original YAML defined "list_events".
	// The original YAML likely used `openapi_service` or `http_service` with pre-defined calls.
	// Since we are replacing the config, we must replicate the relevant parts.
	// For simplicity, we'll use `http_service` with manual call definition which is easier to point to mock.

	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("google_calendar"),
		HttpService: configv1.HttpUpstreamService_builder{
			Address: proto.String(mockServer.URL), // Point to mock server
			Tools: []*configv1.ToolDefinition{
				configv1.ToolDefinition_builder{
					Name: proto.String("list_events"),
					CallId: proto.String("list_events"),
				}.Build(),
			},
			Calls: map[string]*configv1.HttpCallDefinition{
				"list_events": configv1.HttpCallDefinition_builder{
					Method:       configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HttpMethod_value["HTTP_METHOD_GET"]).Enum(),
					EndpointPath: proto.String("/calendars/{{calendarId}}/events"),
					Parameters: []*configv1.HttpParameterMapping{
						configv1.HttpParameterMapping_builder{
							Schema: configv1.ParameterSchema_builder{
								Name: proto.String("calendarId"),
								Type: configv1.ParameterType(configv1.ParameterType_value["STRING"]).Enum(),
							}.Build(),
						}.Build(),
					},
				}.Build(),
			},
		}.Build(),
	}.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	integration.RegisterServiceViaAPI(t, mcpAnyTestServerInfo.RegistrationClient, req)

	// --- 4. Call Tool via MCPANY ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpAnyTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)
	// We might have other default tools if any, but we expect ours.
	// Actually StartMCPANYServer might not load anything by default.
	found := false
	var toolNames []string
	for _, tool := range listToolsResult.Tools {
		toolNames = append(toolNames, tool.Name)
		if tool.Name == "google_calendar.list_events" {
			found = true
			break
		}
	}
	require.Truef(t, found, "Expected google_calendar.list_events tool to be registered. Found: %v", toolNames)

	// --- 5. Test Cases ---
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
			// --- 6. Call list_events Tool ---
			args := json.RawMessage(`{"calendarId": "` + tc.calendarId + `"}`)
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "google_calendar.list_events", Arguments: args})
			require.NoError(t, err)
			require.NotNil(t, res)

			// --- 7. Assert list_events Response ---
			require.Len(t, res.Content, 1, "Expected exactly one content item")
			textContent, ok := res.Content[0].(*mcp.TextContent)
			require.True(t, ok, "Expected text content")

			var response map[string]interface{}
			err = json.Unmarshal([]byte(textContent.Text), &response)
			require.NoError(t, err, "Failed to unmarshal JSON response")

			items, ok := response["items"].([]interface{})
			require.True(t, ok)
			require.NotEmpty(t, items, "Expected at least one event")

			// Verify content matches mock
			item0 := items[0].(map[string]interface{})
			require.Equal(t, "Christmas Day", item0["summary"])
		})
	}

	t.Log("INFO: E2E Test Scenario for Google Calendar Server Completed Successfully!")
}
