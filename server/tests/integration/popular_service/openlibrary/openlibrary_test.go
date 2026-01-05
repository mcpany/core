// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package openlibrary_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_OpenLibrary(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Open Library Server...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EOpenLibraryServerTest", "--config-path", "../../../../examples/popular_services/openlibrary")
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
	t.Run("getBookByISBN", func(t *testing.T) {
		args := json.RawMessage(`{"isbn": "9780140328721"}`)
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "getBookByISBN", Arguments: args})
		require.NoError(t, err)
		require.NotNil(t, res)

		require.Len(t, res.Content, 1, "Expected exactly one content item")
		textContent, ok := res.Content[0].(*mcp.TextContent)
		require.True(t, ok, "Expected text content")

		var response map[string]interface{}
		err = json.Unmarshal([]byte(textContent.Text), &response)
		require.NoError(t, err, "Failed to unmarshal JSON response")

		require.Contains(t, response, "ISBN:9780140328721")
		bookInfo := response["ISBN:9780140328721"].(map[string]interface{})
		require.Equal(t, "The Outsiders", bookInfo["title"])
	})

	t.Run("searchAuthors", func(t *testing.T) {
		args := json.RawMessage(`{"author_name": "S.E. Hinton"}`)
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "searchAuthors", Arguments: args})
		require.NoError(t, err)
		require.NotNil(t, res)

		require.Len(t, res.Content, 1, "Expected exactly one content item")
		textContent, ok := res.Content[0].(*mcp.TextContent)
		require.True(t, ok, "Expected text content")

		var response map[string]interface{}
		err = json.Unmarshal([]byte(textContent.Text), &response)
		require.NoError(t, err, "Failed to unmarshal JSON response")

		require.Contains(t, response, "docs")
		docs := response["docs"].([]interface{})
		require.True(t, len(docs) > 0)
		authorInfo := docs[0].(map[string]interface{})
		require.Equal(t, "S. E. Hinton", authorInfo["name"])
	})

	t.Log("INFO: E2E Test Scenario for Open Library Server Completed Successfully!")
}
