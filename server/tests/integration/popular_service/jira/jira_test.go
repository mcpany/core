// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package jira_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_Jira(t *testing.T) {
	jiraDomain := os.Getenv("JIRA_DOMAIN")
	jiraUsername := os.Getenv("JIRA_USERNAME")
	jiraPat := os.Getenv("JIRA_PAT")
	jiraTestIssueKey := os.Getenv("JIRA_TEST_ISSUE_KEY")
	jiraTestIssueSummary := os.Getenv("JIRA_TEST_ISSUE_SUMMARY")

	if jiraDomain == "" || jiraUsername == "" || jiraPat == "" || jiraTestIssueKey == "" || jiraTestIssueSummary == "" {
		// t.Skip("Skipping Jira integration test because one or more of the required environment variables are not set: JIRA_DOMAIN, JIRA_USERNAME, JIRA_PAT, JIRA_TEST_ISSUE_KEY, JIRA_TEST_ISSUE_SUMMARY")
	}

	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Jira Server...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EJiraServerTest", "--config-path", "../../../../examples/popular_services/jira")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 2. Call Tool via MCPANY ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpAnyTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)
	require.Len(t, listToolsResult.Tools, 1, "Expected exactly one tool to be registered")
	registeredToolName := listToolsResult.Tools[0].Name
	t.Logf("Discovered tool from MCPANY: %s", registeredToolName)

	// --- 3. Test Cases ---
	testCases := []struct {
		name            string
		issueIdOrKey    string
		expectedSummary string
	}{
		{
			name:            "Get issue by key",
			issueIdOrKey:    jiraTestIssueKey,
			expectedSummary: jiraTestIssueSummary,
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
