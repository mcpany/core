/*
 * Copyright 2025 Author(s) of MCP-XY
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

package upstream

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mcpxy/core/pkg/util"
	"github.com/mcpxy/core/tests/framework"
	"github.com/mcpxy/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_GRPC_WithBearerAuth(t *testing.T) {
	testCase := &framework.E2ETestCase{
		Name:                "Authenticated gRPC Calculator Server",
		UpstreamServiceType: "grpc",
		BuildUpstream:       framework.BuildGRPCAuthedServer,
		RegisterUpstream:    framework.RegisterGRPCAuthedService,
		InvokeAIClient: func(t *testing.T, mcpxyEndpoint string) {
			ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
			defer cancel()

			testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
			cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpxyEndpoint}, nil)
			require.NoError(t, err)
			defer cs.Close()

			const calcServiceID = "e2e_grpc_authed_calculator"
			serviceKey, _ := util.SanitizeServiceName(calcServiceID)
			toolName, _ := util.SanitizeToolName(serviceKey, "CalculatorAdd")
			addArgs := `{"a": 10, "b": 20}`
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(addArgs)})
			require.NoError(t, err, "Error calling Add tool with correct auth")
			require.NotNil(t, res, "Nil response from Add tool with correct auth")
			switch content := res.Content[0].(type) {
			case *mcp.TextContent:
				require.JSONEq(t, `{"result": 30}`, content.Text, "The sum is incorrect")
			default:
				t.Fatalf("Unexpected content type: %T", content)
			}
		},
	}

	framework.RunE2ETest(t, testCase)
}
