// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e_public_api

package public_api

import (
	"context"
	"encoding/json"
	"testing"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestUpstreamService_Bored(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Bored Server...")
	t.Parallel()

	// --- 1. Start Mock Server ---
	mockResponse := `{"activity":"Learn a new language","type":"education","participants":1,"price":0.1,"link":"","key":"5881028","accessibility":0.25}`
	mockServer := integration.CreateMockServerWithResponses(t, map[string]string{
		"/api/activity/": mockResponse,
		"/api/activity": mockResponse,
	})
	defer mockServer.Close()

	// --- 2. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EBoredServerTest")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 3. Register Bored Server with MCPANY ---
	const boredServiceID = "e2e_bored"
	boredServiceEndpoint := mockServer.URL
	t.Logf("INFO: Registering '%s' with MCPANY at endpoint %s...", boredServiceID, boredServiceEndpoint)
	registrationGRPCClient := mcpAnyTestServerInfo.RegistrationClient

	callID := "getActivity"
	httpCall := configv1.HttpCallDefinition_builder{
		Id:           proto.String(callID),
		EndpointPath: proto.String("/api/activity"),
		Method:       configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HttpMethod_value["HTTP_METHOD_GET"]).Enum(),
	}.Build()

	toolDef := configv1.ToolDefinition_builder{
		Name:   proto.String("getActivity"),
		CallId: proto.String(callID),
	}.Build()

	httpService := configv1.HttpUpstreamService_builder{
		Address: proto.String(boredServiceEndpoint),
		Tools:   []*configv1.ToolDefinition{toolDef},
		Calls:   map[string]*configv1.HttpCallDefinition{callID: httpCall},
	}.Build()

	config := configv1.UpstreamServiceConfig_builder{
		Name:        proto.String(boredServiceID),
		HttpService: httpService,
	}.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	integration.RegisterServiceViaAPI(t, registrationGRPCClient, req)
	t.Logf("INFO: '%s' registered.", boredServiceID)

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

	serviceID, _ := util.SanitizeServiceName(boredServiceID)
	sanitizedToolName, _ := util.SanitizeToolName("getActivity")
	toolName := serviceID + "." + sanitizedToolName

	// Call the tool directly (mock server is fast, no retries needed)
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(`{}`)})
	require.NoError(t, err, "Error calling getActivity tool")
	require.NotNil(t, res, "Nil response from getActivity tool")

	// --- 4. Assert Response ---
	require.Len(t, res.Content, 1, "Expected exactly one content item")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content")

	var boredResponse map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &boredResponse)
	require.NoError(t, err, "Failed to unmarshal JSON response")

	require.NotEmpty(t, boredResponse["activity"], "The activity should not be empty")
	require.NotEmpty(t, boredResponse["type"], "The type should not be empty")
	require.NotEmpty(t, boredResponse["participants"], "The participants should not be empty")
	require.NotNil(t, boredResponse["price"], "The price should not be nil")
	// The link can be empty strings for some activities. Use NotNil instead.
	require.NotNil(t, boredResponse["link"], "The link should not be nil")
	require.NotEmpty(t, boredResponse["key"], "The key should not be empty")
	// require.NotEmpty(t, boredResponse["accessibility"], "The accessibility should not be empty") // Accessibility can be 0, which is empty?
	// mock returns 0.25 so it's not empty string/nil. NotEmpty works for float? Yes.
	t.Logf("SUCCESS: Received an activity: %s", textContent.Text)

	t.Log("INFO: E2E Test Scenario for Bored Server Completed Successfully!")
}
