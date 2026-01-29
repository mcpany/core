//go:build e2e_public_api

package public_api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestPopularServices_Load(t *testing.T) {
	// t.Parallel() removed because t.Setenv is used

	// 1. Setup Mock OpenAPI Server (Hermetic)
	// We serve a minimal valid OpenAPI spec so that "spec_url" configs can load successfully without internet.
	dummySpec := `
openapi: 3.0.0
info:
  title: Dummy Service
  version: 1.0.0
servers:
  - url: http://127.0.0.1:8080
paths:
  /status:
    get:
      operationId: getStatus
      responses:
        '200':
          description: OK
`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		w.Write([]byte(dummySpec))
	}))
	defer ts.Close()

	// 2. Identify Mock MCP Server binary (Hermetic)
	// We use the locally built mock_server instead of npx/npm/puppeteer which requires internet.
	projectRoot := integration.ProjectRoot(t)
	mockServerPath := filepath.Join(projectRoot, "..", "build", "bin", "mock_server")
	if _, err := os.Stat(mockServerPath); os.IsNotExist(err) {
		t.Logf("Mock server not found at %s. Attempting to build or strictly failing if expected.", mockServerPath)
		// For robustness, we could build it here, but it should be built by previous steps.
		// If missing, we might fail the "chrome" test specifically or fall back.
		// Let's assume it exists as per previous context.
	}

	testCases := []struct {
		serviceName string
		configFile  string
	}{
		{"github", "popular_services/github/config.yaml"},
		{"stripe", "popular_services/stripe/config.yaml"},
		{"jira", "popular_services/jira/config.yaml"},
		{"slack", "popular_services/slack/config.yaml"},
		{"google_calendar", "popular_services/google/calendar/config.yaml"},
		{"spotify", "popular_services/spotify/config.yaml"},
		{"gitlab", "popular_services/gitlab/config.yaml"},
		{"chrome", "popular_services/chrome/config.yaml"},
	}

	for _, tc := range testCases {
		t.Run(tc.serviceName, func(t *testing.T) {
			// t.Parallel() removed because t.Setenv is used

			// Read the example config file
			configPath := filepath.Join(projectRoot, "examples", tc.configFile)
			configBytes, err := os.ReadFile(configPath)
			require.NoError(t, err, "Failed to read config file: %s", configPath)
			configContent := string(configBytes)

			// HERMETIC MODIFICATIONS:
			// 1. Replace spec_url with local mock server
			if strings.Contains(configContent, "spec_url:") {
				// We replace the entire URL line with our local one.
				// This assumes standard formatting "spec_url: <url>".
				// We use regex to be safe or just simple replacement if we identify the URL.
				// A simple way is to replace the URL value.
				// But we don't know the exact URL easily without parsing.
				// We can just use "spec_url: http..." replacement.
				// Let's rely on the fact we know the keys.
				// BETTER: Use regexp to replace "spec_url: \".*\"" with "spec_url: \"<ts.URL>\""
				re := regexp.MustCompile(`spec_url:.*`)
				configContent = re.ReplaceAllString(configContent, fmt.Sprintf("spec_url: \"%s\"", ts.URL))
			}

			// 2. Replace npx command with mock_server
			if strings.Contains(configContent, "command: \"npx\"") {
				configContent = strings.ReplaceAll(configContent, "command: \"npx\"", fmt.Sprintf("command: \"%s\"", mockServerPath))
				reArgs := regexp.MustCompile(`args: \[.*\]`)
				configContent = reArgs.ReplaceAllString(configContent, "args: []")
				if _, err := os.Stat(mockServerPath); err != nil {
					t.Fatalf("Mock server binary missing: %v", err)
				} else {
					// Replace npx command with the mock server binary
					// We want to verify that the config structure is correct and that the "command" field is respected.
					// The mock server will just run and exit or listen depending on args, but for "Load" test we just need it to start.
				}
			}

			// Supply dummy values for valid config loading/validation.
			t.Setenv("GITHUB_TOKEN", "dummy")
			t.Setenv("STRIPE_API_KEY", "dummy")
			t.Setenv("JIRA_USERNAME", "dummy")
			t.Setenv("JIRA_PAT", "dummy")
			t.Setenv("JIRA_DOMAIN", "jira.example.com")
			t.Setenv("SLACK_API_TOKEN", "dummy")
			t.Setenv("GOOGLE_API_KEY", "dummy")
			t.Setenv("GOOGLE_OAUTH_CLIENT_ID", "dummy")
			t.Setenv("GOOGLE_OAUTH_CLIENT_SECRET", "dummy")
			t.Setenv("GOOGLE_OAUTH_REFRESH_TOKEN", "dummy")
			t.Setenv("SPOTIFY_TOKEN", "dummy")
			t.Setenv("GITLAB_TOKEN", "dummy")

			// Inject auto_discover_tool: true for MCP services (using mock_server)
			if strings.Contains(configContent, "mcp_service:") {
				if !strings.Contains(configContent, "auto_discover_tool:") {
					configContent = strings.Replace(configContent, "mcp_service:", "auto_discover_tool: true\n    mcp_service:", 1)
				}
			}

			// Start server with the modified config
			serverInfo := integration.StartMCPANYServerWithConfig(t, "PopularService_"+tc.serviceName, configContent)
			defer serverInfo.CleanupFunc()

			// Check if tools are discoverable
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second) // Fast timeout for hermetic test
			defer cancel()

			// Connect using MCP Client SDK
			testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
			cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: serverInfo.HTTPEndpoint}, nil)
			require.NoError(t, err)
			defer cs.Close()

			// 3. Tools List with Retry
			require.Eventually(t, func() bool {
				toolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
				if err != nil {
					return false
				}
				if len(toolsResult.Tools) == 0 {
					return false
				}
				t.Logf("Discovered %d tools for %s", len(toolsResult.Tools), tc.serviceName)
				return true
			}, 10*time.Second, 500*time.Millisecond, "Timed out waiting for tools to be discovered") // Fast retry

			// Re-fetch final result for assertions
			toolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
			require.NoError(t, err)
			require.NotEmpty(t, toolsResult.Tools, "Tools list should not be empty")
			// For OpenAPI mock, we expect "getStatus" (from dummySpec).
			// For mock_server (chrome replacement), we expect "read_file" / "list_directory".
			// We can assert specifically if we want, but "NotEmpty" is good enough for "Load" test.
		})
	}
}
