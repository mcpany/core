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

func TestUpstreamService_HTTP(t *testing.T) {
	testCase := &framework.E2ETestCase{
		Name:                "HTTP Echo Server",
		UpstreamServiceType: "http",
		BuildUpstream:       framework.BuildHTTPEchoServer,
		RegisterUpstream:    framework.RegisterHTTPEchoService,
		InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
			ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
			defer cancel()

			testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
			cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpanyEndpoint}, nil)
			require.NoError(t, err)
			defer func() { _ = cs.Close() }()

			listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
			require.NoError(t, err)
			for _, tool := range listToolsResult.Tools {
				t.Logf("Discovered tool from MCPANY: %s", tool.Name)
			}

			const echoServiceID = "e2e_http_echo"
			serviceID, _ := util.SanitizeServiceName(echoServiceID)
			sanitizedToolName, _ := util.SanitizeToolName("echo")
			toolName := serviceID + "." + sanitizedToolName
			echoMessage := `{"message": "hello world from http"}`
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(echoMessage)})
			require.NoError(t, err, "Error calling echo tool")
			require.NotNil(t, res, "Nil response from echo tool")
			switch content := res.Content[0].(type) {
			case *mcp.TextContent:
				require.JSONEq(t, echoMessage, content.Text, "The echoed message does not match the original")
			default:
				t.Fatalf("Unexpected content type: %T", content)
			}
		},
	}

	framework.RunE2ETest(t, testCase)
}

func TestUpstreamService_HTTPExample(t *testing.T) {
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)

	testCase := &framework.E2ETestCase{
		Name:                "HTTP IP Info Example",
		UpstreamServiceType: "http",
		BuildUpstream: func(_ *testing.T) *integration.ManagedProcess {
			// This test assumes an external service is already running or mocked at the URL.
			// For the E2E framework, we might need a dummy process or nil if using external URL.
			return nil
		},
		GenerateUpstreamConfig: func(_ string) string {
			configPath := filepath.Join(root, "examples", "upstream", "http", "config", "mcpany_config.yaml")
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

			serviceID, _ := util.SanitizeServiceName("ip-info-service")
			sanitizedToolName, _ := util.SanitizeToolName("get_time_by_ip")
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

			params := json.RawMessage(`{"ip": "8.8.8.8"}`)

			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: params})
			require.NoError(t, err, "Error calling tool '%s'", toolName)
			require.NotNil(t, res, "Nil response from tool '%s'", toolName)
			require.Len(t, res.Content, 1, "Expected exactly one content item from tool '%s'", toolName)

			textContent, ok := res.Content[0].(*mcp.TextContent)
			require.True(t, ok, "Expected content to be of type TextContent")
			t.Logf("Tool output: %s", textContent.Text)

			var ipInfoResponse struct {
				IP       string `json:"ip"`
				Hostname string `json:"hostname"`
				City     string `json:"city"`
				Region   string `json:"region"`
				Country  string `json:"country"`
				Org      string `json:"org"`
			}

			err = json.Unmarshal([]byte(textContent.Text), &ipInfoResponse)
			require.NoError(t, err, "Failed to unmarshal ipinfo.io response")

			require.Equal(t, "8.8.8.8", ipInfoResponse.IP, "Expected IP to be 8.8.8.8")
			require.Equal(t, "dns.google", ipInfoResponse.Hostname, "Expected hostname to be dns.google")
			require.Equal(t, "Mountain View", ipInfoResponse.City, "Expected city to be Mountain View")
			require.Equal(t, "California", ipInfoResponse.Region, "Expected region to be California")
			require.Equal(t, "US", ipInfoResponse.Country, "Expected country to be US")
			require.Contains(t, ipInfoResponse.Org, "Google", "Expected org to contain Google")
		},
	}

	framework.RunE2ETest(t, testCase)
}
