// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package upstream

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/util"
	"github.com/mcpany/core/server/tests/framework"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_HTTP_HealthCheck(t *testing.T) {
	const echoServiceID = "e2e_http_echo_health"
	var upstream *integration.ManagedProcess

	testCase := &framework.E2ETestCase{
		Name:                "HTTP Echo Server with Health Check",
		UpstreamServiceType: "http",
		BuildUpstream: func(t *testing.T) *integration.ManagedProcess {
			upstream = framework.BuildHTTPEchoServer(t)
			return upstream
		},
		GenerateUpstreamConfig: func(upstreamEndpoint string) string {
			return fmt.Sprintf(`
name: %s
connection_pool:
  max_connections: 1
http_service:
  address: %s
  health_check:
    url: %s/health
    expected_code: 200
    interval: 1s
    timeout: 1s
  calls:
  - schema:
      name: echo
      description: Echoes a message back to the user.
    endpoint_path: /echo
    method: HTTP_METHOD_POST
`, echoServiceID, upstreamEndpoint, upstreamEndpoint)
		},
		InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
			ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
			defer cancel()

			testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
			cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpanyEndpoint}, nil)
			require.NoError(t, err)
			defer func() { _ = cs.Close() }()

			serviceID, _ := util.SanitizeServiceName(echoServiceID)
			sanitizedToolName, _ := util.SanitizeToolName("echo")
			toolName := serviceID + "." + sanitizedToolName
			echoMessage := `{"message": "hello world"}`

			// 1. Initial successful call
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(echoMessage)})
			require.NoError(t, err)
			require.NotNil(t, res)
			textContent, ok := res.Content[0].(*mcp.TextContent)
			require.True(t, ok)
			require.JSONEq(t, echoMessage, textContent.Text)

			// 2. Stop the upstream service
			upstream.Stop()

			// 3. Expect a failure
			_, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(echoMessage)})
			require.Error(t, err)

			// 4. Restart the upstream service
			_ = upstream.Start()

			// 5. Expect a successful call again
			require.Eventually(t, func() bool {
				res, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(echoMessage)})
				return err == nil
			}, 5*time.Second, 1*time.Second, "tool did not become healthy again")
			require.NoError(t, err)
			require.NotNil(t, res)
			textContent, ok = res.Content[0].(*mcp.TextContent)
			require.True(t, ok)
			require.JSONEq(t, echoMessage, textContent.Text)
		},
	}
	framework.RunE2ETest(t, testCase)
}
