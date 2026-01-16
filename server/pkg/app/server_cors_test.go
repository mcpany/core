// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer_CORS(t *testing.T) {
	// This test sets up the full application server stack and verifies CORS headers.

	fs := afero.NewMemMapFs()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	l, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	port := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()
	addr := fmt.Sprintf("localhost:%d", port)

	// Config with debug mode enabled to trigger permissive CORS
	configContent := `
global_settings:
  log_level: DEBUG
upstream_services: []
`
	err = afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0o644)
	require.NoError(t, err)

	app := NewApplication()
	errChan := make(chan error, 1)
	go func() {
		// Use ephemeral port for HTTP
		errChan <- app.Run(ctx, fs, false, addr, "", []string{"/config.yaml"}, "", 100*time.Millisecond, 5*time.Second)
	}()

	waitForServerReady(t, addr)

	// 1. Verify CORS on OPTIONS request with Allowed Origin (via *)
	origin := "http://example.com"
	req, err := http.NewRequest("OPTIONS", "http://"+addr+"/upload", nil)
	require.NoError(t, err)
	req.Header.Set("Origin", origin)
	req.Header.Set("Access-Control-Request-Method", "POST")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	// Should allow, BUT return "*" and NO credentials because it's a wildcard match
	assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"), "CORS header should be * when * matches")
	assert.Empty(t, resp.Header.Get("Access-Control-Allow-Credentials"), "Should not allow credentials with wildcard")

	// 2. Verify POST
	req, err = http.NewRequest("POST", "http://"+addr+"/upload", nil)
	require.NoError(t, err)
	req.Header.Set("Origin", origin)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))

	cancel()
	err = <-errChan
	assert.NoError(t, err)
}

func TestServer_CORS_Strict(t *testing.T) {
	// Test strict mode (debug disabled)

	fs := afero.NewMemMapFs()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	l, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	port := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()
	addr := fmt.Sprintf("localhost:%d", port)

	// Config with default log level (INFO) -> IsDebug() false
	configContent := `
global_settings:
  log_level: INFO
upstream_services: []
`
	err = afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0o644)
	require.NoError(t, err)

	app := NewApplication()
	errChan := make(chan error, 1)
	go func() {
		errChan <- app.Run(ctx, fs, false, addr, "", []string{"/config.yaml"}, "", 100*time.Millisecond, 5*time.Second)
	}()

	waitForServerReady(t, addr)

	origin := "http://example.com"
	req, err := http.NewRequest("OPTIONS", "http://"+addr+"/upload", nil)
	require.NoError(t, err)
	req.Header.Set("Origin", origin)
	req.Header.Set("Access-Control-Request-Method", "POST")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	// Should NOT have CORS headers
	assert.Empty(t, resp.Header.Get("Access-Control-Allow-Origin"), "CORS header should be missing in strict mode")

	cancel()
	err = <-errChan
	assert.NoError(t, err)
}
