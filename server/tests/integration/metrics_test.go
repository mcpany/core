// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build integration
// +build integration

package integration

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetrics(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpany-integration-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configContent := `
{
  "upstream_services": [
    {
      "name": "http-echo-server",
      "auto_discover_tool": true,
      "http_service": {
        "address": "http://127.0.0.1:8080",
        "calls": {
          "echo": {
            "method": "HTTP_METHOD_POST",
            "endpoint_path": "/echo"
          }
        }
      }
    }
  ]
}
`
	configPath := filepath.Join(tempDir, "config.json")
	err = os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	// Mock HTTP Echo Server
	echoServer := http.NewServeMux()
	echoServer.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	})
	// Run on 8080 as expected by config
	srv := &http.Server{Addr: "127.0.0.1:8080", Handler: echoServer}
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			t.Logf("Echo server error: %v", err)
		}
	}()
	defer func() { _ = srv.Close() }()
	WaitForTCPPort(t, 8080, 5*time.Second)

	serverInfo := StartMCPANYServer(t, "metrics-test",
		"--config-path", configPath,
		"--metrics-listen-address", "127.0.0.1:9090",
	)
	defer serverInfo.CleanupFunc()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = serverInfo.Initialize(ctx)
	require.NoError(t, err)

	// Wait for the server to be ready/tools discovered
	require.Eventually(t, func() bool {
		tools, err := serverInfo.ListTools(ctx)
		if err != nil {
			return false
		}
		for _, tool := range tools.Tools {
			if tool.Name == "http-echo-server.echo" {
				return true
			}
		}
		return false
	}, 10*time.Second, 100*time.Millisecond, "Tool http-echo-server.echo not found")

	// Make a request to the echo tool
	args := map[string]interface{}{"message": "hello"}
	_, err = serverInfo.CallTool(ctx, &mcp.CallToolParams{
		Name:      "http-echo-server.echo",
		Arguments: args,
	})
	require.NoError(t, err)

	// Make a request to the metrics endpoint
	// Retry a few times to allow metrics to flush
	assert.Eventually(t, func() bool {
		resp, err := http.Get("http://127.0.0.1:9090/metrics")
		if err != nil {
			return false
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return false
		}
		bodyStr := string(body)
		// Check for specific metrics.
		// Use partial matching for labels as ordering might vary.
		return assert.Contains(t, bodyStr, `mcpany_tools_call_total`) &&
			assert.Contains(t, bodyStr, `tool="http-echo-server.echo"`)
	}, 5*time.Second, 500*time.Millisecond, "Metrics not found in response")
}
