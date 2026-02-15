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

func TestUpstreamService_OpenBreweryDB(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Open Brewery DB Server...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EOpenBreweryDBServerTest")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 2. Register Open Brewery DB Server with MCPANY ---
	const openBreweryDBServiceID = "e2e_openbrewerydb"
	openBreweryDBServiceEndpoint := "https://api.openbrewerydb.org/v1"
	t.Logf("INFO: Registering '%s' with MCPANY at endpoint %s...", openBreweryDBServiceID, openBreweryDBServiceEndpoint)
	registrationGRPCClient := mcpAnyTestServerInfo.RegistrationClient

	callID := "getBreweries"
	httpCall := configv1.HttpCallDefinition_builder{
		Id:           proto.String(callID),
		EndpointPath: proto.String("/breweries"),
		Method:       configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HttpMethod_value["HTTP_METHOD_GET"]).Enum(),
	}.Build()

	toolDef := configv1.ToolDefinition_builder{
		Name:   proto.String("getBreweries"),
		CallId: proto.String(callID),
	}.Build()

	httpService := configv1.HttpUpstreamService_builder{
		Address: proto.String(openBreweryDBServiceEndpoint),
		Tools:   []*configv1.ToolDefinition{toolDef},
		Calls:   map[string]*configv1.HttpCallDefinition{callID: httpCall},
	}.Build()

	config := configv1.UpstreamServiceConfig_builder{
		Name:        proto.String(openBreweryDBServiceID),
		HttpService: httpService,
	}.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	integration.RegisterServiceViaAPI(t, registrationGRPCClient, req)
	t.Logf("INFO: '%s' registered.", openBreweryDBServiceID)

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

	serviceID, _ := util.SanitizeServiceName(openBreweryDBServiceID)
	sanitizedToolName, _ := util.SanitizeToolName("getBreweries")
	toolName := serviceID + "." + sanitizedToolName

	const maxRetries = 3
	var res *mcp.CallToolResult

	for i := 0; i < maxRetries; i++ {
		res, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(`{}`)})
		if err == nil {
			break // Success
		}

		if strings.Contains(err.Error(), "503 Service Temporarily Unavailable") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "connection reset by peer") {
			t.Logf("Attempt %d/%d: Call to api.openbrewerydb.org failed with a transient error: %v. Retrying...", i+1, maxRetries, err)
			time.Sleep(2 * time.Second) // Wait before retrying
			continue
		}

		require.NoError(t, err, "unrecoverable error calling getBreweries tool")
	}

	if err != nil {
		// t.Skipf("Skipping test: all %d retries to api.openbrewerydb.org failed with transient errors. Last error: %v", maxRetries, err)
	}

	require.NoError(t, err, "Error calling getBreweries tool")
	require.NotNil(t, res, "Nil response from getBreweries tool")

	// --- 4. Assert Response ---
	require.Len(t, res.Content, 1, "Expected exactly one content item")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content")

	var openBreweryDBResponse []map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &openBreweryDBResponse)
	require.NoError(t, err, "Failed to unmarshal JSON response")

	require.True(t, len(openBreweryDBResponse) > 0, "The response should contain at least one brewery")
	t.Logf("SUCCESS: Received correct data for breweries: %s", textContent.Text)

	t.Log("INFO: E2E Test Scenario for Open Brewery DB Server Completed Successfully!")
}
