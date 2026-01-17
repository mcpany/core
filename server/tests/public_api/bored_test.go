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

func TestUpstreamService_Bored(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Bored Server...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EBoredServerTest")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 2. Register Bored Server with MCPANY ---
	const boredServiceID = "e2e_bored"
	boredServiceEndpoint := "https://www.boredapi.com"
	t.Logf("INFO: Registering '%s' with MCPANY at endpoint %s...", boredServiceID, boredServiceEndpoint)
	registrationGRPCClient := mcpAnyTestServerInfo.RegistrationClient

	callID := "getActivity"
	httpCall := configv1.HttpCallDefinition_builder{
		Id:           proto.String(callID),
		EndpointPath: proto.String("/api/activity"),
		Method:       configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HttpMethod_value["HTTP_METHOD_GET"]).Enum(),
	}.Build()

	toolDef := configv1.ToolDefinition_builder{
		Name:   proto.String("getActivity"),
		CallId: proto.String(callID),
	}.Build()

	httpService := configv1.HttpUpstreamService_builder{
		Address: proto.String(boredServiceEndpoint),
		Tools:   []*configv1.ToolDefinition{toolDef},
		Calls:   map[string]*configv1.HttpCallDefinition{callID: httpCall},
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
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpAnyTestServerInfo.HTTPEndpoint}, nil)
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
		res, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(`{}`)})
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

		if strings.Contains(err.Error(), "503 Service Temporarily Unavailable") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "connection reset by peer") || strings.Contains(err.Error(), "no such host") || strings.Contains(err.Error(), "tool returned error") {
			t.Logf("Attempt %d/%d: Call to boredapi.com failed with a transient error: %v. Retrying...", i+1, maxRetries, err)
			time.Sleep(2 * time.Second) // Wait before retrying
			continue
		}

		require.NoError(t, err, "unrecoverable error calling getActivity tool")
	}

	if err != nil {
		t.Skipf("Skipping test: all %d retries to boredapi.com failed with transient errors. Last error: %v", maxRetries, err)
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
	if err != nil {
		// t.Skipf("Skipping test due to transient error from boredapi.com: %v", err)
	}
	t.Logf("SUCCESS: Received an activity: %s", textContent.Text)

	t.Log("INFO: E2E Test Scenario for Bored Server Completed Successfully!")
}
