// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e_public_api

package public_api

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/util"
	apiv1 "github.com/mcpany/core/proto/api/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestUpstreamService_DeckOfCards(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Deck of Cards Server...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EDeckOfCardsServerTest")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 2. Register Deck of Cards Server with MCPANY ---
	const deckOfCardsServiceID = "e2e_deckofcards"
	deckOfCardsServiceEndpoint := "https://deckofcardsapi.com"
	t.Logf("INFO: Registering '%s' with MCPANY at endpoint %s...", deckOfCardsServiceID, deckOfCardsServiceEndpoint)
	registrationGRPCClient := mcpAnyTestServerInfo.RegistrationClient

	callID := "shuffleDeck"
	httpCall := configv1.HttpCallDefinition_builder{
		Id:           proto.String(callID),
		EndpointPath: proto.String("/api/deck/new/shuffle/?deck_count=1"),
		Method:       configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HttpMethod_value["HTTP_METHOD_GET"]).Enum(),
	}.Build()

	toolDef := configv1.ToolDefinition_builder{
		Name:   proto.String("shuffleDeck"),
		CallId: proto.String(callID),
	}.Build()

	httpService := configv1.HttpUpstreamService_builder{
		Address: proto.String(deckOfCardsServiceEndpoint),
		Tools:   []*configv1.ToolDefinition{toolDef},
		Calls:   map[string]*configv1.HttpCallDefinition{callID: httpCall},
	}.Build()

	config := configv1.UpstreamServiceConfig_builder{
		Name:        proto.String(deckOfCardsServiceID),
		HttpService: httpService,
	}.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	integration.RegisterServiceViaAPI(t, registrationGRPCClient, req)
	t.Logf("INFO: '%s' registered.", deckOfCardsServiceID)

	// --- 3. Call Tool via MCPANY ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpAnyTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)
	for _, tool := range listToolsResult.Tools {
		t.Logf("Discovered tool from MCPANY: %s", tool.Name)
	}

	serviceID, _ := util.SanitizeServiceName(deckOfCardsServiceID)
	sanitizedToolName, _ := util.SanitizeToolName("shuffleDeck")
	toolName := serviceID + "." + sanitizedToolName

	const maxRetries = 3
	var res *mcp.CallToolResult

	for i := 0; i < maxRetries; i++ {
		res, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(`{}`)})
		if err == nil {
			// Check if result is valid JSON. If the upstream returns HTML (e.g. rate limit page) or plain text error,
			// the MCP server might pass it through as TextContent, and subsequent Unmarshal will fail.
			if len(res.Content) > 0 {
				if txt, ok := res.Content[0].(*mcp.TextContent); ok {
					var tmp map[string]interface{}
					if jsonErr := json.Unmarshal([]byte(txt.Text), &tmp); jsonErr != nil {
						t.Logf("Attempt %d/%d: Call succeeded but returned invalid JSON (likely rate limit/upstream error): %v. Body start: %.50s. Retrying...", i+1, maxRetries, jsonErr, txt.Text)
						time.Sleep(2 * time.Second)
						continue
					}
				}
			}
			break // Success
		}

		if strings.Contains(err.Error(), "503 Service Temporarily Unavailable") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "connection reset by peer") {
			t.Logf("Attempt %d/%d: Call to deckofcardsapi.com failed with a transient error: %v. Retrying...", i+1, maxRetries, err)
			time.Sleep(2 * time.Second) // Wait before retrying
			continue
		}

		require.NoError(t, err, "unrecoverable error calling shuffleDeck tool")
	}

	if err != nil {
		// t.Skipf("Skipping test: all %d retries to deckofcardsapi.com failed with transient errors. Last error: %v", maxRetries, err)
	}

	require.NoError(t, err, "Error calling shuffleDeck tool")
	require.NotNil(t, res, "Nil response from shuffleDeck tool")

	// --- 4. Assert Response ---
	require.Len(t, res.Content, 1, "Expected exactly one content item")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content")

	var deckOfCardsResponse map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &deckOfCardsResponse)
	require.NoError(t, err, "Failed to unmarshal JSON response")

	require.Equal(t, true, deckOfCardsResponse["success"], "The success should be true")
	require.NotEmpty(t, deckOfCardsResponse["deck_id"], "The deck_id should not be empty")
	require.Equal(t, true, deckOfCardsResponse["shuffled"], "The shuffled should be true")
	require.NotEmpty(t, deckOfCardsResponse["remaining"], "The remaining should not be empty")
	t.Logf("SUCCESS: Received a shuffled deck: %s", textContent.Text)

	t.Log("INFO: E2E Test Scenario for Deck of Cards Server Completed Successfully!")
}
