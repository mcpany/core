// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package upstream

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/mcpany/core/server/tests/framework"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_OpenAPI(t *testing.T) {
	// Replaced Gemini CLI with direct MCP client to remove external dependency and auth requirement
	// gemini := framework.NewGeminiCLI(t)
	// gemini.Install()

	// apiKey := os.Getenv("GEMINI_API_KEY")
	// if apiKey == "" {
	// 	t.Skip("GEMINI_API_KEY is not set")
	// }

	testCase := &framework.E2ETestCase{
		Name:                "OpenAPI Weather Server",
		UpstreamServiceType: "openapi",
		BuildUpstream:       framework.BuildOpenAPIWeatherServer,
		RegisterUpstream:    framework.RegisterOpenAPIWeatherService,
		InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
			framework.VerifyMCPClient(t, mcpanyEndpoint)

			ctx := context.Background()
			client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0.0"}, nil)
			transport := &mcp.StreamableClientTransport{Endpoint: mcpanyEndpoint}
			session, err := client.Connect(ctx, transport, nil)
			require.NoError(t, err)
			defer session.Close()

			// List tools to find the weather tool
			tools, err := session.ListTools(ctx, &mcp.ListToolsParams{})
			require.NoError(t, err)
			var toolName string
			for _, tool := range tools.Tools {
				// The tool name typically contains "weather" or "get_weather"
				// Service ID is "e2e_openapi_weather"
				if strings.Contains(tool.Name, "weather") || strings.Contains(tool.Name, "Weather") {
					toolName = tool.Name
					break
				}
			}
			require.NotEmpty(t, toolName, "Weather tool not found in list: %v", tools.Tools)

			t.Logf("Calling tool: %s", toolName)
			// Call tool
			// Assuming arguments based on common weather demos: location (string)
			args := map[string]string{"location": "London"}
			argBytes, _ := json.Marshal(args)
			res, err := session.CallTool(ctx, &mcp.CallToolParams{
				Name:      toolName,
				Arguments: json.RawMessage(argBytes),
			})
			require.NoError(t, err)
			require.NotNil(t, res)
			require.False(t, res.IsError)
			require.NotEmpty(t, res.Content)

			// Verify content contains expected weather info (mocked or real from local server)
			// The local server usually returns "Cloudy, 15°C" or similar
			found := false
			for _, content := range res.Content {
				if txt, ok := content.(*mcp.TextContent); ok {
					t.Logf("Tool Output: %s", txt.Text)
					if strings.Contains(txt.Text, "Cloudy") || strings.Contains(txt.Text, "15") {
						found = true
					}
				}
			}
			require.True(t, found, "Expected weather content not found")
		},
	}

	framework.RunE2ETest(t, testCase)
}
