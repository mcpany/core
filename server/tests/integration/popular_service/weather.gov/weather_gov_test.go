// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package weather_gov_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_WeatherGov(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Weather.gov Server...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EWeatherGovServerTest", "--config-path", "../../../../examples/popular_services/weather.gov")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 2. Call Tool via MCPANY ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpAnyTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)
	require.Len(t, listToolsResult.Tools, 2, "Expected exactly two tools to be registered")

	// --- 3. Test Cases ---
	testCases := []struct {
		name         string
		latitude     string
		longitude    string
		expectedCity string
	}{
		{
			name:         "Get weather for a valid location",
			latitude:     "39.7456",
			longitude:    "-97.0892",
			expectedCity: "Fairbury",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// --- 4. Call get_grid Tool ---
			gridArgs := json.RawMessage(`{"latitude": "` + tc.latitude + `", "longitude": "` + tc.longitude + `"}`)
			gridRes, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "get_grid", Arguments: gridArgs})
			require.NoError(t, err)
			require.NotNil(t, gridRes)

			// --- 5. Assert get_grid Response ---
			require.Len(t, gridRes.Content, 1, "Expected exactly one content item")
			gridTextContent, ok := gridRes.Content[0].(*mcp.TextContent)
			require.True(t, ok, "Expected text content")

			var gridResponse map[string]interface{}
			err = json.Unmarshal([]byte(gridTextContent.Text), &gridResponse)
			require.NoError(t, err, "Failed to unmarshal JSON response")

			properties, ok := gridResponse["properties"].(map[string]interface{})
			require.True(t, ok)
			forecastURL, ok := properties["forecast"].(string)
			require.True(t, ok)

			// --- 6. Call get_forecast Tool ---
			forecastArgs := json.RawMessage(`{"forecast_url": "` + forecastURL + `"}`)
			forecastRes, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "get_forecast", Arguments: forecastArgs})
			require.NoError(t, err)
			require.NotNil(t, forecastRes)

			// --- 7. Assert get_forecast Response ---
			require.Len(t, forecastRes.Content, 1, "Expected exactly one content item")
			forecastTextContent, ok := forecastRes.Content[0].(*mcp.TextContent)
			require.True(t, ok, "Expected text content")

			var forecastResponse map[string]interface{}
			err = json.Unmarshal([]byte(forecastTextContent.Text), &forecastResponse)
			require.NoError(t, err, "Failed to unmarshal JSON response")

			forecastProperties, ok := forecastResponse["properties"].(map[string]interface{})
			require.True(t, ok)
			periods, ok := forecastProperties["periods"].([]interface{})
			require.True(t, ok)
			require.NotEmpty(t, periods, "Expected at least one forecast period")
		})
	}

	t.Log("INFO: E2E Test Scenario for Weather.gov Server Completed Successfully!")
}
