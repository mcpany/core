/*
 * Copyright 2025 Author(s) of MCP Any
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

//go:build e2e_public_api

package public_api

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mcpany/core/pkg/util"
	apiv1 "github.com/mcpany/core/proto/api/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestUpstreamService_CatFacts(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Cat Facts Server...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpxTestServerInfo := integration.StartMCPANYServer(t, "E2ECatFactsServerTest")
	defer mcpxTestServerInfo.CleanupFunc()

	// --- 2. Register Cat Facts Server with MCPANY ---
	const serviceID = "e2e_catfacts"
	serviceEndpoint := "https://catfact.ninja"
	t.Logf("INFO: Registering '%s' with MCPANY at endpoint %s...", serviceID, serviceEndpoint)
	registrationGRPCClient := mcpxTestServerInfo.RegistrationClient

	httpCall := configv1.HttpCallDefinition_builder{
		EndpointPath: proto.String("/fact"),
		Schema: configv1.ToolSchema_builder{
			Name: proto.String("getFact"),
		}.Build(),
		Method: configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HttpMethod_value["HTTP_METHOD_GET"]).Enum(),
	}.Build()

	httpService := configv1.HttpUpstreamService_builder{
		Address: proto.String(serviceEndpoint),
		Calls:   []*configv1.HttpCallDefinition{httpCall},
	}.Build()

	config := configv1.UpstreamServiceConfig_builder{
		Name:        proto.String(serviceID),
		HttpService: httpService,
	}.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	integration.RegisterServiceViaAPI(t, registrationGRPCClient, req)
	t.Logf("INFO: '%s' registered.", serviceID)

	// --- 3. Call Tool via MCPANY ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpxTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)
	for _, tool := range listToolsResult.Tools {
		t.Logf("Discovered tool from MCPANY: %s", tool.Name)
	}

	sanitizedServiceID, _ := util.SanitizeServiceName(serviceID)
	sanitizedToolName, _ := util.SanitizeToolName("getFact")
	toolName := sanitizedServiceID + "." + sanitizedToolName

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(`{}`)})
	require.NoError(t, err, "Error calling getFact tool")
	require.NotNil(t, res, "Nil response from getFact tool")

	// --- 4. Assert Response ---
	require.Len(t, res.Content, 1, "Expected exactly one content item")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content")

	var factResponse map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &factResponse)
	require.NoError(t, err, "Failed to unmarshal JSON response")
	require.NotEmpty(t, factResponse["fact"], "The fact should not be empty")
	t.Logf("SUCCESS: Received fact: %s", factResponse["fact"])

	t.Log("INFO: E2E Test Scenario for Cat Facts Server Completed Successfully!")
}
