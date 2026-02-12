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

func TestUpstreamService_PokeAPI(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeMedium)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for PokeAPI Server...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EPokeAPIServerTest")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 2. Register PokeAPI Server with MCPANY ---
	const pokeAPIServiceID = "e2e_pokeapi"
	pokeAPIServiceEndpoint := "https://pokeapi.co"
	t.Logf("INFO: Registering '%s' with MCPANY at endpoint %s...", pokeAPIServiceID, pokeAPIServiceEndpoint)
	registrationGRPCClient := mcpAnyTestServerInfo.RegistrationClient

	callID := "getPokemon"
	httpCall := configv1.HttpCallDefinition_builder{
		Id:           proto.String(callID),
		EndpointPath: proto.String("/api/v2/pokemon/{{name}}"),
		Method:       configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HttpMethod_value["HTTP_METHOD_GET"]).Enum(),
		Parameters: []*configv1.HttpParameterMapping{
			configv1.HttpParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("name"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	toolDef := configv1.ToolDefinition_builder{
		Name:   proto.String("getPokemon"),
		CallId: proto.String(callID),
	}.Build()

	httpService := configv1.HttpUpstreamService_builder{
		Address: proto.String(pokeAPIServiceEndpoint),
		Tools:   []*configv1.ToolDefinition{toolDef},
		Calls:   map[string]*configv1.HttpCallDefinition{callID: httpCall},
	}.Build()

	config := configv1.UpstreamServiceConfig_builder{
		Name:        proto.String(pokeAPIServiceID),
		HttpService: httpService,
	}.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	integration.RegisterServiceViaAPI(t, registrationGRPCClient, req)
	t.Logf("INFO: '%s' registered.", pokeAPIServiceID)

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

	serviceID, _ := util.SanitizeServiceName(pokeAPIServiceID)
	sanitizedToolName, _ := util.SanitizeToolName("getPokemon")
	toolName := serviceID + "." + sanitizedToolName
	name := `{"name": "ditto"}`

	const maxRetries = 3
	var res *mcp.CallToolResult

	for i := 0; i < maxRetries; i++ {
		res, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(name)})
		if err == nil {
			break // Success
		}

		if strings.Contains(err.Error(), "503 Service Temporarily Unavailable") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "connection reset by peer") {
			t.Logf("Attempt %d/%d: Call to pokeapi.co failed with a transient error: %v. Retrying...", i+1, maxRetries, err)
			time.Sleep(2 * time.Second) // Wait before retrying
			continue
		}

		require.NoError(t, err, "unrecoverable error calling getPokemon tool")
	}

	if err != nil {
		// t.Skipf("Skipping test: all %d retries to pokeapi.co failed with transient errors. Last error: %v", maxRetries, err)
	}

	require.NoError(t, err, "Error calling getPokemon tool")
	require.NotNil(t, res, "Nil response from getPokemon tool")

	// --- 4. Assert Response ---
	require.Len(t, res.Content, 1, "Expected exactly one content item")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content")

	var pokeAPIResponse map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &pokeAPIResponse)
	require.NoError(t, err, "Failed to unmarshal JSON response")

	require.Equal(t, "ditto", pokeAPIResponse["name"], "The name does not match")
	require.NotEmpty(t, pokeAPIResponse["abilities"], "The abilities should not be empty")
	t.Logf("SUCCESS: Received correct data for ditto: %s", textContent.Text)

	t.Log("INFO: E2E Test Scenario for PokeAPI Server Completed Successfully!")
}

func TestUpstreamService_PokeAPI_ParameterSubstitution(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeMedium)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for PokeAPI Server...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EPokeAPIServerTest")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 2. Register PokeAPI Server with MCPANY ---
	const pokeAPIServiceID = "e2e_pokeapi_param_subst"
	pokeAPIServiceEndpoint := "https://pokeapi.co"
	t.Logf("INFO: Registering '%s' with MCPANY at endpoint %s...", pokeAPIServiceID, pokeAPIServiceEndpoint)
	registrationGRPCClient := mcpAnyTestServerInfo.RegistrationClient

	callID := "getPokemon"
	httpCall := configv1.HttpCallDefinition_builder{
		Id:           proto.String(callID),
		EndpointPath: proto.String("/api/v2/pokemon/{{name}}"),
		Method:       configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HttpMethod_value["HTTP_METHOD_GET"]).Enum(),
		Parameters: []*configv1.HttpParameterMapping{
			configv1.HttpParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("name"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	toolDef := configv1.ToolDefinition_builder{
		Name:   proto.String("getPokemon"),
		CallId: proto.String(callID),
	}.Build()

	httpService := configv1.HttpUpstreamService_builder{
		Address: proto.String(pokeAPIServiceEndpoint),
		Tools:   []*configv1.ToolDefinition{toolDef},
		Calls:   map[string]*configv1.HttpCallDefinition{callID: httpCall},
	}.Build()

	config := configv1.UpstreamServiceConfig_builder{
		Name:        proto.String(pokeAPIServiceID),
		HttpService: httpService,
	}.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	integration.RegisterServiceViaAPI(t, registrationGRPCClient, req)
	t.Logf("INFO: '%s' registered.", pokeAPIServiceID)

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

	serviceID, _ := util.SanitizeServiceName(pokeAPIServiceID)
	sanitizedToolName, _ := util.SanitizeToolName("getPokemon")
	toolName := serviceID + "." + sanitizedToolName
	pokemonName := "pikachu"
	name := `{"name": "` + pokemonName + `"}`

	const maxRetries = 3
	var res *mcp.CallToolResult

	for i := 0; i < maxRetries; i++ {
		res, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(name)})
		if err == nil {
			break // Success
		}

		if strings.Contains(err.Error(), "503 Service Temporarily Unavailable") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "connection reset by peer") {
			t.Logf("Attempt %d/%d: Call to pokeapi.co failed with a transient error: %v. Retrying...", i+1, maxRetries, err)
			time.Sleep(2 * time.Second) // Wait before retrying
			continue
		}

		require.NoError(t, err, "unrecoverable error calling getPokemon tool")
	}

	if err != nil {
		// t.Skipf("Skipping test: all %d retries to pokeapi.co failed with transient errors. Last error: %v", maxRetries, err)
	}

	require.NoError(t, err, "Error calling getPokemon tool")
	require.NotNil(t, res, "Nil response from getPokemon tool")

	// --- 4. Assert Response ---
	require.Len(t, res.Content, 1, "Expected exactly one content item")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content")

	var pokeAPIResponse map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &pokeAPIResponse)
	require.NoError(t, err, "Failed to unmarshal JSON response")

	require.Equal(t, pokemonName, pokeAPIResponse["name"], "The name does not match")
	t.Logf("SUCCESS: Received correct data for %s: %s", pokemonName, textContent.Text)

	t.Log("INFO: E2E Test Scenario for PokeAPI Server Parameter Substitution Completed Successfully!")
}
