// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_StartupHealthCheck_LogsErrors(t *testing.T) {
	// Capture logs
	logging.ForTestsOnlyResetLogger()
	var buf ThreadSafeBuffer
	logging.Init(slog.LevelInfo, &buf)

	fs := afero.NewMemMapFs()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Find a free port and ensure it is closed
	l, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	port := l.Addr().(*net.TCPAddr).Port
	l.Close() // Close it so connection fails

	// Create a config with a valid URL but unreachable service
	configContent := fmt.Sprintf(`
upstream_services:
  - name: "unreachable-http-service"
    http_service:
      address: "http://localhost:%d"
      tools: []
`, port)

	err = afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0o644)
	require.NoError(t, err)

	app := NewApplication()
	errChan := make(chan error, 1)

	go func() {
		// Use ephemeral ports
		errChan <- app.Run(ctx, fs, false, "localhost:0", "", []string{"/config.yaml"}, "", 5*time.Second)
	}()

	// Wait for the startup check log
	assert.Eventually(t, func() bool {
		logs := buf.String()
		// We expect a standard Startup Check Failed error (not critical, as it's network)
		// It might be "Startup Check Failed" or "Startup Check Warning" depending on implementation
		// In doctor.go: 4xx is Warning, 5xx is Error.
		// Connection failure (dial tcp ...) is returned as Error by checkURL?
		// checkURL returns StatusError for connection failure.
		return strings.Contains(logs, "Startup Check Failed") &&
               strings.Contains(logs, "unreachable-http-service")
	}, 5*time.Second, 100*time.Millisecond, "Should log startup check failure for unreachable service")

	cancel()
	<-errChan
}
