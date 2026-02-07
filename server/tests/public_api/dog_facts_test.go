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

	apiv1 "github.com/mcpany/core/proto/api/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestUpstreamService_DogFacts(t *testing.T) {
	// Unskipped and mocked for stability
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Dog Facts Server...")
	t.Parallel()

	// --- Mock Upstream ---
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/facts" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"facts": ["Dogs are great."], "success": true}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer mockServer.Close()

	// --- 1. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EDogFactsServerTest")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 2. Register Dog Facts Server with MCPANY ---
	const dogFactsServiceID = "e2e_dogfacts"
	dogFactsServiceEndpoint := mockServer.URL
	t.Logf("INFO: Registering '%s' with MCPANY at endpoint %s...", dogFactsServiceID, dogFactsServiceEndpoint)
	registrationGRPCClient := mcpAnyTestServerInfo.RegistrationClient

	callID := "getDogFact"
	httpCall := configv1.HttpCallDefinition_builder{
		Id:           proto.String(callID),
		EndpointPath: proto.String("/api/facts"),
		Method:       configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HTTP_METHOD_GET).Enum(),
	}.Build()

	toolDef := configv1.ToolDefinition_builder{
		Name:   proto.String("getDogFact"),
		CallId: proto.String(callID),
	}.Build()

	httpService := configv1.HttpUpstreamService_builder{
		Address: proto.String(dogFactsServiceEndpoint),
		Tools:   []*configv1.ToolDefinition{toolDef},
		Calls:   map[string]*configv1.HttpCallDefinition{callID: httpCall},
	}.Build()

	config := configv1.UpstreamServiceConfig_builder{
		Name:        proto.String(dogFactsServiceID),
		HttpService: httpService,
	}.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	integration.RegisterServiceViaAPI(t, registrationGRPCClient, req)
	t.Logf("INFO: '%s' registered.", dogFactsServiceID)

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

	serviceID, _ := util.SanitizeServiceName(dogFactsServiceID)
	sanitizedToolName, _ := util.SanitizeToolName("getDogFact")
	toolName := serviceID + "." + sanitizedToolName

	const maxRetries = 3
	var res *mcp.CallToolResult

	for i := 0; i < maxRetries; i++ {
		res, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(`{}`)})
		if err != nil {
			t.Logf("Attempt %d/%d failed: %v", i+1, maxRetries, err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		break
	}

	require.NoError(t, err, "Error calling getDogFact tool")
	require.NotNil(t, res, "Nil response from getDogFact tool")

	// --- 4. Assert Response ---
	require.Len(t, res.Content, 1, "Expected exactly one content item")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content")

	var dogFactResponse map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &dogFactResponse)
	require.NoError(t, err, "Failed to unmarshal JSON response")

	require.NotNil(t, dogFactResponse["facts"], "The facts should not be nil")
	require.Equal(t, true, dogFactResponse["success"], "The success should be true")
	t.Logf("SUCCESS: Received a dog fact: %s", textContent.Text)

	t.Log("INFO: E2E Test Scenario for Dog Facts Server Completed Successfully!")
}
