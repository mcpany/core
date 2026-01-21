// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestAutoDiscoverySilentFailure(t *testing.T) {
	// Setup logger capture
	logging.ForTestsOnlyResetLogger()
	var buf ThreadSafeBuffer
	logging.Init(slog.LevelInfo, &buf)

	// Setup config with auto_discover_local enabled
	fs := afero.NewMemMapFs()
	configContent := `
global_settings:
  auto_discover_local: true
upstream_services: []
`
	err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0o644)
	require.NoError(t, err)

	// Run App
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	app := NewApplication()

	errChan := make(chan error, 1)
	go func() {
		// Run with ephemeral ports
		errChan <- app.Run(RunOptions{
			Ctx:             ctx,
			Fs:              fs,
			Stdio:           false,
			JSONRPCPort:     "127.0.0.1:0",
			GRPCPort:        "127.0.0.1:0",
			ConfigPaths:     []string{"/config.yaml"},
			APIKey:          "",
			ShutdownTimeout: 100 * time.Millisecond,
		})
	}()

	// Wait for startup to complete (which includes discovery phase)
	err = app.WaitForStartup(ctx)
	require.NoError(t, err)

	// Check logs for any warning about discovery failure
	logs := buf.String()

	// If Ollama is not running (which is true in sandbox), discovery fails.
	// We expect a warning. If no warning is found, it's a silent failure.
	hasWarning := strings.Contains(logs, "Failed to auto-discover") ||
		strings.Contains(logs, "ollama not found") ||
		strings.Contains(logs, "Auto-discovery failed")

	if !hasWarning {
		t.Logf("Full Logs:\n%s", logs)
		t.Fatal("Silent failure detected: Auto-discovery should have failed (Ollama not present) but no warning was logged.")
	}

	// Clean shutdown
	cancel()
	<-errChan
}
