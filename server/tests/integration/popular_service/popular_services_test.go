// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package popular_service_test

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

func TestUpstreamService_Trello(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Trello Server...")
	t.Parallel()

	// --- 1. Start Mock Server ---
	mockResponse := `{"id": "123", "name": "Test Board"}`
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockResponse))
	}))
	defer mockServer.Close()

	// --- 2. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2ETrelloServerTest")
	defer mcpAnyTestServerInfo.CleanupFunc()

	serviceID := "e2e_trello"
	config := &configv1.UpstreamServiceConfig{
		Id:   proto.String(serviceID),
		Name: proto.String(serviceID),
		HttpService: &configv1.HttpUpstreamService{
			Address: proto.String(mockServer.URL),
			Tools: []*configv1.ToolDefinition{
				{
					Name:   proto.String("getBoard"),
					CallId: proto.String("getBoardCall"),
				},
			},
			Calls: map[string]*configv1.HttpCallDefinition{
				"getBoardCall": {
					Id:           proto.String("getBoardCall"),
					EndpointPath: proto.String("/1/boards/123"),
					Method:       configv1.HttpCallDefinition_HTTP_METHOD_GET.Enum(),
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
			if tool.Name == serviceID+".getBoard" {
				found = true
				break
			}
		}
		require.True(t, found, "Tool getBoard should be present")
	})

	t.Run("CallTool getBoard", func(t *testing.T) {
		res, err := client.CallTool(ctx, mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
			}{
				Name: serviceID + ".getBoard",
				Arguments: map[string]any{},
			},
		})
		require.NoError(t, err, "Error calling getBoard tool")
		require.NotNil(t, res, "Nil response from getBoard tool")
		require.False(t, res.IsError, "Tool execution returned an error")

		require.Len(t, res.Content, 1, "Expected exactly one content item")
		textContent, ok := res.Content[0].(mcp.TextContent)
		require.True(t, ok, "Expected text content")

		var trelloResponse map[string]interface{}
		err = json.Unmarshal([]byte(textContent.Text), &trelloResponse)
		require.NoError(t, err)

		require.Equal(t, "123", trelloResponse["id"])
		require.Equal(t, "Test Board", trelloResponse["name"])
	})
}

func TestUpstreamService_Miro(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Miro Server...")
	t.Parallel()

	// --- 1. Start Mock Server ---
	mockResponse := `{"type": "board", "name": "Test Miro Board"}`
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockResponse))
	}))
	defer mockServer.Close()

	// --- 2. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EMiroServerTest")
	defer mcpAnyTestServerInfo.CleanupFunc()

	serviceID := "e2e_miro"
	config := &configv1.UpstreamServiceConfig{
		Id:   proto.String(serviceID),
		Name: proto.String(serviceID),
		HttpService: &configv1.HttpUpstreamService{
			Address: proto.String(mockServer.URL),
			Tools: []*configv1.ToolDefinition{
				{
					Name:   proto.String("getBoard"),
					CallId: proto.String("getBoardCall"),
				},
			},
			Calls: map[string]*configv1.HttpCallDefinition{
				"getBoardCall": {
					Id:           proto.String("getBoardCall"),
					EndpointPath: proto.String("/v2/boards/123"),
					Method:       configv1.HttpCallDefinition_HTTP_METHOD_GET.Enum(),
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

	t.Run("CallTool getBoard", func(t *testing.T) {
		res, err := client.CallTool(ctx, mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
			}{
				Name: serviceID + ".getBoard",
				Arguments: map[string]any{},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.False(t, res.IsError)

		textContent, ok := res.Content[0].(mcp.TextContent)
		require.True(t, ok)

		var miroResponse map[string]interface{}
		json.Unmarshal([]byte(textContent.Text), &miroResponse)

		require.Equal(t, "board", miroResponse["type"])
		require.Equal(t, "Test Miro Board", miroResponse["name"])
	})
}

func TestUpstreamService_Figma(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Figma Server...")
	t.Parallel()

	// --- 1. Start Mock Server ---
	mockResponse := `{"name": "Test Figma File", "version": "1.0"}`
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockResponse))
	}))
	defer mockServer.Close()

	// --- 2. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EFigmaServerTest")
	defer mcpAnyTestServerInfo.CleanupFunc()

	serviceID := "e2e_figma"
	config := &configv1.UpstreamServiceConfig{
		Id:   proto.String(serviceID),
		Name: proto.String(serviceID),
		HttpService: &configv1.HttpUpstreamService{
			Address: proto.String(mockServer.URL),
			Tools: []*configv1.ToolDefinition{
				{
					Name:   proto.String("getFile"),
					CallId: proto.String("getFileCall"),
				},
			},
			Calls: map[string]*configv1.HttpCallDefinition{
				"getFileCall": {
					Id:           proto.String("getFileCall"),
					EndpointPath: proto.String("/v1/files/123"),
					Method:       configv1.HttpCallDefinition_HTTP_METHOD_GET.Enum(),
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

	t.Run("CallTool getFile", func(t *testing.T) {
		res, err := client.CallTool(ctx, mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
			}{
				Name: serviceID + ".getFile",
				Arguments: map[string]any{},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.False(t, res.IsError)

		textContent, ok := res.Content[0].(mcp.TextContent)
		require.True(t, ok)

		var figmaResponse map[string]interface{}
		json.Unmarshal([]byte(textContent.Text), &figmaResponse)

		require.Equal(t, "Test Figma File", figmaResponse["name"])
	})
}
