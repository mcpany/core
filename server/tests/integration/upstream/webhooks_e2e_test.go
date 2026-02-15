package upstream_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream/mcp"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	TIMESTAMPCB "google.golang.org/protobuf/types/known/durationpb"
)

func TestWebhooksE2E(t *testing.T) {
	// Build the webhook server
	rootDir := findRootDir(t)
	webhookBin := filepath.Join(rootDir, "build", "bin", "webhooks")
	cmd := exec.Command("go", "build", "-o", webhookBin, "./cmd/webhooks") //nolint:gosec
	cmd.Dir = rootDir
	require.NoError(t, cmd.Run(), "Failed to build webhook server")

	// Start webhook server
	// Start webhook server
	port := getFreePort(t)
	portStr := fmt.Sprintf("%d", port)

	const secret = "dGVzdC1zZWNyZXQtMTIz" //nolint:gosec // base64("test-secret-123")
	secretPtr := secret                   // Create addressable variable
	serverCmd := exec.Command(webhookBin) //nolint:gosec
	serverCmd.Stdout = os.Stdout
	serverCmd.Stderr = os.Stderr
	serverCmd.Env = append(os.Environ(), "WEBHOOK_SECRET="+secret, "PORT="+portStr, "MCPANY_ALLOW_LOOPBACK_RESOURCES=true")
	require.NoError(t, serverCmd.Start(), "Failed to start webhook server")
	defer func() {
		_ = serverCmd.Process.Kill()
	}()

	// Wait for server to start
	require.Eventually(t, func() bool {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/markdown", port)) // Endpoint exists (POST only but connectable)
		if resp != nil {
			defer func() { _ = resp.Body.Close() }()
		}
		return err == nil && (resp.StatusCode == 405 || resp.StatusCode == 200 || resp.StatusCode == 401)
	}, 5*time.Second, 100*time.Millisecond, "Webhook server failed to start")

	t.Run("MarkdownConversion", func(t *testing.T) {
		url := fmt.Sprintf("http://127.0.0.1:%d/markdown", port)
		hook := tool.NewWebhookHook(configv1.WebhookConfig_builder{
			Url:           url,
			Timeout:       TIMESTAMPCB.New(5 * time.Second),
			WebhookSecret: secretPtr,
		}.Build())

		ctx := context.Background()
		req := &tool.ExecutionRequest{
			ToolName: "test-tool",
		}

		html := "<h1>Hello World</h1><p>Test</p>"
		result, err := hook.ExecutePost(ctx, req, html) // Pass string directly
		require.NoError(t, err)

		// Expecting result to be struct with "value" if wrapped?
		// My hooks implementation wraps non-map results in "value" when sending.
		// If implementation returns full map, ExecutePost logic extracts "value" if original was not map.

		// The webhook server converts "value" to markdown.
		// "<h1>Hello World</h1><p>Test</p>" -> "# Hello World\n\nTest"

		markdown, ok := result.(string)
		if !ok {
			// It might have returned a map if wrapping logic changed
			// Let's debug
			t.Logf("Result type: %T, Value: %+v", result, result)
			// Try to extract if it's map
			if m, ok := result.(map[string]any); ok {
				if v, ok := m["value"]; ok {
					markdown = fmt.Sprintf("%v", v)
				}
			}
		}

		assert.Contains(t, markdown, "# Hello World")
		assert.Contains(t, markdown, "Test")
	})

	t.Run("TextTruncation", func(t *testing.T) {
		url := fmt.Sprintf("http://127.0.0.1:%d/truncate?max_chars=5", port)
		hook := tool.NewWebhookHook(configv1.WebhookConfig_builder{
			Url:           url,
			Timeout:       TIMESTAMPCB.New(5 * time.Second),
			WebhookSecret: secretPtr,
		}.Build())

		ctx := context.Background()
		req := &tool.ExecutionRequest{
			ToolName: "test-tool",
		}

		longText := "This is a very long text"
		result, err := hook.ExecutePost(ctx, req, longText)
		require.NoError(t, err)

		truncated, ok := result.(string)
		if !ok {
			t.Logf("Result type: %T, Value: %+v", result, result)
			if m, ok := result.(map[string]any); ok {
				if v, ok := m["value"]; ok {
					truncated = fmt.Sprintf("%v", v)
				}
			}
		}

		assert.Equal(t, "This ...", truncated)
	})
}

