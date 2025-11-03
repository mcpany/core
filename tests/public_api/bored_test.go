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
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

//go:build e2e_public_api

package public_api

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/util"
	apiv1 "github.com/mcpany/core/proto/api/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestUpstreamService_Bored(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Bored Server...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpxTestServerInfo := integration.StartMCPANYServer(t, "E2EBoredServerTest")
	defer mcpxTestServerInfo.CleanupFunc()

	// --- 2. Register Bored Server with MCPANY ---
	const boredServiceID = "e2e_bored"
	boredServiceEndpoint := "https://www.boredapi.com"
	t.Logf("INFO: Registering '%s' with MCPANY at endpoint %s...", boredServiceID, boredServiceEndpoint)
	registrationGRPCClient := mcpxTestServerInfo.RegistrationClient

	httpCall := configv1.HttpCallDefinition_builder{
		EndpointPath: proto.String("/api/activity"),
		Schema: configv1.ToolSchema_builder{
			Name: proto.String("getActivity"),
		}.Build(),
		Method: configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HttpMethod_value["HTTP_METHOD_GET"]).Enum(),
	}.Build()

	httpService := configv1.HttpUpstreamService_builder{
		Address: proto.String(boredServiceEndpoint),
		Calls:   []*configv1.HttpCallDefinition{httpCall},
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
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpxTestServerInfo.HTTPEndpoint}, nil)
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

	const maxRetries = 3
	var res *mcp.CallToolResult

	for i := 0; i < maxRetries; i++ {
		res, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(`{}`)} )
		if err == nil {
			break // Success
		}

		if strings.Contains(err.Error(), "503 Service Temporarily Unavailable") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "connection reset by peer") || strings.Contains(err.Error(), "no such host") {
			t.Logf("Attempt %d/%d: Call to boredapi.com failed with a transient error: %v. Retrying...", i+1, maxRetries, err)
			time.Sleep(2 * time.Second) // Wait before retrying
			continue
		}

		require.NoError(t, err, "unrecoverable error calling getActivity tool")
	}

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
	require.NotEmpty(t, boredResponse["price"], "The price should not be empty")
	require.NotEmpty(t, boredResponse["link"], "The link should not be empty")
	require.NotEmpty(t, boredResponse["key"], "The key should not be empty")
	require.NotEmpty(t, boredResponse["accessibility"], "The accessibility should not be empty")
	t.Logf("SUCCESS: Received an activity: %s", textContent.Text)

	t.Log("INFO: E2E Test Scenario for Bored Server Completed Successfully!")
}
