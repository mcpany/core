// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package upstream

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/mcpany/core/server/tests/framework"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_HTTP_InputSchemaFix(t *testing.T) {
	testCase := &framework.E2ETestCase{
		Name:                "HTTP Input Schema Fix",
		UpstreamServiceType: "http",
		BuildUpstream:       framework.BuildHTTPEchoServer,
		GenerateUpstreamConfig: func(upstreamAddress string) string {
            // upstreamAddress is like "127.0.0.1:1234"
			return fmt.Sprintf(`
upstream_services:
  - name: schema-fix-service
    http_service:
      address: http://%s
      tools:
        - name: echo
          call_id: echo_call
      calls:
        echo_call:
          id: echo_call
          method: HTTP_METHOD_POST
          endpoint_path: /echo
          input_schema:
            properties:
              message:
                type: string
`, upstreamAddress)
		},
		InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
			ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
			defer cancel()

			testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
			cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpanyEndpoint}, nil)
			require.NoError(t, err)
			defer func() { _ = cs.Close() }()

            // Wait for tool
            toolName := "schema-fix-service.echo"
			require.Eventually(t, func() bool {
				result, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
				if err != nil {
					return false
				}
				for _, tool := range result.Tools {
					if tool.Name == toolName {
                        // Check schema type
                        schema, ok := tool.InputSchema.(map[string]interface{})
                        if !ok {
                            t.Logf("InputSchema is not a map")
                            return false
                        }
                        typeVal, ok := schema["type"].(string)
                        if !ok || typeVal != "object" {
                            t.Logf("InputSchema type is %v, expected object", schema["type"])
                            return false
                        }
						return true
					}
				}
				return false
			}, integration.TestWaitTimeMedium, 100*time.Millisecond, "Tool %s did not appear or had incorrect schema", toolName)
		},
        RegistrationMethods: []framework.RegistrationMethod{framework.FileRegistration},
	}

	framework.RunE2ETest(t, testCase)
}
