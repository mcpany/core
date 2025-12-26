// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServiceStartupFailure verifies that a failure in one upstream service
// does not prevent the entire server from starting.
func TestServiceStartupFailure(t *testing.T) {
	// 1. Create a mock failing service and a healthy service.
	// We'll use the http_service upstream type.
	// The "failing" service will point to a port that is closed.
	// The "healthy" service will point to a valid mock server.

	healthyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"result": "ok"}`)
	}))
	defer healthyServer.Close()

	// Find a free port and ensure it is closed for the failing service
	l, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	closedPort := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()

	fs := afero.NewMemMapFs()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 2. Configure the server with both services.
	configContent := fmt.Sprintf(`
upstream_services:
  - name: "failing-service"
    http_service:
      address: "http://localhost:%d"
      tools:
        - name: "fail-tool"
          call_id: "fail-call"
      calls:
        fail-call:
          id: "fail-call"
          endpoint_path: "/fail"
          method: "HTTP_METHOD_POST"
  - name: "healthy-service"
    http_service:
      address: "%s"
      tools:
        - name: "healthy-tool"
          call_id: "healthy-call"
      calls:
        healthy-call:
          id: "healthy-call"
          endpoint_path: "/healthy"
          method: "HTTP_METHOD_POST"
`, closedPort, healthyServer.URL)

	err = afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0o644)
	require.NoError(t, err)

	app := NewApplication()
	errChan := make(chan error, 1)

	// Use ephemeral ports for the main server
	go func() {
		errChan <- app.Run(ctx, fs, false, "localhost:0", "localhost:0", []string{"/config.yaml"}, 5*time.Second)
	}()

	// 3. Verify that the server starts successfully despite the failing service.
	// We check if the healthy tool is registered.
	// We give it a few seconds to stabilize.

	// Wait for up to 5 seconds for the healthy tool to appear.
	success := false
	for i := 0; i < 50; i++ {
		tools := app.ToolManager.ListTools()
		for _, t := range tools {
			if t.Tool().GetName() == "healthy-tool" {
				success = true
				break
			}
		}
		if success {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// 4. If the server crashed, app.Run would return an error to errChan.
	select {
	case err := <-errChan:
		// If we get an error here, it means the server crashed.
		assert.Fail(t, "Server crashed on startup due to service failure: "+err.Error())
	default:
		// No crash yet.
	}

	assert.True(t, success, "Healthy service should be registered even if another service fails")

	// Clean up
	cancel()
	err = <-errChan
	assert.NoError(t, err, "app.Run should return nil on graceful shutdown")
}

func TestConfigDrift(t *testing.T) {
	// 1. Start a mock upstream service
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"result": "ok"}`)
	}))
	defer upstream.Close()

	fs := afero.NewMemMapFs()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 2. Initial Configuration
	configContent := fmt.Sprintf(`
upstream_services:
  - name: "test-service"
    http_service:
      address: "%s"
      tools:
        - name: "tool-v1"
          call_id: "call-v1"
      calls:
        call-v1:
          id: "call-v1"
          endpoint_path: "/v1"
          method: "HTTP_METHOD_POST"
`, upstream.URL)

	err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0o644)
	require.NoError(t, err)

	app := NewApplication()
	errChan := make(chan error, 1)

	go func() {
		errChan <- app.Run(ctx, fs, false, "localhost:0", "localhost:0", []string{"/config.yaml"}, 5*time.Second)
	}()

	// Wait for startup
	time.Sleep(1 * time.Second)

	// Verify v1 tool exists
	_, ok := app.ToolManager.GetTool("test-service.tool-v1")
	assert.True(t, ok, "tool-v1 should exist initially")

	// 3. Update Configuration (Simulate Drift)
	// We change "tool-v1" to "tool-v2" and remove "tool-v1"
	newConfigContent := fmt.Sprintf(`
upstream_services:
  - name: "test-service"
    http_service:
      address: "%s"
      tools:
        - name: "tool-v2"
          call_id: "call-v2"
      calls:
        call-v2:
          id: "call-v2"
          endpoint_path: "/v2"
          method: "HTTP_METHOD_POST"
`, upstream.URL)

	err = afero.WriteFile(fs, "/config.yaml", []byte(newConfigContent), 0o644)
	require.NoError(t, err)

	// Trigger Reload
	err = app.ReloadConfig(fs, []string{"/config.yaml"})
	require.NoError(t, err)

	// 4. Verify Drift
	// tool-v1 should be GONE
	// tool-v2 should EXIST

	_, ok = app.ToolManager.GetTool("test-service.tool-v1")
	assert.False(t, ok, "tool-v1 should be removed after reload")

	_, ok = app.ToolManager.GetTool("test-service.tool-v2")
	assert.True(t, ok, "tool-v2 should exist after reload")
}

// TestResourceLeak verifies that reloading configuration does not leak resources
// (specifically, that it closes old connection pools).
func TestResourceLeak(t *testing.T) {
	// 1. Start a mock upstream service
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"result": "ok"}`)
	}))
	defer upstream.Close()

	fs := afero.NewMemMapFs()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initial Config
	configContent := fmt.Sprintf(`
upstream_services:
  - name: "leak-test-service"
    http_service:
      address: "%s"
      tools:
        - name: "tool-v1"
          call_id: "call-v1"
      calls:
        call-v1:
          id: "call-v1"
          endpoint_path: "/v1"
          method: "HTTP_METHOD_POST"
`, upstream.URL)

	err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0o644)
	require.NoError(t, err)

	app := NewApplication()
	errChan := make(chan error, 1)

	go func() {
		errChan <- app.Run(ctx, fs, false, "localhost:0", "localhost:0", []string{"/config.yaml"}, 5*time.Second)
	}()

	// Wait for startup
	time.Sleep(1 * time.Second)

	// Verify initial Goroutines count
	initialGoroutines := runtime.NumGoroutine()

	// 2. Perform many reloads creating NEW services
	for i := 0; i < 50; i++ {
		serviceName := fmt.Sprintf("service-%d", i)
		cfg := fmt.Sprintf(`
upstream_services:
  - name: "%s"
    connection_pool:
      max_connections: 5
      max_idle_connections: 1
    http_service:
      address: "%s"
      tools:
        - name: "tool"
          call_id: "call"
      calls:
        call:
          id: "call"
          endpoint_path: "/"
          method: "HTTP_METHOD_GET"
`, serviceName, upstream.URL)

		err = afero.WriteFile(fs, "/config.yaml", []byte(cfg), 0o644)
		require.NoError(t, err)

		err = app.ReloadConfig(fs, []string{"/config.yaml"})
		require.NoError(t, err)

		// Small sleep to allow things to settle
		time.Sleep(10 * time.Millisecond)
	}

	// Now remove the last service
	err = afero.WriteFile(fs, "/config.yaml", []byte("upstream_services: []"), 0o644)
	require.NoError(t, err)
	err = app.ReloadConfig(fs, []string{"/config.yaml"})
	require.NoError(t, err)

	// Force GC
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	finalGoroutines := runtime.NumGoroutine()

	// Let's log it.
	t.Logf("Initial Goroutines: %d, Final: %d", initialGoroutines, finalGoroutines)

	// If we see significant increase, it's a leak.
	if finalGoroutines > initialGoroutines + 20 {
		t.Errorf("Potential goroutine leak: started with %d, ended with %d", initialGoroutines, finalGoroutines)
	}
}
