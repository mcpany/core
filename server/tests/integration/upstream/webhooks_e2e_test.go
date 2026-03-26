// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

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
	"github.com/mcpany/core/server/tests/integration"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	TIMESTAMPCB "google.golang.org/protobuf/types/known/durationpb"
)

func runfileBinary(relParts ...string) string {
	workspace := os.Getenv("TEST_WORKSPACE")
	if workspace == "" {
		workspace = "_main"
	}
	for _, base := range []string{os.Getenv("TEST_SRCDIR"), os.Getenv("RUNFILES_DIR")} {
		if base == "" {
			continue
		}
		candidate := filepath.Join(append([]string{base, workspace}, relParts...)...)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ""
}

func webhookBinary(t *testing.T) string {
	t.Helper()
	if bin := runfileBinary("server", "cmd", "webhooks", "webhooks_", "webhooks"); bin != "" {
		return bin
	}
	rootDir := integration.ProjectRoot(t)
	webhookBin := filepath.Join(t.TempDir(), "webhooks")
	cmd := exec.Command("go", "build", "-o", webhookBin, "./cmd/webhooks") //nolint:gosec
	cmd.Dir = rootDir
	require.NoError(t, cmd.Run(), "Failed to build webhook server")
	return webhookBin
}

func mockMCPBinary(t *testing.T) string {
	t.Helper()
	if bin := runfileBinary("server", "tests", "integration", "upstream", "testdata", "mock_mcp", "mock_mcp_", "mock_mcp"); bin != "" {
		return bin
	}
	rootDir := integration.ProjectRoot(t)
	mockBin := filepath.Join(t.TempDir(), "mock_mcp")
	cmd := exec.Command("go", "build", "-o", mockBin, "./tests/integration/upstream/testdata/mock_mcp") //nolint:gosec
	cmd.Dir = rootDir
	require.NoError(t, cmd.Run(), "Failed to build mock MCP server")
	return mockBin
}

func TestWebhooksE2E(t *testing.T) {
	webhookBin := webhookBinary(t)

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
		req, _ := http.NewRequestWithContext(context.Background(), "GET", fmt.Sprintf("http://127.0.0.1:%d/markdown", port), nil)
		resp, err := http.DefaultClient.Do(req) // Endpoint exists (POST only but connectable)
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
	webhookBin := webhookBinary(t)
	mockMcpBin := mockMCPBinary(t)

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
		req, _ := http.NewRequestWithContext(context.Background(), "GET", fmt.Sprintf("http://127.0.0.1:%d/markdown", port), nil)
		resp, err := http.DefaultClient.Do(req)
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
