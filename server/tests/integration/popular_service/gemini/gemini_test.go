// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package gemini_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_Gemini(t *testing.T) {
	if os.Getenv("GEMINI_API_KEY") == "" {
		// t.Skip("Skipping test because GEMINI_API_KEY is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Gemini Server...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EGeminiServerTest", "--config-path", "../../../../examples/popular_services/gemini")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 2. Call Tool via MCPANY ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpAnyTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)
	require.Len(t, listToolsResult.Tools, 1, "Expected exactly one tool to be registered")

	// --- 3. Test Cases ---
	testCases := []struct {
		name   string
		prompt string
	}{
		{
			name:   "Generate content with a valid prompt",
			prompt: "Write a story about a magic backpack.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// --- 4. Call generateContent Tool ---
			args := json.RawMessage(`{"contents": [{"parts": [{"text": "` + tc.prompt + `"}]}]}`)
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "gemini/-/generateContent", Arguments: args})
			require.NoError(t, err)
			require.NotNil(t, res)

			// --- 5. Assert generateContent Response ---
			require.Len(t, res.Content, 1, "Expected exactly one content item")
			textContent, ok := res.Content[0].(*mcp.TextContent)
			require.True(t, ok, "Expected text content")

			var response map[string]interface{}
			err = json.Unmarshal([]byte(textContent.Text), &response)
			require.NoError(t, err, "Failed to unmarshal JSON response")

			candidates, ok := response["candidates"].([]interface{})
			require.True(t, ok)
			require.NotEmpty(t, candidates, "Expected at least one candidate")
		})
	}

	t.Log("INFO: E2E Test Scenario for Gemini Server Completed Successfully!")
}
