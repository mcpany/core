// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/app"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestServiceRetry(t *testing.T) {
	// 1. Create a dynamic mock server that we can toggle
	var shouldFail atomic.Bool
	shouldFail.Store(true) // Start in failing mode

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if shouldFail.Load() {
			// Simulate failure (503 Service Unavailable)
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			return
		}

		// Success mode: Return valid MCP initialization response
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"jsonrpc": "2.0", "id": 1, "result": {"protocolVersion": "2024-11-05", "capabilities": {}, "serverInfo": {"name": "mock", "version": "1.0"}}}`))
	}))
	defer ts.Close()

	// 2. Configure App to point to this server
	fs := afero.NewMemMapFs()

	configContent := fmt.Sprintf(`
upstream_services:
  - name: "delayed-mcp"
    mcp_service:
      http_connection:
        http_address: "%s"
    resilience:
      timeout: "0.5s"
`, ts.URL)

	err := afero.WriteFile(fs, "config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)

	// 3. Start the Application
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a := app.NewApplication()
	go func() {
		// Use a random port for the app itself to avoid conflicts
		err := a.Run(ctx, fs, false, "127.0.0.1:0", "", []string{"config.yaml"}, "", 1*time.Second)
		if err != nil && ctx.Err() == nil {
			t.Logf("Application run error: %v", err)
		}
	}()

	// Wait for app to start
	err = a.WaitForStartup(ctx)
	if err != nil {
		t.Fatalf("Failed to wait for startup: %v", err)
	}

	// 4. Verify service failed to register initially
	require.Eventually(t, func() bool {
		if a.ServiceRegistry == nil {
			return false
		}
		_, hasError := a.ServiceRegistry.GetServiceError("delayed-mcp")
		return hasError
	}, 15*time.Second, 100*time.Millisecond, "ServiceRegistry not initialized or service did not fail as expected")

	errStr, hasError := a.ServiceRegistry.GetServiceError("delayed-mcp")
	t.Logf("Initial Service Error: %s (hasError: %v)", errStr, hasError)
	require.True(t, hasError, "Service should have an error initially")

	// 5. "Fix" the server
	t.Log("Enabling mock service...")
	shouldFail.Store(false)

	// 6. Wait and see if it recovers
	// The worker retries every ~5s (default) or configuration specific.
	// We wait enough time for a retry cycle.
	t.Log("Waiting for retry...")
	require.Eventually(t, func() bool {
		_, hasError := a.ServiceRegistry.GetServiceError("delayed-mcp")
		return !hasError
	}, 15*time.Second, 500*time.Millisecond, "Service failed to recover after server was fixed")

	t.Log("Service recovered successfully!")

}
