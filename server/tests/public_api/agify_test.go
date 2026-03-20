// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e_public_api

package public_api

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mcpany/core/server/pkg/util"
	apiv1 "github.com/mcpany/core/proto/api/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestUpstreamService_Agify(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Agify Server...")
	t.Parallel()

	// --- 1. Start Mock Server ---
	mockResponse := `{"name": "michael", "age": 50, "count": 100}`
	mockServer := integration.CreateMockServerWithResponses(t, map[string]string{
		"/?name=michael": mockResponse,
	})
	defer mockServer.Close()

	// --- 2. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EAgifyServerTest")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 3. Register Agify Server with MCPANY ---
	const agifyServiceID = "e2e_agify"
	agifyServiceEndpoint := mockServer.URL
	t.Logf("INFO: Registering '%s' with MCPANY at endpoint %s...", agifyServiceID, agifyServiceEndpoint)
	registrationGRPCClient := mcpAnyTestServerInfo.RegistrationClient

	callID := "getAge"
	httpCall := &configv1.HttpCallDefinition{
		Id:           proto.String(callID),
		EndpointPath: proto.String("/"),
		Method:       configv1.HttpCallDefinition_HTTP_METHOD_GET.Enum(),
		Parameters: []*configv1.HttpParameterMapping{
			{
				Schema: &configv1.ParameterSchema{
					Name: proto.String("name"),
					Type: configv1.ParameterType_STRING.Enum(),
				},
				In: configv1.HttpParameterMapping_QUERY.Enum(),
			},
		},
	}

	toolDef := &configv1.ToolDefinition{
		Name:   proto.String("getAge"),
		CallId: proto.String(callID),
	}

	httpService := &configv1.HttpUpstreamService{
		Address: proto.String(agifyServiceEndpoint),
		Tools:   []*configv1.ToolDefinition{toolDef},
		Calls:   map[string]*configv1.HttpCallDefinition{callID: httpCall},
	}

	config := &configv1.UpstreamServiceConfig{
		Id:          proto.String(agifyServiceID),
		Name:        proto.String(agifyServiceID),
		HttpService: httpService,
	}

	req := &apiv1.RegisterServiceRequest{
		Config: config,
	}

	integration.RegisterServiceViaAPI(t, registrationGRPCClient, req)
	t.Logf("INFO: '%s' registered.", agifyServiceID)

	// --- 4. Call Tool via MCPANY ---
	client, session, cleanup := integration.ConnectToMCPServer(t, ctx, mcpAnyTestServerInfo.MCPAddress, mcpAnyTestServerInfo.APIKey)
	defer cleanup()
	defer session.Close()

	listToolsResult, err := client.ListTools(ctx, mcp.ListToolsRequest{})
	require.NoError(t, err)
	for _, tool := range listToolsResult.Tools {
		t.Logf("Discovered tool from MCPANY: %s", tool.Name)
	}

	serviceID, _ := util.SanitizeServiceName(agifyServiceID)
	sanitizedToolName, _ := util.SanitizeToolName("getAge")
	toolName := serviceID + "." + sanitizedToolName

	res, err := client.CallTool(ctx, mcp.CallToolRequest{
		Params: struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments,omitempty"`
		}{
			Name: toolName,
			Arguments: map[string]any{
				"name": "michael",
			},
		},
	})
	require.NoError(t, err, "Error calling getAge tool")
	require.NotNil(t, res, "Nil response from getAge tool")

	// --- 5. Assert Response ---
	require.Len(t, res.Content, 1, "Expected exactly one content item")
	textContent, ok := res.Content[0].(mcp.TextContent)
	require.True(t, ok, "Expected text content")

	var agifyResponse map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &agifyResponse)
	require.NoError(t, err, "Failed to unmarshal JSON response")

	require.Equal(t, "michael", agifyResponse["name"], "The name does not match")
	require.NotEmpty(t, agifyResponse["age"], "The age should not be empty")
	require.NotEmpty(t, agifyResponse["count"], "The count should not be empty")
	t.Logf("SUCCESS: Received correct age for michael: %s", textContent.Text)

	t.Log("INFO: E2E Test Scenario for Agify Server Completed Successfully!")
}
