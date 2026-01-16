// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// TestFix_ReloadReliability verifies that if a service fails to register during a reload,
// it is retried and eventually recovers.
//
// Track 1: The Bug Hunter
// Objective: Startup Reliability / Config Drift
// Fix Verification: Ensure retry mechanism works for reloaded services.
func TestFix_ReloadReliability(t *testing.T) {
	// 1. Setup Mock OpenAPI Server
	var shouldFail atomic.Bool
	shouldFail.Store(false)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if shouldFail.Load() {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
  "openapi": "3.0.0",
  "info": {"title": "Test Service", "version": "1.0.0"},
  "paths": {
    "/test": {
      "get": {
        "operationId": "getTest",
        "responses": {"200": {"description": "OK"}}
      }
    }
  }
}`))
	}))
	defer ts.Close()

	// 2. Setup Application
	fs := afero.NewMemMapFs()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Find free ports
	l, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	httpPort := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()

	l2, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	grpcPortStr := fmt.Sprintf("localhost:%d", l2.Addr().(*net.TCPAddr).Port)
	_ = l2.Close()

	configPath := "/config.yaml"

	// Initial Config
	writeConfig := func(priority int) {
		configContent := fmt.Sprintf(`
upstream_services:
  - name: "reload-test-service"
    priority: %d
    openapi_service:
      address: "%s"
      spec_url: "%s"
`, priority, ts.URL, ts.URL) // Use ts.URL as spec_url to hit our handler
		err = afero.WriteFile(fs, configPath, []byte(configContent), 0o644)
		require.NoError(t, err)
	}

	writeConfig(0)

	app := NewApplication()

	// Run in background
	go func() {
		_ = app.Run(ctx, fs, false, fmt.Sprintf("localhost:%d", httpPort), grpcPortStr, []string{configPath}, "", 100*time.Millisecond, 5*time.Second)
	}()

	require.NoError(t, app.WaitForStartup(ctx))

	// Verify tool exists initially
	require.Eventually(t, func() bool {
		_, ok := app.ToolManager.GetTool("reload-test-service.getTest")
		return ok
	}, 5*time.Second, 100*time.Millisecond, "Tool should be present initially")

	// 3. Trigger Failure Scenario
	// Make upstream fail
	shouldFail.Store(true)

	// Update config to trigger reload (change priority)
	writeConfig(1)

	// Force reload manually
	err = app.ReloadConfig(ctx, fs, []string{configPath})
	require.NoError(t, err)

	// Verify tool is GONE (because unregister happened, register failed)
	require.Eventually(t, func() bool {
		_, ok := app.ToolManager.GetTool("reload-test-service.getTest")
		return !ok
	}, 5*time.Second, 100*time.Millisecond, "Tool should be removed after failed reload")

	// 4. Recover Upstream
	shouldFail.Store(false)

	// 5. Verify Tool Recovery
	// The retry loop (5s default in worker) should pick it up.
	// We wait slightly longer than retry delay (5s).
	require.Eventually(t, func() bool {
		_, ok := app.ToolManager.GetTool("reload-test-service.getTest")
		return ok
	}, 15*time.Second, 500*time.Millisecond, "Tool should eventually recover via retry mechanism")
}
