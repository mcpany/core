// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package github_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

import "os"

func TestUpstreamService_GitHub(t *testing.T) {
	if os.Getenv("GITHUB_TOKEN") == "" {
		// t.Skip("GITHUB_TOKEN is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for GitHub Server...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EGitHubServerTest", "--config-path", "../../../../examples/popular_services/github")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 2. Call Tool via MCPANY ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpAnyTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)
	require.Len(t, listToolsResult.Tools, 2, "Expected exactly two tools to be registered")

	// --- 3. Test Cases ---
	t.Run("get_user", func(t *testing.T) {
		args := json.RawMessage(`{"username": "octocat"}`)
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "github-get_user", Arguments: args})
		require.NoError(t, err)
		require.NotNil(t, res)

		require.Len(t, res.Content, 1, "Expected exactly one content item")
		textContent, ok := res.Content[0].(*mcp.TextContent)
		require.True(t, ok, "Expected text content")

		var response map[string]interface{}
		err = json.Unmarshal([]byte(textContent.Text), &response)
		require.NoError(t, err, "Failed to unmarshal JSON response")

		require.Equal(t, "octocat", response["login"], "The login should match the input")
	})

	t.Run("list_repos", func(t *testing.T) {
		args := json.RawMessage(`{"username": "octocat"}`)
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "github-list_repos", Arguments: args})
		require.NoError(t, err)
		require.NotNil(t, res)

		require.Len(t, res.Content, 1, "Expected exactly one content item")
		textContent, ok := res.Content[0].(*mcp.TextContent)
		require.True(t, ok, "Expected text content")

		var response []interface{}
		err = json.Unmarshal([]byte(textContent.Text), &response)
		require.NoError(t, err, "Failed to unmarshal JSON response")

		require.NotEmpty(t, response, "The response should not be empty")
	})

	t.Log("INFO: E2E Test Scenario for GitHub Server Completed Successfully!")
}
