// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStartup_Resilience_UpstreamFailure verifies that the server does not crash if a sub-service fails on startup.
//
// Track 1: The Bug Hunter
// Objective: Startup Reliability - Does the server crash if a sub-service fails? (It shouldn't).
//
// This test serves as a regression test to ensure that the server remains resilient to upstream failures.
// It was confirmed that the server already handles this scenario gracefully by scheduling retries.
func TestStartup_Resilience_UpstreamFailure(t *testing.T) {
	// 1. Setup
	fs := afero.NewMemMapFs()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Find free ports for the server
	l, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	httpPort := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()

	app := NewApplication()

	// Config with a service that will fail to connect.
	// We use a reserved IP in TEST-NET-1 (192.0.2.0/24) which is designated for documentation and examples
	// and should not be reachable, guaranteeing a connection timeout/refusal without race conditions.
	// 192.0.2.1 is commonly used.
	unreachableURL := "http://192.0.2.1:80/openapi.json"

	// Update config to use "openapi_service" which will try to fetch the spec and fail.
	configContent := fmt.Sprintf(`
upstream_services:
  - name: "failing-service"
    openapi_service:
      address: "http://192.0.2.1:80"
      spec_url: "%s"
`, unreachableURL)

	err = afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0o644)
	require.NoError(t, err)

	errChan := make(chan error, 1)
	go func() {
		errChan <- app.Run(ctx, fs, false, fmt.Sprintf("localhost:%d", httpPort), "localhost:0", []string{"/config.yaml"}, "", 5*time.Second)
	}()

	// 2. Verify Startup
	// The app should start successfully even if the upstream service fails to register.
	// We wait for startup signal.
	startupCtx, startupCancel := context.WithTimeout(ctx, 10*time.Second)
	defer startupCancel()

	err = app.WaitForStartup(startupCtx)
	require.NoError(t, err, "Server should have started successfully despite failing upstream")

	// 3. Verify Health
	// Check if the server is responsive
	// Pass io.Discard to avoid panic on nil writer
	err = HealthCheck(io.Discard, fmt.Sprintf("localhost:%d", httpPort), 2*time.Second)
	assert.NoError(t, err, "Server should be healthy")

	// 4. Verify Service Registry
	// The failing service should NOT be in the active tools list (or at least not crash everything)
	tools := app.ToolManager.ListTools()
	for _, tool := range tools {
		if tool.Tool().GetName() == "failing-service" {
			t.Log("Found failing service tools (unexpected if it failed to load spec)")
		}
	}

	// Wait a bit to ensure async registration had time to fail/retry
	time.Sleep(2 * time.Second)

	// Check that the app is still running (errChan is empty)
	select {
	case err := <-errChan:
		t.Fatalf("Server crashed with error: %v", err)
	default:
		// OK
	}
}
