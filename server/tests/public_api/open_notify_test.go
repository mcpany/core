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
	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestUpstreamService_OpenNotify(t *testing.T) {
	// Increased timeout to handle flaky public API connection
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeMedium)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Open Notify Server...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EOpenNotifyServerTest")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 2. Register Open Notify Server with MCPANY ---
	const openNotifyServiceID = "e2e_opennnotify"
	openNotifyServiceEndpoint := "http://api.open-notify.org"
	t.Logf("INFO: Registering '%s' with MCPANY at endpoint %s...", openNotifyServiceID, openNotifyServiceEndpoint)
	registrationGRPCClient := mcpAnyTestServerInfo.RegistrationClient

	callID := "getAstronauts"
	httpCall := configv1.HttpCallDefinition_builder{
		Id:           proto.String(callID),
		EndpointPath: proto.String("/astros.json"),
		Method:       configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HttpMethod_value["HTTP_METHOD_GET"]).Enum(),
	}.Build()

	toolDef := configv1.ToolDefinition_builder{
		Name:   proto.String("getAstronauts"),
		CallId: proto.String(callID),
	}.Build()

	httpService := configv1.HttpUpstreamService_builder{
		Address: proto.String(openNotifyServiceEndpoint),
		Tools:   []*configv1.ToolDefinition{toolDef},
		Calls:   map[string]*configv1.HttpCallDefinition{callID: httpCall},
	}.Build()

	config := configv1.UpstreamServiceConfig_builder{
		Name:        proto.String(openNotifyServiceID),
		HttpService: httpService,
	}.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	integration.RegisterServiceViaAPI(t, registrationGRPCClient, req)
	t.Logf("INFO: '%s' registered.", openNotifyServiceID)

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

	serviceID, _ := util.SanitizeServiceName(openNotifyServiceID)
	sanitizedToolName, _ := util.SanitizeToolName("getAstronauts")
	toolName := serviceID + "." + sanitizedToolName

	const maxRetries = 3
	var res *mcp.CallToolResult

	for i := 0; i < maxRetries; i++ {
		res, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(`{}`)})
		if err == nil {
			// Check if response is valid JSON before declaring success
			if len(res.Content) > 0 {
				if textContent, ok := res.Content[0].(*mcp.TextContent); ok {
					var js map[string]interface{}
					if jsonErr := json.Unmarshal([]byte(textContent.Text), &js); jsonErr == nil {
						break // Success and valid JSON
					} else {
						// Log invalid JSON error and continue retry
						t.Logf("Attempt %d/%d: Received non-JSON response from api.open-notify.org (likely rate limit message): %q. Retrying...", i+1, maxRetries, textContent.Text)
						err = jsonErr // Set err to retry logic works if this was the last attempt
						time.Sleep(2 * time.Second)
						continue
					}
				}
			}
			break // Success (or empty content, handled later)
		}

		errMsg := err.Error()
		if strings.Contains(errMsg, "503 Service Temporarily Unavailable") ||
			strings.Contains(errMsg, "context deadline exceeded") ||
			strings.Contains(errMsg, "connection reset by peer") ||
			strings.Contains(errMsg, "i/o timeout") ||
			strings.Contains(errMsg, "connection timed out") ||
			strings.Contains(errMsg, "Client.Timeout exceeded") ||
			strings.Contains(errMsg, "Too Many Requests") {
			t.Logf("Attempt %d/%d: Call to api.open-notify.org failed with a transient error: %v. Retrying...", i+1, maxRetries, err)
			time.Sleep(2 * time.Second) // Wait before retrying
			continue
		}

		require.NoError(t, err, "unrecoverable error calling getAstronauts tool")
	}

	if err != nil {
		t.Skipf("Skipping test: all %d retries to api.open-notify.org failed with transient errors. Last error: %v", maxRetries, err)
		return
	}

	require.NotNil(t, res, "Nil response from getAstronauts tool")

	// --- 4. Assert Response ---
	require.Len(t, res.Content, 1, "Expected exactly one content item")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content")

	var openNotifyResponse map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &openNotifyResponse)
	if err != nil {
		// If we still have invalid JSON after retries, skip instead of failing, as this is an external dependency issue
		t.Skipf("Skipping test: Failed to unmarshal JSON response from api.open-notify.org: %v. Body: %s", err, textContent.Text)
		return
	}

	require.Equal(t, "success", openNotifyResponse["message"], "The message should be success")
	require.NotEmpty(t, openNotifyResponse["number"], "The number should not be empty")
	require.NotEmpty(t, openNotifyResponse["people"], "The people should not be empty")
	t.Logf("SUCCESS: Received correct data for astronauts: %s", textContent.Text)

	t.Log("INFO: E2E Test Scenario for Open Notify Server Completed Successfully!")
}
