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

func TestUpstreamService_CatFacts(t *testing.T) {
	// t.Skip("Skipping flaky cat facts test due to rate limiting issues")
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Cat Facts Server...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2ECatFactsServerTest")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 2. Register Cat Facts Server with MCPANY ---
	const catFactsServiceID = "e2e_catfacts"
	catFactsServiceEndpoint := "https://catfact.ninja"
	t.Logf("INFO: Registering '%s' with MCPANY at endpoint %s...", catFactsServiceID, catFactsServiceEndpoint)
	registrationGRPCClient := mcpAnyTestServerInfo.RegistrationClient

	callID := "getCatFact"
	httpCall := configv1.HttpCallDefinition_builder{
		Id:           proto.String(callID),
		EndpointPath: proto.String("/fact"),
		Method:       configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HttpMethod_value["HTTP_METHOD_GET"]).Enum(),
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

		// Check for SDK error
		if err != nil {
			if strings.Contains(err.Error(), "503 Service Temporarily Unavailable") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "connection reset by peer") {
				t.Logf("Attempt %d/%d: Call to catfact.ninja failed with a transient error (SDK error): %v. Retrying...", i+1, maxRetries, err)
				time.Sleep(2 * time.Second) // Wait before retrying
				continue
			}
			require.NoError(t, err, "unrecoverable error calling getCatFact tool")
		}

		// Check for MCP protocol error (isError: true)
		if res.IsError {
			var errorMsg string
			if len(res.Content) > 0 {
				if textContent, ok := res.Content[0].(*mcp.TextContent); ok {
					errorMsg = textContent.Text
				}
			}

			if strings.Contains(errorMsg, "503 Service Temporarily Unavailable") || strings.Contains(errorMsg, "context deadline exceeded") || strings.Contains(errorMsg, "connection reset by peer") {
				t.Logf("Attempt %d/%d: Call to catfact.ninja failed with a transient error (MCP error): %s. Retrying...", i+1, maxRetries, errorMsg)
				time.Sleep(2 * time.Second) // Wait before retrying
				continue
			}
			// If it's a non-transient error, we stop retrying and let the assertions fail
			break
		}

		// Success
		break
	}

	if err != nil {
		// t.Skipf("Skipping test: all %d retries to catfact.ninja failed with transient errors. Last error: %v", maxRetries, err)
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

	require.NotEmpty(t, catFactResponse["fact"], "The fact should not be empty")
	require.NotEmpty(t, catFactResponse["length"], "The length should not be empty")
	t.Logf("SUCCESS: Received a cat fact: %s", textContent.Text)

	t.Log("INFO: E2E Test Scenario for Cat Facts Server Completed Successfully!")
}
