// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package popular_service_test

import (
	"context"
	"os"
	"testing"

	"github.com/mcpany/core/server/pkg/metrics"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	metrics.Initialize()
	os.Exit(m.Run())
}

func TestIPInfoService(t *testing.T) {
	t.Setenv("IPINFO_API_TOKEN", os.Getenv("IPINFO_API_TOKEN"))
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for ipinfo.io Service...")

	// --- 1. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EIPInfoServiceTest", "--config-path", "../../../examples/popular_services/ipinfo.io")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 2. Call Tool via MCPANY ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpAnyTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	callToolResult, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "ipinfo.io.ipinfo",
		Arguments: map[string]any{
			"ip": "8.8.8.8",
		},
	})
	require.NoError(t, err)

	// --- 3. Assert Response ---
	require.NotNil(t, callToolResult)
	textContent, ok := callToolResult.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	require.Contains(t, textContent.Text, "dns.google")
	require.Contains(t, textContent.Text, "abuse")
}
