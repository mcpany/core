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

func TestUpstreamService_FunTranslations(t *testing.T) {
	// t.SkipNow()
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Fun Translations Server...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EFunTranslationsServerTest")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 2. Register Fun Translations Server with MCPANY ---
	const funTranslationsServiceID = "e2e_funtranslations"
	funTranslationsServiceEndpoint := "https://api.funtranslations.com"
	t.Logf("INFO: Registering '%s' with MCPANY at endpoint %s...", funTranslationsServiceID, funTranslationsServiceEndpoint)
	registrationGRPCClient := mcpAnyTestServerInfo.RegistrationClient

	callID := "translateToYoda"

	httpCall := configv1.HttpCallDefinition_builder{
		Id:           proto.String(callID),
		EndpointPath: proto.String("/translate/yoda.json"),
		Method:       configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HttpMethod_value["HTTP_METHOD_POST"]).Enum(),
		Parameters: []*configv1.HttpParameterMapping{
			configv1.HttpParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("text"),
					Type: configv1.ParameterType_STRING.Enum(),
				}.Build(),
			}.Build(),
		},
	}.Build()

	inputSchema, err := structpb.NewStruct(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"text": map[string]interface{}{
				"type": "string",
			},
		},
		"required": []interface{}{"text"},
	})
	require.NoError(t, err)

	toolDef := configv1.ToolDefinition_builder{
		Name:        proto.String("translateToYoda"),
		Description: proto.String("Translate text to Yoda speak"),
		CallId:      proto.String(callID),
		InputSchema: inputSchema,
	}.Build()

	httpService := configv1.HttpUpstreamService_builder{
		Address: proto.String(funTranslationsServiceEndpoint),
		Tools:   []*configv1.ToolDefinition{toolDef},
		Calls:   map[string]*configv1.HttpCallDefinition{callID: httpCall},
	}.Build()

	config := configv1.UpstreamServiceConfig_builder{
		Name:        proto.String(funTranslationsServiceID),
		HttpService: httpService,
	}.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	_, err = registrationGRPCClient.RegisterService(ctx, req)
	require.NoError(t, err)
	t.Logf("INFO: '%s' registered.", funTranslationsServiceID)

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

	serviceID, _ := util.SanitizeServiceName(funTranslationsServiceID)
	sanitizedToolName, _ := util.SanitizeToolName("translateToYoda")
	toolName := serviceID + "." + sanitizedToolName
	text := `{"text": "Hello, how are you?"}`

	const maxRetries = 3
	var res *mcp.CallToolResult

	for i := 0; i < maxRetries; i++ {
		t.Logf("Attempt %d/%d calling tool %s...", i+1, maxRetries, toolName)
		res, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(text)})
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
		if strings.Contains(errMsg, "429") || strings.Contains(errMsg, "Too Many Requests") || strings.Contains(errMsg, "403") {
			t.Skipf("Skipping test due to rate limiting or access denied (403/429): %v", err)
			return
		}

		// Fail fast on other errors
		require.NoError(t, err, "unrecoverable error calling tool")
	}

	if err != nil {
		t.Skipf("Skipping test: all %d retries failed. Last error: %v", maxRetries, err)
		return
	}

	require.NotNil(t, res, "Nil response from tool")

	// --- 4. Assert Response ---
	require.Len(t, res.Content, 1, "Expected exactly one content item")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content")

	var funTranslationsResponse map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &funTranslationsResponse)
	if err != nil {
		t.Logf("Failed to unmarshal JSON. Raw response body: %s", textContent.Text)

		// Check for tool execution failure message in content
		if strings.Contains(textContent.Text, "Tool execution failed") {
			if strings.Contains(textContent.Text, "403") || strings.Contains(textContent.Text, "429") {
				t.Skipf("Skipping test due to upstream error in response: %s", textContent.Text)
				return
			}
			t.Fatalf("Tool execution failed: %s", textContent.Text)
		}

		// Check if raw response indicates an error page (HTML)
		if strings.Contains(textContent.Text, "<html") || strings.Contains(textContent.Text, "<!DOCTYPE html>") {
			t.Log("Response seems to be HTML (likely error page). Skipping test.")
			t.Skip("Skipping test due to upstream returning HTML instead of JSON (likely rate limit or maintenance).")
			return
		}
		require.NoError(t, err, "Failed to unmarshal JSON response")
	}

	if contents, ok := funTranslationsResponse["contents"].(map[string]interface{}); ok {
		require.NotEmpty(t, contents["translated"], "The translated text should not be empty")
		t.Logf("SUCCESS: Received translation: %s", contents["translated"])
	} else {
		// Sometimes API returns error JSON
		if errorObj, ok := funTranslationsResponse["error"].(map[string]interface{}); ok {
			t.Logf("API returned error: %v", errorObj)
			// Decide whether to fail or skip
			if msg, ok := errorObj["message"].(string); ok && (strings.Contains(msg, "Limit") || strings.Contains(msg, "Too Many Requests")) {
				t.Skip("Skipping due to API limit error in JSON response")
				return
			}
			t.Fail()
		} else {
			t.Logf("Unexpected JSON structure: %v", funTranslationsResponse)
			t.Fail()
		}
	}

	t.Log("INFO: E2E Test Scenario for Fun Translations Server Completed Successfully!")
}
