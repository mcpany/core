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
				"start": {
					"date": "2023-12-25"
				},
				"end": {
					"date": "2023-12-26"
				}
			}
		]
	}`

	// Note: mockHandler uses r.URL.Path which is decoded, so we use the unencoded path here.
	mockHandler := integration.DefaultMockHandler(t, map[string]string{
		"/calendars/en.usa#holiday@group.v.calendar.google.com/events": mockResponse,
	})
	mockServer := integration.StartMockServer(t, mockHandler)
	defer mockServer.Close()

	// --- 2. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EGoogleCalendarTest")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// We manually register the service to point to our mock server instead of Google API
	// because the mock server address changes on every run.

	serviceID := "e2e_google_calendar"
	// For simplicity, we'll use `http_service` with manual call definition which is easier to point to mock.
	config := &configv1.UpstreamServiceConfig{
		Id:   proto.String(serviceID),
		Name: proto.String(serviceID),
		HttpService: &configv1.HttpUpstreamService{
			Address: proto.String(mockServer.URL), // Point to mock server
			Tools: []*configv1.ToolDefinition{
				{
					Name:   proto.String("getHolidays"),
					CallId: proto.String("getHolidaysCall"),
				},
			},
			Calls: map[string]*configv1.HttpCallDefinition{
				"getHolidaysCall": {
					Id:           proto.String("getHolidaysCall"),
					EndpointPath: proto.String("/calendars/en.usa#holiday@group.v.calendar.google.com/events"),
					Method:       configv1.HttpCallDefinition_HTTP_METHOD_GET.Enum(),
					Parameters: []*configv1.HttpParameterMapping{
						{
							Schema: &configv1.ParameterSchema{
								Name: proto.String("timeMin"),
							},
							In: configv1.HttpParameterMapping_QUERY.Enum(),
						},
						{
							Schema: &configv1.ParameterSchema{
								Name: proto.String("timeMax"),
							},
							In: configv1.HttpParameterMapping_QUERY.Enum(),
						},
					},
				},
			},
		},
	}

	req := &apiv1.RegisterServiceRequest{
		Config: config,
	}

	integration.RegisterServiceViaAPI(t, mcpAnyTestServerInfo.RegistrationClient, req)
	t.Logf("INFO: '%s' registered.", serviceID)

	// --- 3. Call Tool via MCPANY ---
	client, session, cleanup := integration.ConnectToMCPServer(t, ctx, mcpAnyTestServerInfo.MCPAddress, mcpAnyTestServerInfo.APIKey)
	defer cleanup()
	defer session.Close()

	t.Run("ListTools", func(t *testing.T) {
		listToolsResult, err := client.ListTools(ctx, mcp.ListToolsRequest{})
		require.NoError(t, err)

		found := false
		for _, tool := range listToolsResult.Tools {
			if tool.Name == serviceID+".getHolidays" {
				found = true
				break
			}
		}
		require.True(t, found, "Tool getHolidays should be present")
	})

	t.Run("CallTool getHolidays", func(t *testing.T) {
		res, err := client.CallTool(ctx, mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
			}{
				Name: serviceID + ".getHolidays",
				Arguments: map[string]any{
					"timeMin": "2023-12-01T00:00:00Z",
					"timeMax": "2023-12-31T23:59:59Z",
				},
			},
		})
		require.NoError(t, err, "Error calling getHolidays tool")
		require.NotNil(t, res, "Nil response from getHolidays tool")
		require.False(t, res.IsError, "Tool execution returned an error")

		require.Len(t, res.Content, 1, "Expected exactly one content item")
		textContent, ok := res.Content[0].(mcp.TextContent)
		require.True(t, ok, "Expected text content")

		t.Logf("Response body: %s", textContent.Text)

		// Parse the result as JSON
		var calendarResponse map[string]interface{}
		err = json.Unmarshal([]byte(textContent.Text), &calendarResponse)
		require.NoError(t, err, "Failed to unmarshal JSON response")

		// Basic assertions based on the mock data
		require.Equal(t, "calendar#events", calendarResponse["kind"], "Unexpected kind")
		require.Equal(t, "Holidays in United States", calendarResponse["summary"], "Unexpected summary")

		items, ok := calendarResponse["items"].([]interface{})
		require.True(t, ok, "Items should be an array")
		require.Len(t, items, 1, "Expected 1 item")

		firstItem, ok := items[0].(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, "Christmas Day", firstItem["summary"])

		t.Logf("SUCCESS: Received correct Google Calendar events: %v", firstItem["summary"])
	})

	t.Log("INFO: E2E Test Scenario for Google Calendar Server Completed Successfully!")
}
