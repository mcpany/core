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

func TestUpstreamService_FunTranslations(t *testing.T) {
	t.SkipNow()
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
		InputTransformer: configv1.InputTransformer_builder{
			Template: proto.String("{\"text\": \"{{.input.text}}\"}"),
		}.Build(),
	}.Build()

	toolDef := configv1.ToolDefinition_builder{
		Name:   proto.String("translateToYoda"),
		CallId: proto.String(callID),
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

	integration.RegisterServiceViaAPI(t, registrationGRPCClient, req)
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
		res, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(text)})
		if err == nil {
			break // Success
		}

		if strings.Contains(err.Error(), "503 Service Temporarily Unavailable") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "connection reset by peer") {
			t.Logf("Attempt %d/%d: Call to api.funtranslations.com failed with a transient error: %v. Retrying...", i+1, maxRetries, err)
			time.Sleep(2 * time.Second) // Wait before retrying
			continue
		}
		if strings.Contains(err.Error(), "upstream HTTP request failed with status 429") {
			t.Skipf("Skipping test due to rate limiting from api.funtranslations.com: %v", err)
		}

		require.NoError(t, err, "unrecoverable error calling translateToYoda tool")
	}

	if err != nil {
		t.Skipf("Skipping test: all %d retries to api.funtranslations.com failed with transient errors. Last error: %v", maxRetries, err)
	}

	require.NoError(t, err, "Error calling translateToYoda tool")
	require.NotNil(t, res, "Nil response from translateToYoda tool")

	// --- 4. Assert Response ---
	require.Len(t, res.Content, 1, "Expected exactly one content item")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content")

	var funTranslationsResponse map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &funTranslationsResponse)
	require.NoError(t, err, "Failed to unmarshal JSON response")

	contents := funTranslationsResponse["contents"].(map[string]interface{})
	require.NotEmpty(t, contents["translated"], "The translated text should not be empty")

	t.Logf("SUCCESS: Received correct translation: %s", textContent.Text)

	t.Log("INFO: E2E Test Scenario for Fun Translations Server Completed Successfully!")
}
