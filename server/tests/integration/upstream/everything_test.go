// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package upstream contains integration tests for upstream services.
package upstream

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mcpany/core/server/pkg/util"
	"github.com/mcpany/core/server/tests/framework"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_HTTP_Everything(t *testing.T) {
	testCase := &framework.E2ETestCase{
		Name:                "HTTP Everything",
		UpstreamServiceType: "streamablehttp",
		BuildUpstream:       framework.BuildEverythingServer,
		RegisterUpstream:    framework.RegisterEverythingService,
		ValidateTool: func(t *testing.T, mcpanyEndpoint string) {
			serviceID := "e2e_everything_server_streamable"
			toolName := "add"
			serviceID, err := util.SanitizeServiceName(serviceID)
			require.NoError(t, err)
			sanitizedToolName, err := util.SanitizeToolName(toolName)
			require.NoError(t, err)
			expectedToolName := serviceID + "." + sanitizedToolName

			expectedTool := &mcp.Tool{
				Name:        expectedToolName,
				Description: "Adds two integers.",
			}
			framework.ValidateRegisteredTool(t, mcpanyEndpoint, expectedTool)
		},
		InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
			ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
			defer cancel()

			testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
			cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpanyEndpoint}, nil)
			require.NoError(t, err)
			defer func() { _ = cs.Close() }()

			serviceID, _ := util.SanitizeServiceName("e2e_everything_server_streamable")
			sanitizedToolName, _ := util.SanitizeToolName("add")
			toolName := serviceID + "." + sanitizedToolName
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(`{"a": 10, "b": 15}`)})
			require.NoError(t, err)
			require.NotNil(t, res)
			require.Len(t, res.Content, 1)
			textContent, ok := res.Content[0].(*mcp.TextContent)
			require.True(t, ok)
			require.Contains(t, textContent.Text, "25")
		},
	}

	framework.RunE2ETest(t, testCase)
}
