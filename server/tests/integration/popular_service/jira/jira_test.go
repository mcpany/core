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
	"path/filepath"
	"strings"
	"testing"

	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_Jira(t *testing.T) {
	// Setup Mock Jira Server
	mockJira := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock Swagger/OpenAPI spec
		if strings.Contains(r.URL.Path, "swagger.json") {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
  "openapi": "3.0.0",
  "info": { "title": "Jira Mock", "version": "1.0.0" },
  "paths": {
    "/rest/api/3/issue/{issueIdOrKey}": {
      "get": {
        "operationId": "getIssue",
        "parameters": [
          { "name": "issueIdOrKey", "in": "path", "required": true, "schema": { "type": "string" } }
        ],
        "responses": {
          "200": { "description": "OK", "content": { "application/json": { "schema": { "type": "object" } } } }
        }
      }
    }
  }
}`))
			return
		}
		// Mock Issue API
		// The path in spec is /rest/api/3/issue/{issueIdOrKey}
		if strings.Contains(r.URL.Path, "/rest/api/3/issue/") {
			w.Header().Set("Content-Type", "application/json")
			// Return a mock issue response
			w.Write([]byte(`{
				"key": "TEST-1",
				"fields": {
					"summary": "Test Summary"
				}
			}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockJira.Close()

	// Set Env Vars for Config to satisfy the file interpolation
	// The config uses:
	// address: "https://${JIRA_DOMAIN:jira.atlassian.com}"
	// upstream_auth uses JIRA_USERNAME and JIRA_PAT

	// We will override address and spec_url via --set, but we still need to provide dummy env vars
	// because the config file references them.
	t.Setenv("JIRA_DOMAIN", "localhost") // Not used due to override, but needed to avoid interpolation error if strict?
	t.Setenv("JIRA_USERNAME", "dummy-user")
	t.Setenv("JIRA_PAT", "dummy-token")
	t.Setenv("JIRA_TEST_ISSUE_KEY", "TEST-1")
	t.Setenv("JIRA_TEST_ISSUE_SUMMARY", "Test Summary")

	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Jira Server...")
	// t.Parallel() // Removed to allow safe os.Setenv usage if needed, although t.Setenv handles it locally.
    // StartMCPANYServer inherits os.Environ(), so t.Setenv is sufficient.

	// Override config values to point to mock server
	// The mock server URL is http://127.0.0.1:xxxxx
	// We need to override:
	// 1. openapi_service.address -> mockJira.URL
	// 2. openapi_service.spec_url -> mockJira.URL + "/swagger.json"

	rootDir := integration.ProjectRoot(t)
	configPath := filepath.Join(rootDir, "examples", "popular_services", "jira")

	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EJiraServerTest",
		"--config-path", configPath,
		"--set", fmt.Sprintf("upstream_services[0].openapi_service.address=%s", mockJira.URL),
		"--set", fmt.Sprintf("upstream_services[0].openapi_service.spec_url=%s/swagger.json", mockJira.URL),
        // Force auto_discover_tool=true just in case the example relied on default behavior changing
        "--set", "upstream_services[0].auto_discover_tool=true",
	)
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 2. Call Tool via MCPANY ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpAnyTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)

	var toolNames []string
	var registeredToolName string
	for _, tool := range listToolsResult.Tools {
		toolNames = append(toolNames, tool.Name)
		if strings.Contains(tool.Name, "getIssue") {
			registeredToolName = tool.Name
		}
	}
	t.Logf("Discovered tools: %v", toolNames)
	require.NotEmpty(t, registeredToolName, "Expected tool 'getIssue' to be registered")

	// --- 3. Test Cases ---
	testCases := []struct {
		name            string
		issueIdOrKey    string
		expectedSummary string
	}{
		{
			name:            "Get issue by key",
			issueIdOrKey:    "TEST-1",
			expectedSummary: "Test Summary",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// --- 4. Call Tool ---
			args := json.RawMessage(`{"issueIdOrKey": "` + tc.issueIdOrKey + `"}`)
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

			require.Contains(t, jiraIssueResponse, "key", "The response should contain an issue key")
			require.Equal(t, tc.issueIdOrKey, jiraIssueResponse["key"], "The issue key should match the input")

			fields, ok := jiraIssueResponse["fields"].(map[string]interface{})
			require.True(t, ok, "The response should contain a fields object")
			require.Contains(t, fields, "summary", "The fields object should contain a summary")
			require.Equal(t, tc.expectedSummary, fields["summary"], "The summary should match the expected value")
		})
	}

	t.Log("INFO: E2E Test Scenario for Jira Server Completed Successfully!")
}
