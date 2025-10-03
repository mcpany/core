/*
 * Copyright 2025 Author(s) of MCPX
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

package public_api

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/mcpxy/mcpx/pkg/consts"
	"github.com/mcpxy/mcpx/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestPublicTimeMCPAPI(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for 'time' server (Stdio via Docker)...")

	// --- 1. Start MCPX Server ---
	mcpxTestServerInfo := integration.StartMCPXServer(t, "E2ETimeServerTest")
	defer mcpxTestServerInfo.CleanupFunc()

	// --- 2. Register 'time' server with MCPX ---
	const timeServiceID = "e2e-time-server"
	command := "docker"
	args := []string{"run", "--rm", "-i", "mcpx-e2e-time-server"}

	t.Logf("INFO: Registering '%s' with MCPX...", timeServiceID)
	registrationGRPCClient := mcpxTestServerInfo.RegistrationClient
	integration.RegisterStdioService(t, registrationGRPCClient, timeServiceID, command, true, args...)
	t.Logf("INFO: '%s' registered with command: %s %v", timeServiceID, command, args)

	// --- 3. Use MCP SDK to connect and call the tool ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpxTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err, "Failed to connect to MCPX server")
	defer cs.Close()

	toolName := fmt.Sprintf("%s%sget_local_time", timeServiceID, consts.ToolNameServiceSeparator)

	// Wait for the tool to be available
	require.Eventually(t, func() bool {
		result, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
		if err != nil {
			t.Logf("Failed to list tools: %v", err)
			return false
		}
		for _, tool := range result.Tools {
			if tool.Name == toolName {
				return true
			}
		}
		t.Logf("Tool %s not yet available", toolName)
		return false
	}, integration.TestWaitTimeLong, 1*time.Second, "Tool %s did not become available in time", toolName)

	params := json.RawMessage(`{}`)

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: params})
	require.NoError(t, err, "Error calling tool '%s'", toolName)
	require.NotNil(t, res, "Nil response from tool '%s'", toolName)
	require.Len(t, res.Content, 1, "Expected exactly one content item from tool '%s'", toolName)

	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected content to be of type TextContent")

	// Log the raw text content for debugging
	t.Logf("Raw tool output: %s", textContent.Text)

	var jsonResponse map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &jsonResponse)
	require.NoError(t, err, "Failed to unmarshal JSON response from time server")

	timeStr, ok := jsonResponse["current_time"].(string)
	require.True(t, ok, "Response should contain a 'current_time' string")
	_, err = time.Parse("2006-01-02 15:04:05", timeStr)
	require.NoError(t, err, "Failed to parse time string: %s", timeStr)

	tzStr, ok := jsonResponse["timezone"].(string)
	require.True(t, ok, "Response should contain a 'timezone' string")
	require.NotEmpty(t, tzStr, "Timezone should not be empty")

	t.Logf("SUCCESS: Call to tool '%s' via MCPX completed. Received time: %s, timezone: %s", toolName, timeStr, tzStr)
	t.Log("INFO: E2E Test Scenario for 'time' server (Stdio via Docker) Completed Successfully!")
}
