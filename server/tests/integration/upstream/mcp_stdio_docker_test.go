// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package upstream

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mcpany/core/server/tests/framework"
	"github.com/mcpany/core/server/tests/integration"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_MCP_Stdio_WithSetupCommandsInDocker(t *testing.T) {
	if !integration.IsDockerSocketAccessible() {
		t.Skip("Docker socket not accessible, skipping test")
	}
	if os.Getenv("CI") != "" {
		t.Skip("Skipping Docker-in-Docker test in CI environment")
	}

	testCase := &framework.E2ETestCase{
		Name:                "cowsay server (Stdio via Docker with setup)",
		UpstreamServiceType: "stdio",
		BuildUpstream:       framework.BuildStdioDockerServer,
		RegisterUpstream:    framework.RegisterStdioDockerService,
		InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
			ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
			defer cancel()

			testMCPClient := sdk.NewClient(&sdk.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
			cs, err := testMCPClient.Connect(ctx, &sdk.StreamableClientTransport{Endpoint: mcpanyEndpoint}, nil)
			require.NoError(t, err, "Failed to connect to MCPANY server")
			defer func() { _ = cs.Close() }()

			toolName := "e2e-cowsay-server/-/say"

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
		},
	}

	framework.RunE2ETest(t, testCase)
}
