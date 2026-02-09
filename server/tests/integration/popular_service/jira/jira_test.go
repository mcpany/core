// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package jira_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_Jira(t *testing.T) {
	// Mock Jira Server
	mockJira := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Serve minimal Swagger Spec if requested
		if strings.Contains(r.URL.Path, "swagger.v3.json") {
			w.Write([]byte(`{
				"openapi": "3.0.0",
				"info": { "title": "Jira", "version": "3" },
				"paths": {
					"/rest/api/3/issue/{id}": {
						"get": {
							"operationId": "getIssue",
							"parameters": [
								{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }
							],
							"responses": {
								"200": {
									"description": "Success",
									"content": {
										"application/json": {
											"schema": { "type": "object", "properties": { "key": { "type": "string" }, "fields": { "type": "object", "properties": { "summary": { "type": "string" } } } } }
										}
									}
								}
							}
						}
					}
				}
			}`))
			return
		}

		// Handle Issue Request
		if strings.HasPrefix(r.URL.Path, "/rest/api/3/issue/") {
			key := strings.TrimPrefix(r.URL.Path, "/rest/api/3/issue/")
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"key": key,
				"fields": map[string]interface{}{
					"summary": "Mock Issue Summary",
				},
			})
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockJira.Close()

	// Create Config
	configContent := fmt.Sprintf(`
upstream_services:
- name: jira
  upstream_auth:
    basic_auth:
      username: "user"
      password:
        plain_text: "token"
  openapi_service:
    address: "%s"
    spec_url: "%s/swagger.v3.json"
`, mockJira.URL, mockJira.URL)

	configFile := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(configFile, []byte(configContent), 0644))

	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Jira Server (Mocked)...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EJiraServerTest", "--config-path", configFile)
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 2. Call Tool via MCPANY ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpAnyTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	// Wait for tool registration? ListTools poller?
	// integration.StartMCPANYServer waits for health check, but tool registration is async.
	// We might need to retry ListTools or wait.
	var listToolsResult *mcp.ListToolsResult
	var registeredToolName string
	require.Eventually(t, func() bool {
		listToolsResult, err = cs.ListTools(ctx, &mcp.ListToolsParams{})
		if err != nil {
			return false
		}
		for _, tool := range listToolsResult.Tools {
			// Look for our jira tool
			if strings.Contains(tool.Name, "getIssue") {
				registeredToolName = tool.Name
				return true
			}
		}
		return false
	}, integration.TestWaitTimeShort, 100*time.Millisecond, "Expected jira tools to be registered. Got: %v", func() []string {
		if listToolsResult == nil {
			return nil
		}
		var names []string
		for _, t := range listToolsResult.Tools {
			names = append(names, t.Name)
		}
		return names
	}())

	t.Logf("Discovered tool from MCPANY: %s", registeredToolName)
	for _, tool := range listToolsResult.Tools {
		if tool.Name == registeredToolName {
			schema, _ := json.Marshal(tool.InputSchema)
			t.Logf("Tool Schema: %s", string(schema))
		}
	}

	// --- 3. Test Cases ---
	testCases := []struct {
		name            string
		issueIdOrKey    string
		expectedSummary string
	}{
		{
			name:            "Get issue by key",
			issueIdOrKey:    "TEST-123",
			expectedSummary: "Mock Issue Summary",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// --- 4. Call Tool ---
			// We map "issueIdOrKey" from test case to "id" argument for tool
			args := json.RawMessage(`{"id": "` + tc.issueIdOrKey + `"}`)
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: registeredToolName, Arguments: args})
			require.NoError(t, err)
			require.NotNil(t, res)

			// --- 5. Assert Response ---
			require.Len(t, res.Content, 1, "Expected exactly one content item")
			textContent, ok := res.Content[0].(*mcp.TextContent)
			require.True(t, ok, "Expected text content")

			var jiraIssueResponse map[string]interface{}
			err = json.Unmarshal([]byte(textContent.Text), &jiraIssueResponse)
			require.NoError(t, err, "Failed to unmarshal JSON response")

			require.Equal(t, tc.issueIdOrKey, jiraIssueResponse["key"], "The issue key should match the input")

			fields, ok := jiraIssueResponse["fields"].(map[string]interface{})
			require.True(t, ok, "The response should contain a fields object")
			require.Equal(t, tc.expectedSummary, fields["summary"], "The summary should match the expected value")
		})
	}

	t.Log("INFO: E2E Test Scenario for Jira Server Completed Successfully!")
}
