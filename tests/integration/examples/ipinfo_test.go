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

package examples

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

func TestUpstreamService_IPInfo(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for IP Info Server...")

	// --- 1. Start MCPXY Server ---
	mcpxTestServerInfo := integration.StartMCPXYServer(t, "E2EIPInfoServerTest")
	defer mcpxTestServerInfo.CleanupFunc()

	// --- 2. Register IP Info Service with MCPXY ---
	const serviceID = "e2e_ipinfo"
	serviceURL := "http://ip-api.com"
	endpointPath := "/json/{{ip}}"
	operationID := "getIPInfo"

	t.Logf("INFO: Registering '%s' with MCPXY at endpoint %s...", serviceID, serviceURL)
	registrationGRPCClient := mcpxTestServerInfo.RegistrationClient

	method := configv1.HttpCallDefinition_HTTP_METHOD_GET
	param := configv1.HttpParameterMapping_builder{
		Name: proto.String("ip"),
	}.Build()

	callDef := configv1.HttpCallDefinition_builder{
		OperationId:  proto.String(operationID),
		EndpointPath: proto.String(endpointPath),
		Method:       &method,
		Parameters:   []*configv1.HttpParameterMapping{param},
	}.Build()

	httpService := configv1.HttpUpstreamService_builder{
		Address: proto.String(serviceURL),
		Calls:   []*configv1.HttpCallDefinition{callDef},
	}.Build()

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name:        proto.String(serviceID),
		HttpService: httpService,
	}.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: serviceConfig,
	}.Build()

	resp, err := registrationGRPCClient.RegisterService(context.Background(), req)
	require.NoError(t, err, "Failed to register service via API")
	require.NotNil(t, resp, "Nil response from RegisterService API")
	t.Logf("INFO: '%s' registered.", serviceID)

	// --- 3. Call Tool via MCPXY ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpxTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	serviceKey, _ := util.GenerateID(serviceID)
	toolName, _ := util.GenerateToolID(serviceKey, operationID)
	// Using Google's public DNS for a stable test target
	params := json.RawMessage(`{"ip": "8.8.8.8"}`)
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: params})
	require.NoError(t, err, "Error calling ip-api tool")
	require.NotNil(t, res, "Nil response from ip-api tool")

	require.Len(t, res.Content, 1, "Expected exactly one content block in the response")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content but got %T", res.Content[0])

	t.Logf("Raw response from ip-api: %s", textContent.Text)

	var ipInfoResponse struct {
		Status      string  `json:"status"`
		Country     string  `json:"country"`
		CountryCode string  `json:"countryCode"`
		Region      string  `json:"region"`
		RegionName  string  `json:"regionName"`
		City        string  `json:"city"`
		Zip         string  `json:"zip"`
		Lat         float64 `json:"lat"`
		Lon         float64 `json:"lon"`
		Timezone    string  `json:"timezone"`
		ISP         string  `json:"isp"`
		Org         string  `json:"org"`
		AS          string  `json:"as"`
		Query       string  `json:"query"`
	}

	err = json.Unmarshal([]byte(textContent.Text), &ipInfoResponse)
	require.NoError(t, err, "Failed to unmarshal ip-api response")

	require.Equal(t, "success", ipInfoResponse.Status, "Expected status to be success")
	require.Equal(t, "8.8.8.8", ipInfoResponse.Query, "Expected query to be the requested IP")
	require.Equal(t, "United States", ipInfoResponse.Country, "Expected country to be United States")
	t.Logf("SUCCESS: Received correct IP info for 8.8.8.8: %s", textContent.Text)
	t.Log("INFO: E2E Test Scenario for IP Info Server Completed Successfully!")
}
