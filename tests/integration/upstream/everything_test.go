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
