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
	"strings"
	"testing"
	"time"

	"github.com/mcpxy/core/pkg/util"
	apiv1 "github.com/mcpxy/core/proto/api/v1"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	"github.com/mcpxy/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestUpstreamService_IPInfo(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for IP Info Server...")
	t.Parallel()

	// --- 1. Start MCPXY Server ---
	mcpxTestServerInfo := integration.StartMCPXYServer(t, "E2EIPInfoServerTest")
	defer mcpxTestServerInfo.CleanupFunc()

	// --- 2. Register IP Info Server with MCPXY ---
	const ipInfoServiceID = "e2e_ipinfo"
	ipInfoServiceEndpoint := "http://ip-api.com"
	t.Logf("INFO: Registering '%s' with MCPXY at endpoint %s...", ipInfoServiceID, ipInfoServiceEndpoint)
	registrationGRPCClient := mcpxTestServerInfo.RegistrationClient

	httpCall := configv1.HttpCallDefinition_builder{
		EndpointPath: proto.String("/json/{{ip}}"),
		Schema: configv1.ToolSchema_builder{
			Name: proto.String("getIPInfo"),
		}.Build(),
		Method: configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HttpMethod_value["HTTP_METHOD_GET"]).Enum(),
		Parameters: []*configv1.HttpParameterMapping{
			configv1.HttpParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("ip"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	httpService := configv1.HttpUpstreamService_builder{
		Address: proto.String(ipInfoServiceEndpoint),
		Calls:   []*configv1.HttpCallDefinition{httpCall},
	}.Build()

	config := configv1.UpstreamServiceConfig_builder{
		Name:        proto.String(ipInfoServiceID),
		HttpService: httpService,
	}.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	integration.RegisterServiceViaAPI(t, registrationGRPCClient, req)
	t.Logf("INFO: '%s' registered.", ipInfoServiceID)

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

	serviceID, _ := util.SanitizeServiceName(ipInfoServiceID)
	sanitizedToolName, _ := util.SanitizeToolName("getIPInfo")
	toolName := serviceID + "." + sanitizedToolName
	ipAddress := `{"ip": "8.8.8.8"}`

	const maxRetries = 3
	var res *mcp.CallToolResult

	for i := 0; i < maxRetries; i++ {
		res, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(ipAddress)})
		if err == nil {
			break // Success
		}

		// If the error is a 503 or a timeout, we can retry. Otherwise, fail fast.
		if strings.Contains(err.Error(), "503 Service Temporarily Unavailable") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "connection reset by peer") {
			t.Logf("Attempt %d/%d: Call to ip-api.com failed with a transient error: %v. Retrying...", i+1, maxRetries, err)
			time.Sleep(2 * time.Second) // Wait before retrying
			continue
		}

		// For any other error, fail the test immediately.
		require.NoError(t, err, "unrecoverable error calling getIPInfo tool")
	}

	if err != nil {
		t.Skipf("Skipping test: all %d retries to ip-api.com failed with transient errors. Last error: %v", maxRetries, err)
	}

	require.NoError(t, err, "Error calling getIPInfo tool")
	require.NotNil(t, res, "Nil response from getIPInfo tool")

	// --- 4. Assert Response ---
	require.Len(t, res.Content, 1, "Expected exactly one content item")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content")

	var ipInfoResponse map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &ipInfoResponse)
	require.NoError(t, err, "Failed to unmarshal JSON response")

	require.Equal(t, "8.8.8.8", ipInfoResponse["query"], "The query IP does not match")
	require.Equal(t, "success", ipInfoResponse["status"], "The status is not success")
	t.Logf("SUCCESS: Received correct IP info for 8.8.8.8: %s", textContent.Text)

	t.Log("INFO: E2E Test Scenario for IP Info Server Completed Successfully!")
}
