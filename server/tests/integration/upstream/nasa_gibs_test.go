package upstream

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/util"
	"github.com/mcpany/core/server/tests/framework"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_NASAGIBS(t *testing.T) {
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)

	testCase := &framework.E2ETestCase{
		Name:                "NASA GIBS Example",
		UpstreamServiceType: "http",
		BuildUpstream: func(_ *testing.T) *integration.ManagedProcess {
			// NASA GIBS is a real external service, so no local process to build/start.
			return nil
		},
		GenerateUpstreamConfig: func(_ string) string {
			configPath := filepath.Join(root, "examples", "popular_services", "nasa", "config.yaml")
			content, err := os.ReadFile(configPath) //nolint:gosec // Test file
			require.NoError(t, err)
			return string(content)
		},
		InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
			ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
			defer cancel()

			testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
			cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpanyEndpoint}, nil)
			require.NoError(t, err, "Failed to connect to MCPANY server")
			defer func() { _ = cs.Close() }()

			serviceID, _ := util.SanitizeServiceName("nasa-gibs")
			sanitizedToolName, _ := util.SanitizeToolName("get_tile")
			toolName := serviceID + "." + sanitizedToolName
			// Wait for the tool to be available
			require.Eventually(t, func() bool {
				result, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
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
			}, integration.TestWaitTimeMedium, 1*time.Second, "Tool %s did not become available in time", toolName)

			params := json.RawMessage(`{"LayerIdentifier": "MODIS_Terra_CorrectedReflectance_TrueColor", "Time": "2012-07-09", "TileMatrixSet": "250m", "TileMatrix": "6", "TileRow": "13", "TileCol": "36"}`)

			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: params})
			require.NoError(t, err, "Error calling tool '%s'", toolName)
			require.NotNil(t, res, "Nil response from tool '%s'", toolName)
			require.Len(t, res.Content, 1, "Expected exactly one content item from tool '%s'", toolName)

			textContent, ok := res.Content[0].(*mcp.TextContent)
			require.True(t, ok, "Expected content to be of type TextContent")
			t.Logf("Tool output: %s", textContent.Text)
		},
	}

	framework.RunE2ETest(t, testCase)
}
