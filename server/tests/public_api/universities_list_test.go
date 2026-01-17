// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e_public_api

package public_api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestUpstreamService_UniversitiesList(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Universities List Server...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EUniversitiesListServerTest")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 2. Register Universities List Server with MCPANY ---
	const universitiesListServiceID = "e2e_universitieslist"
	universitiesListServiceEndpoint := "http://universities.hipolabs.com"
	t.Logf("INFO: Registering '%s' with MCPANY at endpoint %s...", universitiesListServiceID, universitiesListServiceEndpoint)
	registrationGRPCClient := mcpAnyTestServerInfo.RegistrationClient

	callID := "getUniversities"
	httpCall := configv1.HttpCallDefinition_builder{
		Id:           proto.String(callID),
		EndpointPath: proto.String("/search?country={{country}}"),
		Method:       configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HttpMethod_value["HTTP_METHOD_GET"]).Enum(),
		Parameters: []*configv1.HttpParameterMapping{
			configv1.HttpParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("country"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	toolDef := configv1.ToolDefinition_builder{
		Name:   proto.String("getUniversities"),
		CallId: proto.String(callID),
	}.Build()

	httpService := configv1.HttpUpstreamService_builder{
		Address: proto.String(universitiesListServiceEndpoint),
		Tools:   []*configv1.ToolDefinition{toolDef},
		Calls:   map[string]*configv1.HttpCallDefinition{callID: httpCall},
	}.Build()

	config := configv1.UpstreamServiceConfig_builder{
		Name:        proto.String(universitiesListServiceID),
		HttpService: httpService,
	}.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	integration.RegisterServiceViaAPI(t, registrationGRPCClient, req)
	t.Logf("INFO: '%s' registered.", universitiesListServiceID)

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

	serviceID, _ := util.SanitizeServiceName(universitiesListServiceID)
	sanitizedToolName, _ := util.SanitizeToolName("getUniversities")
	toolName := serviceID + "." + sanitizedToolName
	country := `{"country": "United States"}`

	const maxRetries = 3
	var res *mcp.CallToolResult

	for i := 0; i < maxRetries; i++ {
		res, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(country)})
		if err == nil && res != nil && res.IsError {
			// Convert to error for retry check
			if len(res.Content) > 0 {
				if txt, ok := res.Content[0].(*mcp.TextContent); ok {
					err = fmt.Errorf("tool returned error: %s", txt.Text)
				} else {
					err = fmt.Errorf("tool returned error result")
				}
			} else {
				err = fmt.Errorf("tool returned error result with no content")
			}
		}

		if err == nil {
			break // Success
		}

		if strings.Contains(err.Error(), "503 Service Temporarily Unavailable") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "connection reset by peer") || strings.Contains(err.Error(), "unexpected EOF") || strings.Contains(err.Error(), "tool returned error") {
			t.Logf("Attempt %d/%d: Call to universities.hipolabs.com failed with a transient error: %v. Retrying...", i+1, maxRetries, err)
			time.Sleep(2 * time.Second) // Wait before retrying
			continue
		}

		require.NoError(t, err, "unrecoverable error calling getUniversities tool")
	}

	if err != nil {
		t.Skipf("Skipping test: all %d retries to universities.hipolabs.com failed with transient errors. Last error: %v", maxRetries, err)
	}

	require.NoError(t, err, "Error calling getUniversities tool")
	require.NotNil(t, res, "Nil response from getUniversities tool")

	// --- 4. Assert Response ---
	require.Len(t, res.Content, 1, "Expected exactly one content item")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content")

	var universitiesListResponse []map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &universitiesListResponse)
	require.NoError(t, err, "Failed to unmarshal JSON response")

	require.True(t, len(universitiesListResponse) > 0, "The response should contain at least one university")
	require.Equal(t, "United States", universitiesListResponse[0]["country"], "The country does not match")
	t.Logf("SUCCESS: Received correct data for universities in the United States: %s", textContent.Text)

	t.Log("INFO: E2E Test Scenario for Universities List Server Completed Successfully!")
}
