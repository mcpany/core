// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/app"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestServiceRetry(t *testing.T) {
	// 1. Create a config file pointing to a non-existent MCP server
	fs := afero.NewMemMapFs()

    // Get an ephemeral port by listening on port 0
    l, err := net.Listen("tcp", "127.0.0.1:0")
    require.NoError(t, err)
    // The actual address is deferred to the httptest.Server
    targetURL := fmt.Sprintf("http://%s/mcp", l.Addr().String())

	configContent := fmt.Sprintf(`
upstream_services:
  - name: "delayed-mcp"
    mcp_service:
      http_connection:
        http_address: "%s"
`, targetURL)

	err = afero.WriteFile(fs, "config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)

	// 2. Start the Application
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a := app.NewApplication()
	go func() {
		err := a.Run(ctx, fs, false, ":8085", "", []string{"config.yaml"}, 1*time.Second)
        if err != nil && ctx.Err() == nil {
            t.Logf("Application run error: %v", err)
        }
	}()

	// Wait for app to start
	err = a.WaitForStartup(ctx)
	if err != nil {
		t.Fatalf("Failed to wait for startup: %v", err)
	}

	// Verify service failed to register
    require.Eventually(t, func() bool {
        return a.ServiceRegistry != nil
    }, 5*time.Second, 100*time.Millisecond, "ServiceRegistry not initialized")

    errStr, hasError := a.ServiceRegistry.GetServiceError("delayed-mcp")
    t.Logf("Initial Service Error: %s (hasError: %v)", errStr, hasError)

    if !hasError {
         _, infoOk := a.ServiceRegistry.GetServiceInfo("delayed-mcp")
         if infoOk {
             t.Log("Service registered successfully unexpectedly!")
         }
    } else {
        t.Log("Service correctly failed to register initially.")
    }

    // 3. Start the mock MCP service
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Minimal MCP response for initialization
        w.Header().Set("Content-Type", "application/json")

        // If it's a POST (JSON-RPC), return result.
        // MCP HTTP transport spec is SSE + POST.
        // But the SDK might fallback or check something.
        // The error was "unsupported content type", so fixing that might be enough to pass "connection" check.
        // But likely it sends `initialize` request and expects a response.
        // JSON-RPC response:
        w.Write([]byte(`{"jsonrpc": "2.0", "id": 1, "result": {"protocolVersion": "2024-11-05", "capabilities": {}, "serverInfo": {"name": "mock", "version": "1.0"}}}`))
	}))

    // Assign the listener with the dynamic port
    ts.Listener = l
	ts.Start()
	defer ts.Close()

    t.Logf("Started mock service at %s", targetURL)

	// 4. Wait and see if it recovers
    t.Log("Waiting for retry...")
    time.Sleep(10 * time.Second)

    // Check if service is now healthy
    errStr, hasError = a.ServiceRegistry.GetServiceError("delayed-mcp")
    if hasError {
        t.Logf("Service still has error: %s (failed to recover)", errStr)
        t.Fail() // IT FAILED TO RECOVER
    } else {
        t.Log("Service recovered successfully (error cleared)!")
    }
}
