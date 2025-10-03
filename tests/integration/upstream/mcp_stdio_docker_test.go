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

package upstream_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/mcpxy/mcpx/tests/integration"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_MCP_Stdio_WithSetupCommandsInDocker(t *testing.T) {
	if !integration.IsDockerSocketAccessible() {
		t.Skip("Docker socket not accessible, skipping test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for 'cowsay' server (Stdio via Docker with setup)...")

	// --- 1. Start MCPX Server ---
	mcpxTestServerInfo := integration.StartMCPXServer(t, "E2ECowsayServerTest")
	defer mcpxTestServerInfo.CleanupFunc()

	// --- 2. Register 'cowsay' server with MCPX ---
	const cowsayServiceID = "e2e-cowsay-server"
	command := "python"
	args := []string{"-u", "main.py", "--mcp-stdio"}
	setupCommands := []string{"pip install cowsay"}

	t.Logf("INFO: Registering '%s' with MCPX...", cowsayServiceID)
	registrationGRPCClient := mcpxTestServerInfo.RegistrationClient
	integration.RegisterStdioServiceWithSetup(
		t,
		registrationGRPCClient,
		cowsayServiceID,
		command,
		true,
		"tests/integration/cmd/mocks/python_cowsay_server", // working directory
		"", // No explicit container image
		setupCommands,
		args...,
	)
	t.Logf("INFO: '%s' registered with command: %s %v", cowsayServiceID, command, args)

	// --- 3. Use MCP SDK to connect and call the tool ---
	testMCPClient := sdk.NewClient(&sdk.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &sdk.StreamableClientTransport{Endpoint: mcpxTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err, "Failed to connect to MCPX server")
	defer cs.Close()

	toolName := fmt.Sprintf("%s/-/say", cowsayServiceID)

	// Wait for the tool to be available
	require.Eventually(t, func() bool {
		result, err := cs.ListTools(ctx, &sdk.ListToolsParams{})
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
	}, integration.TestWaitTimeLong, 5*time.Second, "Tool %s did not become available in time", toolName)

	params := json.RawMessage(`{"message": "hello from docker"}`)

	res, err := cs.CallTool(ctx, &sdk.CallToolParams{Name: toolName, Arguments: params})
	require.NoError(t, err, "Error calling tool '%s'", toolName)
	require.NotNil(t, res, "Nil response from tool '%s'", toolName)
	require.Len(t, res.Content, 1, "Expected exactly one content item from tool '%s'", toolName)

	textContent, ok := res.Content[0].(*sdk.TextContent)
	require.True(t, ok, "Expected content to be of type TextContent")

	// Log the raw text content for debugging
	t.Logf("Raw tool output:\n%s", textContent.Text)

	require.True(t, strings.Contains(textContent.Text, "hello from docker"), "Output should contain the message")
	require.True(t, strings.Contains(textContent.Text, "< hello from docker >"), "Output should be from cowsay")

	t.Logf("SUCCESS: Call to tool '%s' via MCPX completed.", toolName)
	t.Log("INFO: E2E Test Scenario for 'cowsay' server (Stdio via Docker with setup) Completed Successfully!")
}
