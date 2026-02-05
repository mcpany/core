// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package twilio_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_Twilio(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Twilio Server...")

	// --- Mock Twilio API Server ---
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		require.True(t, ok, "Basic auth credentials not provided")
		require.Equal(t, "test_account_sid", username, "Unexpected username")
		require.Equal(t, "test_auth_token", password, "Unexpected password")

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"sid":           "SM12345",
			"status":        "queued",
			"error_message": nil,
		})
	}))
	defer mockServer.Close()

	// --- 1. Start MCPANY Server ---
	t.Setenv("TWILIO_ACCOUNT_SID", "test_account_sid")
	t.Setenv("TWILIO_AUTH_TOKEN", "test_auth_token")
	t.Setenv("TWILIO_API_KEY", "test_key")
	t.Setenv("TWILIO_API_SECRET", "test_secret")
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2ETwilioServerTest", "--config-path", "../../../../examples/popular_services/twilio", "--set", "upstream_services[0].mcp_service.http_connection.http_address="+mockServer.URL)
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

	// --- 3. Call Tool ---
	args := json.RawMessage(`{"To": "+15551234567", "From": "+15557654321", "Body": "Hello, world!"}`)
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: registeredToolName, Arguments: args})
	require.NoError(t, err)
	require.NotNil(t, res)

	// --- 4. Assert Response ---
	require.Len(t, res.Content, 1, "Expected exactly one content item")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content")

	var twilioResponse map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &twilioResponse)
	require.NoError(t, err, "Failed to unmarshal JSON response")

	require.Contains(t, twilioResponse, "sid", "The response should contain a message SID")
	require.Contains(t, twilioResponse, "status", "The response should contain a status")

	require.Equal(t, "SM12345", twilioResponse["sid"])
	require.Equal(t, "queued", twilioResponse["status"])

	t.Log("INFO: E2E Test Scenario for Twilio Server Completed Successfully!")
}
