package app

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAutoDiscoveryStatus verifies that auto-discovery failure is recorded in the DiscoveryManager
// and accessible via internal state (and thus via API if we were testing the full API stack).
func TestAutoDiscoveryStatus(t *testing.T) {
	// Setup logger capture to ensure we don't spam stdout, though we are not relying on logs primarily now.
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

	// Check Discovery Status
	require.NotNil(t, app.DiscoveryManager, "DiscoveryManager should be initialized")

	status, ok := app.DiscoveryManager.GetProviderStatus("ollama")
	require.True(t, ok, "Ollama provider should be registered")

	assert.Equal(t, "ERROR", status.Status, "Status should be ERROR because Ollama is not running")
	assert.Contains(t, status.LastError, "ollama not found", "Error message should contain 'ollama not found'")
	assert.Contains(t, status.LastError, "connection refused", "Error message should contain 'connection refused'")

	// Clean shutdown
	cancel()
	select {
	case <-errChan:
	case <-time.After(1 * time.Second):
		t.Log("Timed out waiting for shutdown")
	}
}
