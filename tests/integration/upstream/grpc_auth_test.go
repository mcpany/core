/*
 * Copyright 2025 Author(s) of MCP Any
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

	"github.com/mcpany/core/pkg/util"
	"github.com/mcpany/core/tests/framework"
	"github.com/mcpany/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_GRPC_WithBearerAuth(t *testing.T) {
	testCase := &framework.E2ETestCase{
		Name:                "Authenticated gRPC Weather Server",
		UpstreamServiceType: "grpc",
		BuildUpstream:       framework.BuildGRPCAuthedWeatherServer,
		RegisterUpstream:    framework.RegisterGRPCAuthedWeatherService,
		InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
			ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
			defer cancel()

			testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
			cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpanyEndpoint}, nil)
			require.NoError(t, err)
			defer func() { _ = cs.Close() }()

			const weatherServiceID = "e2e_grpc_authed_weather"
			serviceID, _ := util.SanitizeServiceName(weatherServiceID)
			sanitizedToolName, _ := util.SanitizeToolName("GetWeather")
			toolName := serviceID + "." + sanitizedToolName
			weatherArgs := `{"location": "london"}`
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(weatherArgs)})
			require.NoError(t, err, "Error calling GetWeather tool with correct auth")
			require.NotNil(t, res, "Nil response from GetWeather tool with correct auth")
			switch content := res.Content[0].(type) {
			case *mcp.TextContent:
				require.JSONEq(t, `{"weather": "Cloudy, 15°C"}`, content.Text, "The weather is incorrect")
			default:
				t.Fatalf("Unexpected content type: %T", content)
			}
		},
	}

	framework.RunE2ETest(t, testCase)
}
