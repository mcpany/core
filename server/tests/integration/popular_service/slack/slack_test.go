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
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2ESlackTest")
	defer mcpAnyTestServerInfo.CleanupFunc()

	serviceID := "e2e_slack"
	config := &configv1.UpstreamServiceConfig{
		Id:   proto.String(serviceID),
		Name: proto.String(serviceID),
		HttpService: &configv1.HttpUpstreamService{
			Address: proto.String(mockServer.URL),
			Tools: []*configv1.ToolDefinition{
				{
					Name:   proto.String("postMessage"),
					CallId: proto.String("postMessageCall"),
				},
			},
			Calls: map[string]*configv1.HttpCallDefinition{
				"postMessageCall": {
					Id:           proto.String("postMessageCall"),
					EndpointPath: proto.String("/chat.postMessage"),
					Method:       configv1.HttpCallDefinition_HTTP_METHOD_POST.Enum(),
					Parameters: []*configv1.HttpParameterMapping{
						{
							Schema: &configv1.ParameterSchema{
								Name: proto.String("channel"),
							},
							In: configv1.HttpParameterMapping_BODY.Enum(),
						},
						{
							Schema: &configv1.ParameterSchema{
								Name: proto.String("text"),
							},
							In: configv1.HttpParameterMapping_BODY.Enum(),
						},
					},
				},
			},
		},
	}

	req := &apiv1.RegisterServiceRequest{
		Config: config,
	}

	integration.RegisterServiceViaAPI(t, mcpAnyTestServerInfo.RegistrationClient, req)
	t.Logf("INFO: '%s' registered.", serviceID)

	// --- 3. Call Tool via MCPANY ---
	client, session, cleanup := integration.ConnectToMCPServer(t, ctx, mcpAnyTestServerInfo.MCPAddress, mcpAnyTestServerInfo.APIKey)
	defer cleanup()
	defer session.Close()

	t.Run("ListTools", func(t *testing.T) {
		listToolsResult, err := client.ListTools(ctx, mcp.ListToolsRequest{})
		require.NoError(t, err)

		found := false
		for _, tool := range listToolsResult.Tools {
			if tool.Name == serviceID+".postMessage" {
				found = true
				break
			}
		}
		require.True(t, found, "Tool postMessage should be present")
	})

	t.Run("CallTool postMessage", func(t *testing.T) {
		res, err := client.CallTool(ctx, mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
			}{
				Name: serviceID + ".postMessage",
				Arguments: map[string]any{
					"channel": "C12345678",
					"text":    "Hello, World!",
				},
			},
		})
		require.NoError(t, err, "Error calling postMessage tool")
		require.NotNil(t, res, "Nil response from postMessage tool")
		require.False(t, res.IsError, "Tool execution returned an error")

		require.Len(t, res.Content, 1, "Expected exactly one content item")
		textContent, ok := res.Content[0].(mcp.TextContent)
		require.True(t, ok, "Expected text content")

		t.Logf("Response body: %s", textContent.Text)

		// Parse the result as JSON
		var slackResponse map[string]interface{}
		err = json.Unmarshal([]byte(textContent.Text), &slackResponse)
		require.NoError(t, err, "Failed to unmarshal JSON response")

		// Basic assertions based on the mock data
		require.Equal(t, true, slackResponse["ok"], "Slack response should be OK")
		require.Equal(t, "C12345678", slackResponse["channel"], "Unexpected channel ID")

		message, ok := slackResponse["message"].(map[string]interface{})
		require.True(t, ok, "message should be an object")
		require.Equal(t, "Hello, World!", message["text"], "Unexpected message text")

		t.Logf("SUCCESS: Received correct Slack response for channel: %v", slackResponse["channel"])
	})

	t.Log("INFO: E2E Test Scenario for Slack Server Completed Successfully!")
}
