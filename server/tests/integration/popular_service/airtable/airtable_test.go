// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package airtable_test

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

func TestUpstreamService_Airtable(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Airtable Server...")
	t.Parallel()

	// --- 1. Start Mock Server ---
	mockResponse := `{"records": [{"id": "rec123", "fields": {"Name": "Test Record"}}]}`
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockResponse))
	}))
	defer mockServer.Close()

	// --- 2. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EAirtableServerTest")
	defer mcpAnyTestServerInfo.CleanupFunc()

	serviceID := "e2e_airtable"
	config := &configv1.UpstreamServiceConfig{
		Id:   proto.String(serviceID),
		Name: proto.String(serviceID),
		HttpService: &configv1.HttpUpstreamService{
			Address: proto.String(mockServer.URL),
			Tools: []*configv1.ToolDefinition{
				{
					Name:   proto.String("listRecords"),
					CallId: proto.String("listRecordsCall"),
				},
			},
			Calls: map[string]*configv1.HttpCallDefinition{
				"listRecordsCall": {
					Id:           proto.String("listRecordsCall"),
					EndpointPath: proto.String("/v0/app123/tbl123"),
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

	t.Run("CallTool listRecords", func(t *testing.T) {
		res, err := client.CallTool(ctx, mcp.CallToolRequest{
			Params: struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments,omitempty"`
			}{
				Name: serviceID + ".listRecords",
				Arguments: map[string]any{},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.False(t, res.IsError)

		textContent, ok := res.Content[0].(mcp.TextContent)
		require.True(t, ok)

		var airtableResponse map[string]interface{}
		json.Unmarshal([]byte(textContent.Text), &airtableResponse)

		records := airtableResponse["records"].([]interface{})
		require.Len(t, records, 1)

		record := records[0].(map[string]interface{})
		require.Equal(t, "rec123", record["id"])
	})
}
