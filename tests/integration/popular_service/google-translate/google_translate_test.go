/*
 * Copyright 2024 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

//go:build e2e

package googletranslate_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/mcpany/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_GoogleTranslate(t *testing.T) {
	if os.Getenv("GOOGLE_API_KEY") == "" {
		t.Skip("Skipping Google Translate test because GOOGLE_API_KEY is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Google Translate Server...")
	t.Parallel()

	// --- 1. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EGoogleTranslateServerTest", "--config-path", "../../../../examples/popular_services/google-translate")
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

	// --- 3. Test Case ---
	args := json.RawMessage(`{"q": "Hello, world!", "source": "en", "target": "es"}`)
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: registeredToolName, Arguments: args})
	require.NoError(t, err)
	require.NotNil(t, res)

	// --- 4. Assert Response ---
	require.Len(t, res.Content, 1, "Expected exactly one content item")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content")

	var translateResponse map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &translateResponse)
	require.NoError(t, err, "Failed to unmarshal JSON response")

	data, ok := translateResponse["data"].(map[string]interface{})
	require.True(t, ok, "The response should contain a data object")

	translations, ok := data["translations"].([]interface{})
	require.True(t, ok, "The data object should contain a translations array")
	require.Len(t, translations, 1, "The translations array should contain one item")

	translation, ok := translations[0].(map[string]interface{})
	require.True(t, ok, "The translation item should be a map")
	require.Contains(t, translation, "translatedText", "The translation should contain translatedText")
	require.Equal(t, "Hola Mundo", translation["translatedText"], "The translated text should be 'Hola Mundo'")

	t.Log("INFO: E2E Test Scenario for Google Translate Server Completed Successfully!")
}
