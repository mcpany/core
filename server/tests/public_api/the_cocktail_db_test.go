// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e_public_api

package public_api

import (
	"context"
	"encoding/json"
	"fmt"
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

func TestUpstreamService_TheCocktailDB(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for TheCocktailDB Server...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2ETheCocktailDBServerTest")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 2. Register TheCocktailDB Server with MCPANY ---
	const theCocktailDBServiceID = "e2e_thecocktaildb"
	theCocktailDBServiceEndpoint := "https://www.thecocktaildb.com"
	t.Logf("INFO: Registering '%s' with MCPANY at endpoint %s...", theCocktailDBServiceID, theCocktailDBServiceEndpoint)
	registrationGRPCClient := mcpAnyTestServerInfo.RegistrationClient

	callID := "searchCocktail"
	httpCall := configv1.HttpCallDefinition_builder{
		Id:           proto.String(callID),
		EndpointPath: proto.String("/api/json/v1/1/search.php?s={{drink}}"),
		Method:       configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HttpMethod_value["HTTP_METHOD_GET"]).Enum(),
		Parameters: []*configv1.HttpParameterMapping{
			configv1.HttpParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("drink"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	toolDef := configv1.ToolDefinition_builder{
		Name:   proto.String("searchCocktail"),
		CallId: proto.String(callID),
	}.Build()

	httpService := configv1.HttpUpstreamService_builder{
		Address: proto.String(theCocktailDBServiceEndpoint),
		Tools:   []*configv1.ToolDefinition{toolDef},
		Calls:   map[string]*configv1.HttpCallDefinition{callID: httpCall},
	}.Build()

	config := configv1.UpstreamServiceConfig_builder{
		Name:        proto.String(theCocktailDBServiceID),
		HttpService: httpService,
	}.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	integration.RegisterServiceViaAPI(t, registrationGRPCClient, req)
	t.Logf("INFO: '%s' registered.", theCocktailDBServiceID)

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

	serviceID, _ := util.SanitizeServiceName(theCocktailDBServiceID)
	sanitizedToolName, _ := util.SanitizeToolName("searchCocktail")
	toolName := serviceID + "." + sanitizedToolName
	drink := `{"drink": "margarita"}`

	const maxRetries = 3
	var res *mcp.CallToolResult

	for i := 0; i < maxRetries; i++ {
		res, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(drink)})

		// Check for soft failure (successful tool call, but content indicates failure)
		if err == nil && len(res.Content) > 0 {
			if textContent, ok := res.Content[0].(*mcp.TextContent); ok {
				if strings.Contains(textContent.Text, "upstream HTTP request failed with status 500") {
					err = fmt.Errorf("upstream 500 error: %s", textContent.Text)
				}
			}
		}

		if err == nil {
			break // Success
		}

		if strings.Contains(err.Error(), "503 Service Temporarily Unavailable") ||
			strings.Contains(err.Error(), "context deadline exceeded") ||
			strings.Contains(err.Error(), "connection reset by peer") ||
			strings.Contains(err.Error(), "upstream 500 error") {
			t.Logf("Attempt %d/%d: Call to thecocktaildb.com failed with a transient error: %v. Retrying...", i+1, maxRetries, err)
			time.Sleep(2 * time.Second) // Wait before retrying
			continue
		}

		require.NoError(t, err, "unrecoverable error calling searchCocktail tool")
	}

	if err != nil {
		// t.Skipf("Skipping test: all %d retries to thecocktaildb.com failed with transient errors. Last error: %v", maxRetries, err)
	}

	require.NoError(t, err, "Error calling searchCocktail tool")
	require.NotNil(t, res, "Nil response from searchCocktail tool")

	// --- 4. Assert Response ---
	require.Len(t, res.Content, 1, "Expected exactly one content item")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content")

	t.Logf("Response body: %s", textContent.Text)

	var theCocktailDBResponse map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &theCocktailDBResponse)
	require.NoError(t, err, "Failed to unmarshal JSON response")

	if _, ok := theCocktailDBResponse["drinks"].(string); ok {
		// t.Skip("Skipping test, no drinks found in response")
	}
	if theCocktailDBResponse["drinks"] == nil {
		// t.Skip("Skipping test, no drinks found in response")
	}
	drinks, ok := theCocktailDBResponse["drinks"].([]interface{})
	require.True(t, ok, "The drinks should be an array")
	require.True(t, len(drinks) > 0, "The response should contain at least one drink")

	t.Logf("SUCCESS: Received correct data for margarita: %s", textContent.Text)

	t.Log("INFO: E2E Test Scenario for TheCocktailDB Server Completed Successfully!")
}
