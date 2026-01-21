// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package stripe_test

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

func TestUpstreamService_Stripe(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Stripe Server...")

	// --- Mock Stripe API Server ---
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":       "ch_12345",
			"amount":   1000,
			"currency": "usd",
			"status":   "succeeded",
		})
	}))
	defer mockServer.Close()

	// --- 1. Start MCPANY Server ---
	t.Setenv("STRIPE_API_KEY", "test_key")
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EStripeServerTest", "--config-path", "../../../../examples/popular_services/stripe", "--set", "upstream_services[0].openapi_service.address="+mockServer.URL)
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
	args := json.RawMessage(`{"amount": 1000, "currency": "usd", "source": "tok_visa"}`)
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: registeredToolName, Arguments: args})
	require.NoError(t, err)
	require.NotNil(t, res)

	// --- 4. Assert Response ---
	require.Len(t, res.Content, 1, "Expected exactly one content item")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content")

	var stripeResponse map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &stripeResponse)
	require.NoError(t, err, "Failed to unmarshal JSON response")

	require.Contains(t, stripeResponse, "id", "The response should contain a charge ID")
	require.Contains(t, stripeResponse, "amount", "The response should contain an amount")
	require.Contains(t, stripeResponse, "currency", "The response should contain a currency")
	require.Contains(t, stripeResponse, "status", "The response should contain a status")

	require.Equal(t, "ch_12345", stripeResponse["id"])
	require.Equal(t, float64(1000), stripeResponse["amount"])
	require.Equal(t, "usd", stripeResponse["currency"])
	require.Equal(t, "succeeded", stripeResponse["status"])

	t.Log("INFO: E2E Test Scenario for Stripe Server Completed Successfully!")
}
