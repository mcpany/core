// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e_public_api

package public_api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestUpstreamService_CatFacts(t *testing.T) {
	// Unskipped and mocked for stability
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Cat Facts Server...")
	t.Parallel()

	// --- Mock Upstream ---
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/fact" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"fact": "Cats are cool", "length": 13}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer mockServer.Close()

	// --- 1. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2ECatFactsServerTest")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 2. Register Cat Facts Server with MCPANY ---
	const catFactsServiceID = "e2e_catfacts"
	catFactsServiceEndpoint := mockServer.URL
	t.Logf("INFO: Registering '%s' with MCPANY at endpoint %s...", catFactsServiceID, catFactsServiceEndpoint)
	registrationGRPCClient := mcpAnyTestServerInfo.RegistrationClient

	callID := "getCatFact"
	httpCall := configv1.HttpCallDefinition_builder{
		Id:           proto.String(callID),
		EndpointPath: proto.String("/fact"),
		Method:       configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HTTP_METHOD_GET).Enum(),
	}.Build()

	toolDef := configv1.ToolDefinition_builder{
		Name:   proto.String("getCatFact"),
		CallId: proto.String(callID),
	}.Build()

	httpService := configv1.HttpUpstreamService_builder{
		Address: proto.String(catFactsServiceEndpoint),
		Tools:   []*configv1.ToolDefinition{toolDef},
		Calls:   map[string]*configv1.HttpCallDefinition{callID: httpCall},
	}.Build()

	config := configv1.UpstreamServiceConfig_builder{
		Name:        proto.String(catFactsServiceID),
		HttpService: httpService,
	}.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	integration.RegisterServiceViaAPI(t, registrationGRPCClient, req)
	t.Logf("INFO: '%s' registered.", catFactsServiceID)

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

	serviceID, _ := util.SanitizeServiceName(catFactsServiceID)
	sanitizedToolName, _ := util.SanitizeToolName("getCatFact")
	toolName := serviceID + "." + sanitizedToolName

	const maxRetries = 3
	var res *mcp.CallToolResult

	for i := 0; i < maxRetries; i++ {
		res, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(`{}`)})

		if err != nil {
			// Should not happen with mock server
			t.Logf("Attempt %d/%d failed: %v", i+1, maxRetries, err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		break
	}

	require.NoError(t, err, "Error calling getCatFact tool")
	require.NotNil(t, res, "Nil response from getCatFact tool")

	// --- 4. Assert Response ---
	require.Len(t, res.Content, 1, "Expected exactly one content item")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content")

	var catFactResponse map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &catFactResponse)
	require.NoError(t, err, "Failed to unmarshal JSON response")

	require.Equal(t, "Cats are cool", catFactResponse["fact"])
	require.Equal(t, float64(13), catFactResponse["length"])
	t.Logf("SUCCESS: Received a cat fact: %s", textContent.Text)

	t.Log("INFO: E2E Test Scenario for Cat Facts Server Completed Successfully!")
}
