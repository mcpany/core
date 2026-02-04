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

	apiv1 "github.com/mcpany/core/proto/api/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/mcpany/core/server/tests/framework"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestUpstreamService_Genderize(t *testing.T) {
	// t.SkipNow()
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Genderize Server...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EGenderizeServerTest")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 2. Start Fake Upstream ---
	fakeUpstream := framework.NewFakeUpstream(t, map[string]interface{}{
		"name":        "michael",
		"gender":      "male",
		"probability": 0.99,
		"count":       12345,
	})
	defer fakeUpstream.Close()

	// --- 3. Register Genderize Server with MCPANY ---
	const genderizeServiceID = "e2e_genderize"
	genderizeServiceEndpoint := fakeUpstream.URL
	t.Logf("INFO: Registering '%s' with MCPANY at endpoint %s...", genderizeServiceID, genderizeServiceEndpoint)
	registrationGRPCClient := mcpAnyTestServerInfo.RegistrationClient

	callID := "getGender"
	httpCall := configv1.HttpCallDefinition_builder{
		Id:           proto.String(callID),
		EndpointPath: proto.String("/?name=michael"),
		Method:       configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HttpMethod_value["HTTP_METHOD_GET"]).Enum(),
	}.Build()

	toolDef := configv1.ToolDefinition_builder{
		Name:   proto.String("getGender"),
		CallId: proto.String(callID),
	}.Build()

	httpService := configv1.HttpUpstreamService_builder{
		Address: proto.String(genderizeServiceEndpoint),
		Tools:   []*configv1.ToolDefinition{toolDef},
		Calls:   map[string]*configv1.HttpCallDefinition{callID: httpCall},
	}.Build()

	config := configv1.UpstreamServiceConfig_builder{
		Name:        proto.String(genderizeServiceID),
		HttpService: httpService,
	}.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	integration.RegisterServiceViaAPI(t, registrationGRPCClient, req)
	t.Logf("INFO: '%s' registered.", genderizeServiceID)

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

	serviceID, _ := util.SanitizeServiceName(genderizeServiceID)
	sanitizedToolName, _ := util.SanitizeToolName("getGender")
	toolName := serviceID + "." + sanitizedToolName
	name := `{"name": "michael"}`

	const maxRetries = 3
	var res *mcp.CallToolResult

	for i := 0; i < maxRetries; i++ {
		res, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(name)})
		if err == nil {
			break // Success
		}

		if strings.Contains(err.Error(), "503 Service Temporarily Unavailable") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "connection reset by peer") {
			t.Logf("Attempt %d/%d: Call to api.genderize.io failed with a transient error: %v. Retrying...", i+1, maxRetries, err)
			time.Sleep(2 * time.Second) // Wait before retrying
			continue
		}

		require.NoError(t, err, "unrecoverable error calling getGender tool")
	}

	if err != nil {
		// t.Skipf("Skipping test: all %d retries to api.genderize.io failed with transient errors. Last error: %v", maxRetries, err)
	}

	require.NoError(t, err, "Error calling getGender tool")
	require.NotNil(t, res, "Nil response from getGender tool")

	// --- 4. Assert Response ---
	require.Len(t, res.Content, 1, "Expected exactly one content item")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content")

	var genderizeResponse map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &genderizeResponse)
	require.NoError(t, err, "Failed to unmarshal JSON response")

	require.Equal(t, "michael", genderizeResponse["name"], "The name does not match")
	require.Equal(t, "male", genderizeResponse["gender"], "The gender does not match")
	require.NotEmpty(t, genderizeResponse["probability"], "The probability should not be empty")
	require.NotEmpty(t, genderizeResponse["count"], "The count should not be empty")
	t.Logf("SUCCESS: Received correct gender for michael: %s", textContent.Text)

	t.Log("INFO: E2E Test Scenario for Genderize Server Completed Successfully!")
}
