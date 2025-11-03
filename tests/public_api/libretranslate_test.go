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

func TestUpstreamService_LibreTranslate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for LibreTranslate Server...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpxTestServerInfo := integration.StartMCPANYServer(t, "E2ELibreTranslateServerTest")
	defer mcpxTestServerInfo.CleanupFunc()

	// --- 2. Register LibreTranslate Server with MCPANY ---
	const libreTranslateServiceID = "e2e_libretranslate"
	libreTranslateServiceEndpoint := "https://translate.terraprint.co"
	t.Logf("INFO: Registering '%s' with MCPANY at endpoint %s...", libreTranslateServiceID, libreTranslateServiceEndpoint)
	registrationGRPCClient := mcpxTestServerInfo.RegistrationClient

	httpCall := configv1.HttpCallDefinition_builder{
		EndpointPath: proto.String("/translate"),
		Schema: configv1.ToolSchema_builder{
			Name: proto.String("translate"),
		}.Build(),
		Method: configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HttpMethod_value["HTTP_METHOD_POST"]).Enum(),
		Parameters: []*configv1.HttpParameterMapping{
			configv1.HttpParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("q"),
				}.Build(),
			}.Build(),
			configv1.HttpParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("source"),
				}.Build(),
			}.Build(),
			configv1.HttpParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("target"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	httpService := configv1.HttpUpstreamService_builder{
		Address: proto.String(libreTranslateServiceEndpoint),
		Calls:   []*configv1.HttpCallDefinition{httpCall},
	}.Build()

	config := configv1.UpstreamServiceConfig_builder{
		Name:        proto.String(libreTranslateServiceID),
		HttpService: httpService,
	}.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	integration.RegisterServiceViaAPI(t, registrationGRPCClient, req)
	t.Logf("INFO: '%s' registered.", libreTranslateServiceID)

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

	serviceID, _ := util.SanitizeServiceName(libreTranslateServiceID)
	sanitizedToolName, _ := util.SanitizeToolName("translate")
	toolName := serviceID + "." + sanitizedToolName
	args := `{"q": "Hello, how are you?", "source": "en", "target": "es"}`

	const maxRetries = 3
	var res *mcp.CallToolResult

	for i := 0; i < maxRetries; i++ {
		res, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(args)})
		if err == nil {
			break // Success
		}

		if strings.Contains(err.Error(), "503 Service Temporarily Unavailable") || strings.Contains(err.Error(), "502 Bad Gateway") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "connection reset by peer") || strings.Contains(err.Error(), "429") {
			t.Logf("Attempt %d/%d: Call to libretranslate.com failed with a transient error: %v. Retrying...", i+1, maxRetries, err)
			time.Sleep(2 * time.Second) // Wait before retrying
			continue
		}

		require.NoError(t, err, "unrecoverable error calling translate tool")
	}

	if err != nil {
		t.Skipf("Skipping test: all %d retries to libretranslate.com failed with transient errors. Last error: %v", maxRetries, err)
	}

	require.NoError(t, err, "Error calling translate tool")
	require.NotNil(t, res, "Nil response from translate tool")

	// --- 4. Assert Response ---
	require.Len(t, res.Content, 1, "Expected exactly one content item")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content")

	var libreTranslateResponse map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &libreTranslateResponse)
	require.NoError(t, err, "Failed to unmarshal JSON response")

	require.NotEmpty(t, libreTranslateResponse["translatedText"], "The translated text should not be empty")

	t.Logf("SUCCESS: Received correct translation: %s", textContent.Text)

	t.Log("INFO: E2E Test Scenario for LibreTranslate Server Completed Successfully!")
}
