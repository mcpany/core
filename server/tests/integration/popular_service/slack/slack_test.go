// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package slack_test

import (
	"context"
	"encoding/json"
	"testing"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestUpstreamService_Slack(t *testing.T) {
	// if os.Getenv("SLACK_API_TOKEN") == "" {
	// 	t.Skip("SLACK_API_TOKEN is not set")
	// }
	// if os.Getenv("SLACK_TEST_CHANNEL") == "" {
	// 	t.Skip("SLACK_TEST_CHANNEL is not set")
	// }

	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Slack Server...")
	t.Parallel()

	// --- 1. Start Mock Slack API Server ---
	mockResponse := `{
		"ok": true,
		"channel": "C12345678",
		"ts": "1503435956.000247",
		"message": {
			"text": "Hello, World!",
			"username": "ecto",
			"bot_id": "B12345678",
			"attachments": [
				{
					"text": "This is an attachment",
					"id": 1,
					"fallback": "This is an attachment's fallback"
				}
			],
			"type": "message",
			"subtype": "bot_message",
			"ts": "1503435956.000247"
		}
	}`
	mockHandler := integration.DefaultMockHandler(t, map[string]string{
		"/chat.postMessage": mockResponse,
	})
	mockServer := integration.StartMockServer(t, mockHandler)
	defer mockServer.Close()

	// --- 2. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2ESlackServerTest")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 3. Register Service Dynamically ---
	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("slack"),
		HttpService: configv1.HttpUpstreamService_builder{
			Address: proto.String(mockServer.URL),
			Tools: []*configv1.ToolDefinition{
				configv1.ToolDefinition_builder{
					Name: proto.String("send_message"),
					CallId: proto.String("send_message"),
				}.Build(),
			},
			Calls: map[string]*configv1.HttpCallDefinition{
				"send_message": configv1.HttpCallDefinition_builder{
					Method:       configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HttpMethod_value["HTTP_METHOD_POST"]).Enum(),
					EndpointPath: proto.String("/chat.postMessage"),
					Parameters: []*configv1.HttpParameterMapping{
						configv1.HttpParameterMapping_builder{
							Schema: configv1.ParameterSchema_builder{
								Name: proto.String("channel"),
								Type: configv1.ParameterType(configv1.ParameterType_value["STRING"]).Enum(),
							}.Build(),
						}.Build(),
						configv1.HttpParameterMapping_builder{
							Schema: configv1.ParameterSchema_builder{
								Name: proto.String("text"),
								Type: configv1.ParameterType(configv1.ParameterType_value["STRING"]).Enum(),
							}.Build(),
						}.Build(),
					},
				}.Build(),
			},
		}.Build(),
	}.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	integration.RegisterServiceViaAPI(t, mcpAnyTestServerInfo.RegistrationClient, req)

	// --- 4. Call Tool via MCPANY ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpAnyTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)

	found := false
	var toolNames []string
	for _, tool := range listToolsResult.Tools {
		toolNames = append(toolNames, tool.Name)
		if tool.Name == "slack.send_message" {
			found = true
			break
		}
	}
	require.Truef(t, found, "Expected slack.send_message tool to be registered. Found: %v", toolNames)

	// --- 5. Test Cases ---
	t.Run("send_message", func(t *testing.T) {
		args := json.RawMessage(`{"channel": "C12345678", "text": "Hello, World!"}`)
		// Tool name is usually ServiceName.ToolName for http_service
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "slack.send_message", Arguments: args})
		require.NoError(t, err)
		require.NotNil(t, res)

		require.Len(t, res.Content, 1, "Expected exactly one content item")
		textContent, ok := res.Content[0].(*mcp.TextContent)
		require.True(t, ok, "Expected text content")

		var response map[string]interface{}
		err = json.Unmarshal([]byte(textContent.Text), &response)
		require.NoError(t, err, "Failed to unmarshal JSON response")

		require.True(t, response["ok"].(bool), "The response should be ok")
	})

	t.Log("INFO: E2E Test Scenario for Slack Server Completed Successfully!")
}
