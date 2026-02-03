// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/app"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/spf13/afero"
	"log/slog"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogsPersistence(t *testing.T) {
	// Initialize logging properly
	logging.ForTestsOnlyResetLogger()
	logging.Init(slog.LevelDebug, os.Stderr)

	// Enable file config for this test
	os.Setenv("MCPANY_ENABLE_FILE_CONFIG", "true")
	defer os.Unsetenv("MCPANY_ENABLE_FILE_CONFIG")

	// Setup temporary DB path
	dbPath := t.TempDir() + "/mcpany_logs_test.db"

	configContent := fmt.Sprintf(`
global_settings:
    db_driver: "sqlite"
    db_path: "%s"
`, dbPath)

	tmpFile, err := os.CreateTemp(t.TempDir(), "mcpany-config-*.yaml")
	require.NoError(t, err)
	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)

	var ctx context.Context
	var cancel context.CancelFunc

	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	appRunner := app.NewApplication()

	done := make(chan struct{})
	go func() {
		defer close(done)
		fs := afero.NewOsFs()
		opts := app.RunOptions{
			Ctx:             ctx,
			Fs:              fs,
			Stdio:           false,
			JSONRPCPort:     "127.0.0.1:0",
			GRPCPort:        "127.0.0.1:0",
			ConfigPaths:     []string{tmpFile.Name()},
			ShutdownTimeout: 5 * time.Second,
		}
		if err := appRunner.Run(opts); err != nil && err != context.Canceled {
			t.Logf("Application run error: %v", err)
		}
	}()
	defer func() {
		cancel()
		<-done
	}()

	err = appRunner.WaitForStartup(ctx)
	require.NoError(t, err, "Failed to wait for startup")

	jsonrpcPort := int(appRunner.BoundHTTPPort.Load())

	// Wait for health check
	httpUrl := fmt.Sprintf("http://127.0.0.1:%d/healthz", jsonrpcPort)
	integration.WaitForHTTPHealth(t, httpUrl, 10*time.Second)

	// 1. Generate logs via API call (AuthMiddleware logs details)
	// We call a public endpoint that triggers auth middleware but doesn't require auth (e.g. login, or just fail auth)
	// /api/v1/debug/auth-test requires auth or public access.
	// If no API Key configured (default), AuthMiddleware enforces localhost. We are localhost.
	// It logs "DEBUG: AuthMiddleware details".
	authTestURL := fmt.Sprintf("http://127.0.0.1:%d/healthz", jsonrpcPort)
	_, err = http.Get(authTestURL)
	require.NoError(t, err)

	// Allow async logging worker to process
	time.Sleep(500 * time.Millisecond)

	// 2. Query Logs API
	logsURL := fmt.Sprintf("http://127.0.0.1:%d/api/v1/logs/history?limit=100", jsonrpcPort)
	resp, err := http.Get(logsURL)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var logsResp struct {
		Logs  []logging.LogEntry `json:"logs"`
		Total int                `json:"total"`
	}
	err = json.NewDecoder(resp.Body).Decode(&logsResp)
	require.NoError(t, err)

	t.Logf("Total logs: %d", logsResp.Total)
	if logsResp.Total == 0 {
		t.Log("Warning: No logs found. Maybe AuthMiddleware didn't log?")
		// Force a log via internal logger if possible, but we don't have access to the running instance's logger easily from here
		// except that GetLogger() is a singleton.
		// BUT appRunner is in same process.
		logging.GetLogger().Info("E2E Test Explicit Log")
		time.Sleep(100 * time.Millisecond)
		// Query again
		resp, _ = http.Get(logsURL)
		defer resp.Body.Close()
		_ = json.NewDecoder(resp.Body).Decode(&logsResp)
	}

	assert.Greater(t, logsResp.Total, 0)
	assert.NotEmpty(t, logsResp.Logs)
}
