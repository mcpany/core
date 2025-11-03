//go:build e2e_public_api

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

func TestUpstreamService_DogFacts(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Dog Facts Server...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpxTestServerInfo := integration.StartMCPANYServer(t, "E2EDogFactsServerTest")
	defer mcpxTestServerInfo.CleanupFunc()

	// --- 2. Register Dog Facts Server with MCPANY ---
	const dogFactsServiceID = "e2e_dogfacts"
	dogFactsServiceEndpoint := "https://dog-api.kinduff.com"
	t.Logf("INFO: Registering '%s' with MCPANY at endpoint %s...", dogFactsServiceID, dogFactsServiceEndpoint)
	registrationGRPCClient := mcpxTestServerInfo.RegistrationClient

	httpCall := configv1.HttpCallDefinition_builder{
		EndpointPath: proto.String("/api/facts"),
		Schema: configv1.ToolSchema_builder{
			Name: proto.String("getDogFact"),
		}.Build(),
		Method: configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HttpMethod_value["HTTP_METHOD_GET"]).Enum(),
	}.Build()

	httpService := configv1.HttpUpstreamService_builder{
		Address: proto.String(dogFactsServiceEndpoint),
		Calls:   []*configv1.HttpCallDefinition{httpCall},
	}.Build()

	config := configv1.UpstreamServiceConfig_builder{
		Name:        proto.String(dogFactsServiceID),
		HttpService: httpService,
	}.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	integration.RegisterServiceViaAPI(t, registrationGRPCClient, req)
	t.Logf("INFO: '%s' registered.", dogFactsServiceID)

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

	serviceID, _ := util.SanitizeServiceName(dogFactsServiceID)
	sanitizedToolName, _ := util.SanitizeToolName("getDogFact")
	toolName := serviceID + "." + sanitizedToolName

	const maxRetries = 3
	var res *mcp.CallToolResult

	for i := 0; i < maxRetries; i++ {
		res, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(`{}`)})
		if err == nil {
			break // Success
		}

		if strings.Contains(err.Error(), "503 Service Temporarily Unavailable") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "connection reset by peer") {
			t.Logf("Attempt %d/%d: Call to dog-api.kinduff.com failed with a transient error: %v. Retrying...", i+1, maxRetries, err)
			time.Sleep(2 * time.Second) // Wait before retrying
			continue
		}

		require.NoError(t, err, "unrecoverable error calling getDogFact tool")
	}

	require.NoError(t, err, "Error calling getDogFact tool")
	require.NotNil(t, res, "Nil response from getDogFact tool")

	// --- 4. Assert Response ---
	require.Len(t, res.Content, 1, "Expected exactly one content item")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content")

	var dogFactResponse map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &dogFactResponse)
	require.NoError(t, err, "Failed to unmarshal JSON response")

	require.NotNil(t, dogFactResponse["facts"], "The facts should not be nil")
	require.Equal(t, true, dogFactResponse["success"], "The success should be true")
	t.Logf("SUCCESS: Received a dog fact: %s", textContent.Text)

	t.Log("INFO: E2E Test Scenario for Dog Facts Server Completed Successfully!")
}
