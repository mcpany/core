/*
 * Copyright 2025 Author(s) of MCPX
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package upstream

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/mcpxy/mcpx/pkg/consts"
	apiv1 "github.com/mcpxy/mcpx/proto/api/v1"
	configv1 "github.com/mcpxy/mcpx/proto/config/v1"
	"github.com/mcpxy/mcpx/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestUpstreamService_Universities(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Universities Server...")
	t.Parallel()

	// --- 1. Start MCPX Server ---
	mcpxTestServerInfo := integration.StartMCPXServer(t, "E2EUniversitiesServerTest")
	defer mcpxTestServerInfo.CleanupFunc()

	// --- 2. Register University Server with MCPX ---
	const universityServiceID = "e2e_universities"
	universityServiceEndpoint := "http://universities.hipolabs.com"
	t.Logf("INFO: Registering '%s' with MCPX at endpoint %s...", universityServiceID, universityServiceEndpoint)
	registrationGRPCClient := mcpxTestServerInfo.RegistrationClient

	httpCall := configv1.HttpCallDefinition_builder{
		EndpointPath: proto.String("/search"),
		OperationId:  proto.String("searchUniversities"),
		Method:       configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HttpMethod_value["HTTP_METHOD_GET"]).Enum(),
		ParameterMappings: []*configv1.HttpParameterMapping{
			configv1.HttpParameterMapping_builder{
				InputParameterName:  proto.String("country"),
				Location:            configv1.HttpParameterMapping_QUERY.Enum(),
				TargetParameterName: proto.String("country"),
			}.Build(),
		},
	}.Build()

	httpService := configv1.HttpUpstreamService_builder{
		Address: proto.String(universityServiceEndpoint),
		Calls:   []*configv1.HttpCallDefinition{httpCall},
	}.Build()

	config := configv1.UpstreamServiceConfig_builder{
		Name:        proto.String(universityServiceID),
		HttpService: httpService,
	}.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	integration.RegisterServiceViaAPI(t, registrationGRPCClient, req)
	t.Logf("INFO: '%s' registered.", universityServiceID)

	// --- 3. Call Tool via MCPX ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpxTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)
	for _, tool := range listToolsResult.Tools {
		t.Logf("Discovered tool from MCPX: %s", tool.Name)
	}

	toolName := fmt.Sprintf("%s%ssearchUniversities", universityServiceID, consts.ToolNameServiceSeparator)
	country := `{"country": "United Kingdom"}`
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(country)})
	require.NoError(t, err, "Error calling searchUniversities tool")
	require.NotNil(t, res, "Nil response from searchUniversities tool")

	// --- 4. Assert Response ---
	require.Len(t, res.Content, 1, "Expected exactly one content item")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content")

	// The upstream API returns a JSON array, which gets wrapped as a JSON string by the server.
	// We need to unmarshal it once to get the raw JSON array string.
	var jsonBodyString string
	err = json.Unmarshal([]byte(textContent.Text), &jsonBodyString)
	if err != nil {
		// It might be that the server is not wrapping the response, and is returning raw JSON.
		// In this case, we unmarshal directly into the target struct.
		t.Log("Failed to unmarshal response into a string, attempting to unmarshal directly into the target slice.")
		var universityResponse []map[string]interface{}
		err = json.Unmarshal([]byte(textContent.Text), &universityResponse)
		require.NoError(t, err, "Failed to unmarshal JSON response directly")
		require.Greater(t, len(universityResponse), 0, "Expected at least one university in the response")
		require.Equal(t, "United Kingdom", universityResponse[0]["country"], "The country does not match")
	} else {
		var universityResponse []map[string]interface{}
		err = json.Unmarshal([]byte(jsonBodyString), &universityResponse)
		require.NoError(t, err, "Failed to unmarshal JSON response from string")
		require.Greater(t, len(universityResponse), 0, "Expected at least one university in the response")
		require.Equal(t, "United Kingdom", universityResponse[0]["country"], "The country does not match")
	}
	t.Logf("SUCCESS: Received correct university info for United Kingdom")

	t.Log("INFO: E2E Test Scenario for Universities Server Completed Successfully!")
}

func TestUpstreamService_UniversitiesByName(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Universities API Server (by name)...")
	t.Parallel()

	// --- 1. Start MCPX Server ---
	mcpxTestServerInfo := integration.StartMCPXServer(t, "E2EUniversitiesByNameServerTest")
	defer mcpxTestServerInfo.CleanupFunc()

	// --- 2. Register University Search Service with MCPX ---
	const universityServiceID = "e2e_universities_by_name"
	universityServiceEndpoint := "http://universities.hipolabs.com"
	t.Logf("INFO: Registering '%s' with MCPX at endpoint %s...", universityServiceID, universityServiceEndpoint)
	registrationGRPCClient := mcpxTestServerInfo.RegistrationClient

	httpCall := configv1.HttpCallDefinition_builder{
		EndpointPath: proto.String("/search"),
		OperationId:  proto.String("searchUniversitiesByName"),
		Method:       configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HttpMethod_value["HTTP_METHOD_GET"]).Enum(),
		ParameterMappings: []*configv1.HttpParameterMapping{
			configv1.HttpParameterMapping_builder{
				InputParameterName:  proto.String("name"),
				Location:            configv1.HttpParameterMapping_QUERY.Enum(),
				TargetParameterName: proto.String("name"),
			}.Build(),
		},
	}.Build()

	httpService := configv1.HttpUpstreamService_builder{
		Address: proto.String(universityServiceEndpoint),
		Calls:   []*configv1.HttpCallDefinition{httpCall},
	}.Build()

	config := configv1.UpstreamServiceConfig_builder{
		Name:        proto.String(universityServiceID),
		HttpService: httpService,
	}.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	integration.RegisterServiceViaAPI(t, registrationGRPCClient, req)
	t.Logf("INFO: '%s' registered.", universityServiceID)

	// --- 3. Call Tool via MCPX ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpxTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)
	for _, tool := range listToolsResult.Tools {
		t.Logf("Discovered tool from MCPX: %s", tool.Name)
	}

	toolName := fmt.Sprintf("%s%ssearchUniversitiesByName", universityServiceID, consts.ToolNameServiceSeparator)
	searchParams := `{"name": "middle"}`
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(searchParams)})
	require.NoError(t, err, "Error calling searchUniversitiesByName tool")
	require.NotNil(t, res, "Nil response from searchUniversitiesByName tool")

	// --- 4. Assert Response ---
	require.Len(t, res.Content, 1, "Expected exactly one content item")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content")

	var universityResponse []map[string]interface{}
	// The response from this API is a JSON array string, so we need to unmarshal twice.
	var jsonBodyString string
	err = json.Unmarshal([]byte(textContent.Text), &jsonBodyString)
	if err != nil {
		// It might be that the server is not wrapping the response, and is returning raw JSON.
		t.Log("Failed to unmarshal response into a string, attempting to unmarshal directly into the target slice.")
		err = json.Unmarshal([]byte(textContent.Text), &universityResponse)
		require.NoError(t, err, "Failed to unmarshal JSON response directly")
	} else {
		err = json.Unmarshal([]byte(jsonBodyString), &universityResponse)
		require.NoError(t, err, "Failed to unmarshal JSON response from string")
	}

	require.NotEmpty(t, universityResponse, "Expected at least one university in the response")
	found := false
	for _, uni := range universityResponse {
		if uni["name"] == "Middlebury College" {
			found = true
			break
		}
	}
	require.True(t, found, "Expected to find 'Middlebury College' in the search results")
	t.Logf("SUCCESS: Received correct university info: %s", textContent.Text)

	t.Log("INFO: E2E Test Scenario for Universities API Server (by name) Completed Successfully!")
}