func TestFullSystemWebhooks(t *testing.T) {
	// 1. Build Webhook Server
	rootDir := findRootDir(t)
	webhookBin := filepath.Join(rootDir, "build", "bin", "webhooks")
	cmd := exec.Command("go", "build", "-o", webhookBin, "./cmd/webhooks") //nolint:gosec
	cmd.Dir = rootDir
	require.NoError(t, cmd.Run(), "Failed to build webhook server")

	// 2. Build Mock MCP Server
	mockMcpBin := filepath.Join(rootDir, "build", "bin", "mock_mcp")
	cmd = exec.Command("go", "build", "-o", mockMcpBin, "./tests/integration/upstream/testdata/mock_mcp") //nolint:gosec
	cmd.Dir = rootDir
	require.NoError(t, cmd.Run(), "Failed to build mock MCP server")

	// 3. Start Webhook Server
	port := getFreePort(t)
	portStr := fmt.Sprintf("%d", port)

	const secret = "dGVzdC1zZWNyZXQtMTIz" //nolint:gosec
	serverCmd := exec.Command(webhookBin) //nolint:gosec
	serverCmd.Stdout = os.Stdout
	serverCmd.Stderr = os.Stderr
	serverCmd.Env = append(os.Environ(), "WEBHOOK_SECRET="+secret, "PORT="+portStr, "MCPANY_ALLOW_LOOPBACK_RESOURCES=true")
	require.NoError(t, serverCmd.Start(), "Failed to start webhook server")
	defer func() { _ = serverCmd.Process.Kill() }()

	// Wait for webhook server
	require.Eventually(t, func() bool {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/markdown", port))
		if resp != nil && resp.Body != nil {
			defer func() { _ = resp.Body.Close() }()
		}
		return err == nil
	}, 5*time.Second, 100*time.Millisecond)

	// 4. Configure Upstream Service (Mcpany Core Logic)
	webhookURL := fmt.Sprintf("http://127.0.0.1:%d/markdown", port)

	upsConfig := configv1.UpstreamServiceConfig_builder{
		Name:             proto.String("mock-service"),
		AutoDiscoverTool: proto.Bool(true),
		McpService: configv1.McpUpstreamService_builder{
			StdioConnection: configv1.McpStdioConnection_builder{
				Command: proto.String(mockMcpBin),
			}.Build(),
		}.Build(),
		PostCallHooks: []*configv1.CallHook{
			configv1.CallHook_builder{
				Name: proto.String("markdown-converter"),
				Webhook: configv1.WebhookConfig_builder{
					Url:           webhookURL,
					WebhookSecret: secret,
					Timeout:       TIMESTAMPCB.New(5 * time.Second),
				}.Build(),
			}.Build(),
		},
	}.Build()

	toolManager := tool.NewManager(nil)
	ctx := context.Background()
	upstreamService := mcp.NewUpstream(nil)

	// Register service
	serviceID, _, _, err := upstreamService.Register(
		ctx,
		upsConfig,
		toolManager,
		nil, // prompt manager
		nil, // resource manager
		false,
	)
	require.NoError(t, err, "Failed to register upstream service")

	// 5. Execute Tool
	toolID := serviceID + ".get_html"

	// Use Manager.ExecuteTool to trigger hooks
	mcpReq := &tool.ExecutionRequest{
		ToolName:   toolID,
		ToolInputs: json.RawMessage(`{}`),
	}

	resultCallTool, err := toolManager.ExecuteTool(ctx, mcpReq)
	require.NoError(t, err, "Failed to execute tool")

	// Unwrap the result from CallToolResult
	var result any
	if callToolRes, ok := resultCallTool.(*mcpsdk.CallToolResult); ok {
		if len(callToolRes.Content) > 0 {
			if text, ok := callToolRes.Content[0].(*mcpsdk.TextContent); ok {
				result = text.Text
			}
		}
	} else {
		result = resultCallTool
	}

	// 6. Verify Result
	t.Logf("Result: %v", result)

	var resultStr string
	if s, ok := result.(string); ok {
		resultStr = s
	} else if m, ok := result.(map[string]any); ok {
		if v, ok := m["value"]; ok {
			resultStr = fmt.Sprintf("%v", v)
		} else {
			// Fallback json dump
			b, _ := json.Marshal(result)
			resultStr = string(b)
		}
	}



	assert.Contains(t, resultStr, "# Mock Title")
	assert.Contains(t, resultStr, "Mock content")
}

func getFreePort(t *testing.T) int {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	l, err := net.ListenTCP("tcp", addr)
	require.NoError(t, err)
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func findRootDir(t *testing.T) string {
	dir, err := os.Getwd()
	require.NoError(t, err)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("Root directory not found")
		}
		dir = parent
	}
}
