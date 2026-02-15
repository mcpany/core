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

		// Handle JSON unmarshal errors in the response (e.g. HTML error pages from upstream)
		if err == nil {
			// Validate content is JSON
			if len(res.Content) > 0 {
				if textContent, ok := res.Content[0].(*mcp.TextContent); ok {
					var js map[string]interface{}
					if jsonErr := json.Unmarshal([]byte(textContent.Text), &js); jsonErr != nil {
						err = jsonErr // Treat as error to trigger retry
					}
				}
			}
		}

		if err == nil {
			break // Success
		}

		if strings.Contains(err.Error(), "503 Service Temporarily Unavailable") ||
		   strings.Contains(err.Error(), "context deadline exceeded") ||
		   strings.Contains(err.Error(), "connection reset by peer") ||
		   strings.Contains(err.Error(), "invalid character") { // JSON error
			t.Logf("Attempt %d/%d: Call to themealdb.com failed with a transient error: %v. Retrying...", i+1, maxRetries, err)
			time.Sleep(2 * time.Second) // Wait before retrying
			continue
		}

		require.NoError(t, err, "unrecoverable error calling searchMeal tool")
	}

	if err != nil {
		t.Skipf("Skipping test: all %d retries to themealdb.com failed with transient errors. Last error: %v", maxRetries, err)
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
