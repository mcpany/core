package upstream_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	TIMESTAMPCB "google.golang.org/protobuf/types/known/durationpb"
)

func TestWebhooksE2E(t *testing.T) {
	// Build the webhook server
	rootDir := findRootDir(t)
	webhookBin := filepath.Join(rootDir, "build", "bin", "webhooks")
	cmd := exec.Command("go", "build", "-o", webhookBin, "./cmd/webhooks")
	cmd.Dir = rootDir
	require.NoError(t, cmd.Run(), "Failed to build webhook server")

	// Start webhook server
	serverCmd := exec.Command(webhookBin)
	serverCmd.Stdout = os.Stdout
	serverCmd.Stderr = os.Stderr
	require.NoError(t, serverCmd.Start(), "Failed to start webhook server")
	defer func() {
		_ = serverCmd.Process.Kill()
	}()

	// Wait for server to start
	require.Eventually(t, func() bool {
		resp, err := http.Get("http://localhost:8080/markdown") // Endpoint exists (POST only but connectable)
		return err == nil && (resp.StatusCode == 405 || resp.StatusCode == 200)
	}, 5*time.Second, 100*time.Millisecond, "Webhook server failed to start")

	t.Run("MarkdownConversion", func(t *testing.T) {
		url := "http://localhost:8080/markdown"
		hook := tool.NewWebhookHook(&configv1.WebhookConfig{
			Url:     &url,
			Timeout: TIMESTAMPCB.New(5 * time.Second),
		})

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
		url := "http://localhost:8080/truncate?max_chars=5"
		hook := tool.NewWebhookHook(&configv1.WebhookConfig{
			Url:     &url,
			Timeout: TIMESTAMPCB.New(5 * time.Second),
		})

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
