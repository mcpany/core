/*
 * Copyright 2025 Author(s) of MCP-XY
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
	"testing"

	"github.com/mcpxy/core/pkg/util"
	apiv1 "github.com/mcpxy/core/proto/api/v1"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	"github.com/mcpxy/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestUpstreamService_JsonPlaceholder(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for JSONPlaceholder Server...")
	t.Parallel()

	// --- 1. Start MCPXY Server ---
	mcpxTestServerInfo := integration.StartMCPXYServer(t, "E2EJsonPlaceholderServerTest")
	defer mcpxTestServerInfo.CleanupFunc()

	// --- 2. Register JSONPlaceholder Server with MCPXY ---
	const serviceID = "e2e_jsonplaceholder"
	serviceEndpoint := "https://jsonplaceholder.typicode.com"
	t.Logf("INFO: Registering '%s' with MCPXY at endpoint %s...", serviceID, serviceEndpoint)
	registrationGRPCClient := mcpxTestServerInfo.RegistrationClient

	httpCall := configv1.HttpCallDefinition_builder{
		EndpointPath: proto.String("/posts"),
		Annotation: configv1.ToolAnnotation_builder{
			Name: proto.String("getPosts"),
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

	// --- 3. Call Tool via MCPXY ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpxTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)
	for _, tool := range listToolsResult.Tools {
		t.Logf("Discovered tool from MCPXY: %s", tool.Name)
	}

	serviceKey, _ := util.GenerateID(serviceID)
	toolName, _ := util.GenerateToolID(serviceKey, "getPosts")

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(`{}`)})
	require.NoError(t, err, "Error calling getPosts tool")
	require.NotNil(t, res, "Nil response from getPosts tool")

	// --- 4. Assert Response ---
	require.Len(t, res.Content, 1, "Expected exactly one content item")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content")

	var postsResponse []map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &postsResponse)
	require.NoError(t, err, "Failed to unmarshal JSON response")
	require.Greater(t, len(postsResponse), 0, "Expected at least one post in the response")
	require.NotNil(t, postsResponse[0]["title"], "The title should not be nil")
	t.Logf("SUCCESS: Received posts")

	t.Log("INFO: E2E Test Scenario for JSONPlaceholder Server Completed Successfully!")
}
