// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package gemini_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestUpstreamService_Gemini(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Gemini Server...")
	t.Parallel()

	// --- 1. Start Mock Server ---
	mockResponse := `{"candidates": [{"content": {"parts": [{"text": "Hello, how can I help you?"}]}}]}`
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockResponse))
	}))
	defer mockServer.Close()

	// --- 2. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EGeminiServerTest")
	defer mcpAnyTestServerInfo.CleanupFunc()

	serviceID := "e2e_gemini"
	config := &configv1.UpstreamServiceConfig{
		Id:   proto.String(serviceID),
		Name: proto.String(serviceID),
		HttpService: &configv1.HttpUpstreamService{
			Address: proto.String(mockServer.URL),
			Tools: []*configv1.ToolDefinition{
				{
					Name:   proto.String("generateContent"),
					CallId: proto.String("generateContentCall"),
				},
			},
			Calls: map[string]*configv1.HttpCallDefinition{
				"generateContentCall": {
					Id:           proto.String("generateContentCall"),
					EndpointPath: proto.String("/v1/models/gemini-pro:generateContent"),
					Method:       configv1.HttpCallDefinition_HTTP_METHOD_POST.Enum(),
					Parameters: []*configv1.HttpParameterMapping{
						{
							Schema: &configv1.ParameterSchema{
								Name: proto.String("contents"),
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

	t.Run("CallTool generateContent", func(t *testing.T) {
		res, err := client.CallTool(ctx, mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
			}{
				Name: serviceID + ".generateContent",
				Arguments: map[string]any{
					"contents": []interface{}{
						map[string]interface{}{
							"parts": []interface{}{
								map[string]interface{}{
									"text": "Hello!",
								},
							},
						},
					},
				},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.False(t, res.IsError)

		textContent, ok := res.Content[0].(mcp.TextContent)
		require.True(t, ok)

		var geminiResponse map[string]interface{}
		json.Unmarshal([]byte(textContent.Text), &geminiResponse)

		candidates := geminiResponse["candidates"].([]interface{})
		require.Len(t, candidates, 1)

		candidate := candidates[0].(map[string]interface{})
		content := candidate["content"].(map[string]interface{})
		parts := content["parts"].([]interface{})
		part := parts[0].(map[string]interface{})

		require.Equal(t, "Hello, how can I help you?", part["text"])
	})
}
