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
	"google.golang.org/protobuf/types/known/structpb"
)

func TestUpstreamService_Nationalize(t *testing.T) {
	// t.SkipNow()
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Nationalize Server...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2ENationalizeServerTest")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 2. Register Nationalize Server with MCPANY ---
	const nationalizeServiceID = "e2e_nationalize"
	nationalizeServiceEndpoint := "https://api.nationalize.io"
	t.Logf("INFO: Registering '%s' with MCPANY at endpoint %s...", nationalizeServiceID, nationalizeServiceEndpoint)
	registrationGRPCClient := mcpAnyTestServerInfo.RegistrationClient

	callID := "getNationality"
	httpCall := configv1.HttpCallDefinition_builder{
		Id:           proto.String(callID),
		EndpointPath: proto.String("/?name=michael"),
		Method:       configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HttpMethod_value["HTTP_METHOD_GET"]).Enum(),
	}.Build()

	inputSchema, err := structpb.NewStruct(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type": "string",
			},
		},
	})
	require.NoError(t, err)

	toolDef := configv1.ToolDefinition_builder{
		Name:        proto.String("getNationality"),
		CallId:      proto.String(callID),
		InputSchema: inputSchema,
	}.Build()

	httpService := configv1.HttpUpstreamService_builder{
		Address: proto.String(nationalizeServiceEndpoint),
		Tools:   []*configv1.ToolDefinition{toolDef},
		Calls:   map[string]*configv1.HttpCallDefinition{callID: httpCall},
	}.Build()

	config := configv1.UpstreamServiceConfig_builder{
		Name:        proto.String(nationalizeServiceID),
		HttpService: httpService,
	}.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	integration.RegisterServiceViaAPI(t, registrationGRPCClient, req)
	t.Logf("INFO: '%s' registered.", nationalizeServiceID)

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

	serviceID, _ := util.SanitizeServiceName(nationalizeServiceID)
	sanitizedToolName, _ := util.SanitizeToolName("getNationality")
	toolName := serviceID + "." + sanitizedToolName
	name := `{"name": "michael"}`

	const maxRetries = 3
	var res *mcp.CallToolResult

	for i := 0; i < maxRetries; i++ {
		t.Logf("Attempt %d/%d calling tool %s...", i+1, maxRetries, toolName)
		res, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(name)})
		if err == nil {
			break // Success
		}

		errMsg := err.Error()
		if strings.Contains(errMsg, "503 Service Temporarily Unavailable") ||
		   strings.Contains(errMsg, "context deadline exceeded") ||
		   strings.Contains(errMsg, "connection reset by peer") ||
		   strings.Contains(errMsg, "504 Gateway Timeout") {
			t.Logf("Attempt %d failed with transient error: %v. Retrying...", i+1, err)
			time.Sleep(2 * time.Second)
			continue
		}
		if strings.Contains(errMsg, "429") || strings.Contains(errMsg, "Too Many Requests") {
			t.Skipf("Skipping test due to rate limiting: %v", err)
			return
		}

		require.NoError(t, err, "unrecoverable error calling getNationality tool")
	}

	if err != nil {
		t.Skipf("Skipping test: all %d retries failed. Last error: %v", maxRetries, err)
		return
	}

	require.NotNil(t, res, "Nil response from getNationality tool")

	// --- 4. Assert Response ---
	require.Len(t, res.Content, 1, "Expected exactly one content item")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content")

	var nationalizeResponse map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &nationalizeResponse)
	if err != nil {
		t.Logf("Failed to unmarshal JSON. Raw response body: %s", textContent.Text)
		// Check if raw response indicates an error page (HTML)
		if strings.Contains(textContent.Text, "<html") || strings.Contains(textContent.Text, "<!DOCTYPE html>") {
			t.Log("Response seems to be HTML (likely error page). Skipping test.")
			t.Skip("Skipping test due to upstream returning HTML instead of JSON (likely rate limit or maintenance).")
			return
		}
		require.NoError(t, err, "Failed to unmarshal JSON response")
	}

	// Handle error response from API
	if errorVal, ok := nationalizeResponse["error"]; ok {
		t.Logf("API returned error: %v", errorVal)
		t.Skip("Skipping test due to API error (likely rate limit)")
		return
	}

	require.Equal(t, "michael", nationalizeResponse["name"], "The name does not match")
	require.NotEmpty(t, nationalizeResponse["country"], "The country should not be empty")
	t.Logf("SUCCESS: Received correct nationality for michael: %s", textContent.Text)

	t.Log("INFO: E2E Test Scenario for Nationalize Server Completed Successfully!")
}
