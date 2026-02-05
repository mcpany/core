// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package ipinfo_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_IPInfo(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for IP Info Server...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EIPInfoServerTest", "--config-path", "../../../../examples/popular_services/ipinfo.io")
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
		name            string
		ip              string
		expectedCity    string
		expectedCountry string
	}{
		{
			name:            "IPv4 address",
			ip:              "130.184.0.1",
			expectedCity:    "Tulsa",
			expectedCountry: "US",
		},
		{
			name:            "IPv6 address",
			ip:              "2607:f6d0:0:53::64:53",
			expectedCity:    "San Jose",
			expectedCountry: "US",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// --- 4. Call Tool ---
			args := json.RawMessage(`{}`)
			if tc.ip != "" {
				args = json.RawMessage(`{"ip": "` + tc.ip + `"}`)
			}
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: registeredToolName, Arguments: args})
			require.NoError(t, err)
			require.NotNil(t, res)

			// --- 5. Assert Response ---
			require.Len(t, res.Content, 1, "Expected exactly one content item")
			textContent, ok := res.Content[0].(*mcp.TextContent)
			require.True(t, ok, "Expected text content")

			var ipInfoResponse map[string]interface{}
			err = json.Unmarshal([]byte(textContent.Text), &ipInfoResponse)
			require.NoError(t, err, "Failed to unmarshal JSON response")

			require.Contains(t, ipInfoResponse, "ip", "The response should contain an IP address")
			require.Contains(t, ipInfoResponse, "city", "The response should contain a city")
			require.Contains(t, ipInfoResponse, "country", "The response should contain a country")

			if tc.ip != "" {
				require.Equal(t, tc.ip, ipInfoResponse["ip"], "The IP address should match the input")
			}
			require.Equal(t, tc.expectedCity, ipInfoResponse["city"], "The city should match the expected value")
			require.Equal(t, tc.expectedCountry, ipInfoResponse["country"], "The country should match the expected value")

			// --- 6. Test with token ---
			if os.Getenv("IPINFO_API_TOKEN") != "" {
				t.Run("with token", func(t *testing.T) {
					res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: registeredToolName, Arguments: args})
					require.NoError(t, err)
					require.NotNil(t, res)
				})
			}
		})
	}

	t.Log("INFO: E2E Test Scenario for IP Info Server Completed Successfully!")
}
