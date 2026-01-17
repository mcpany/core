// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStartup_Fails_On_UpstreamFailure verifies that the server fails to start if a sub-service fails on startup.
//
// Track 1: The Friction Fighter
// Objective: Startup Reliability - Does the server fail loudly if a sub-service fails? (It should).
//
// This test ensures that the server validates connectivity for all services at startup, preventing "silent failures".
func TestStartup_Fails_On_UpstreamFailure(t *testing.T) {
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

	// 2. Run and Expect Error
	// The app should return an error because the upstream service fails to register.
	err = app.Run(ctx, fs, false, fmt.Sprintf("localhost:%d", httpPort), "localhost:0", []string{"/config.yaml"}, "", 5*time.Second)

	require.Error(t, err, "Server should have failed to start due to failing upstream")
	assert.Contains(t, err.Error(), "failed to register service", "Error should mention registration failure")
}
