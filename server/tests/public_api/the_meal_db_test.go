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

func TestUpstreamService_TheMealDB(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for TheMealDB Server...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2ETheMealDBServerTest")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 2. Register TheMealDB Server with MCPANY ---
	const theMealDBServiceID = "e2e_themealdb"
	theMealDBServiceEndpoint := "https://www.themealdb.com"
	t.Logf("INFO: Registering '%s' with MCPANY at endpoint %s...", theMealDBServiceID, theMealDBServiceEndpoint)
	registrationGRPCClient := mcpAnyTestServerInfo.RegistrationClient

	callID := "searchMeal"
	httpCall := configv1.HttpCallDefinition_builder{
		Id:           proto.String(callID),
		EndpointPath: proto.String("/api/json/v1/1/search.php?s={{meal}}"),
		Method:       configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HttpMethod_value["HTTP_METHOD_GET"]).Enum(),
		Parameters: []*configv1.HttpParameterMapping{
			configv1.HttpParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("meal"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	toolDef := configv1.ToolDefinition_builder{
		Name:   proto.String("searchMeal"),
		CallId: proto.String(callID),
	}.Build()

	httpService := configv1.HttpUpstreamService_builder{
		Address: proto.String(theMealDBServiceEndpoint),
		Tools:   []*configv1.ToolDefinition{toolDef},
		Calls:   map[string]*configv1.HttpCallDefinition{callID: httpCall},
	}.Build()

	config := configv1.UpstreamServiceConfig_builder{
		Name:        proto.String(theMealDBServiceID),
		HttpService: httpService,
	}.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	integration.RegisterServiceViaAPI(t, registrationGRPCClient, req)
	t.Logf("INFO: '%s' registered.", theMealDBServiceID)

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

	serviceID, _ := util.SanitizeServiceName(theMealDBServiceID)
	sanitizedToolName, _ := util.SanitizeToolName("searchMeal")
	toolName := serviceID + "." + sanitizedToolName
	meal := `{"meal": "arrabiata"}`

	const maxRetries = 3
	var res *mcp.CallToolResult

	for i := 0; i < maxRetries; i++ {
		res, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(meal)})
		if err == nil {
			break // Success
		}

		// Check for common public API failures (5xx, rate limits, blocks)
		errStr := err.Error()
		if strings.Contains(errStr, "503 Service Temporarily Unavailable") ||
			strings.Contains(errStr, "521") || // Cloudflare/Web Server Down
			strings.Contains(errStr, "500") ||
			strings.Contains(errStr, "429") || // Rate limit
			strings.Contains(errStr, "context deadline exceeded") ||
			strings.Contains(errStr, "connection reset by peer") ||
			strings.Contains(errStr, "upstream HTTP request failed") {

			t.Logf("Attempt %d/%d: Call to themealdb.com failed with a transient/blocking error: %v. Retrying...", i+1, maxRetries, err)
			if i < maxRetries-1 {
				time.Sleep(2 * time.Second) // Wait before retrying
				continue
			} else {
				// If we exhausted retries on transient errors, SKIP instead of fail
				t.Skipf("Skipping test: all %d retries to themealdb.com failed with external errors. Last error: %v", maxRetries, err)
				return
			}
		}

		require.NoError(t, err, "unrecoverable error calling searchMeal tool")
	}

	require.NoError(t, err, "Error calling searchMeal tool")
	require.NotNil(t, res, "Nil response from searchMeal tool")

	// --- 4. Assert Response ---
	require.Len(t, res.Content, 1, "Expected exactly one content item")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content")

	var theMealDBResponse map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &theMealDBResponse)
	require.NoError(t, err, "Failed to unmarshal JSON response")

	if _, ok := theMealDBResponse["meals"].(string); ok {
		// t.Skip("Skipping test, no meals found in response")
	}
	if theMealDBResponse["meals"] == nil {
		// t.Skip("Skipping test, no meals found in response")
	}
	meals, ok := theMealDBResponse["meals"].([]interface{})
	require.True(t, ok, "The meals should be an array")
	require.True(t, len(meals) > 0, "The response should contain at least one meal")

	t.Logf("SUCCESS: Received correct data for arrabiata: %s", textContent.Text)

	t.Log("INFO: E2E Test Scenario for TheMealDB Server Completed Successfully!")
}
