// Copyright 2024 Author(s) of MCP Any
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package popular_service_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mcpany/core/tests/framework"
	"github.com/mcpany/core/tests/integration"
	"github.com/mcpany/core/tests/integration/upstream"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestStripe(t *testing.T) {
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)

	testCase := &framework.E2ETestCase{
		Name:          "Stripe Create Customer",
		BuildUpstream: upstream.BuildStripeMockServer,
		GenerateUpstreamConfig: func(upstreamEndpoint string) string {
			configPath := filepath.Join(root, "examples", "popular_services", "stripe", "config.yaml")
			content, err := os.ReadFile(configPath)
			require.NoError(t, err)

			return strings.Replace(string(content), "https://api.stripe.com", upstreamEndpoint, 1)
		},
		InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
			ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
			defer cancel()

			testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
			cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpanyEndpoint}, nil)
			require.NoError(t, err)
			defer cs.Close()

			toolName := "stripe.create_customer"
			arguments := json.RawMessage(`{"name": "John Doe", "email": "john.doe@example.com"}`)
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: arguments})
			require.NoError(t, err, "Error calling tool")
			require.NotNil(t, res, "Nil response from tool")

			textContent, ok := res.Content[0].(*mcp.TextContent)
			require.True(t, ok, "Expected content to be of type TextContent")

			var result map[string]interface{}
			err = json.Unmarshal([]byte(textContent.Text), &result)
			require.NoError(t, err)
			require.Equal(t, "John Doe", result["name"])
			require.Equal(t, "john.doe@example.com", result["email"])
			require.Equal(t, "cus_12345", result["id"])
		},
	}

	framework.RunE2ETest(t, testCase)
}
