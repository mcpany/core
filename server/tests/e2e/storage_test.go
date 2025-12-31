// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/app"
	"github.com/mcpany/core/pkg/config"
	"github.com/mcpany/core/pkg/storage/postgres"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPostgresStorageE2E verifies that the server can start with Postgres configuration.
// Note: This test requires a running Postgres instance. It skips if POSTGRES_DSN is not set.
func TestPostgresStorageE2E(t *testing.T) {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		t.Skip("POSTGRES_DSN not set, skipping Postgres E2E test")
	}

	// Create a unique table prefix or ensure clean state?
	// For this E2E, we'll rely on the app logic to init schema.
	// Best effort cleanup.
	db, err := postgres.NewDB(dsn)
	require.NoError(t, err)
	_, _ = db.Exec("DROP TABLE IF EXISTS upstream_services") // CLEANUP
	_ = db.Close()

	fs := afero.NewMemMapFs()
	configContent := fmt.Sprintf(`
global_settings:
  db_driver: postgres
  db_dsn: "%s"
  log_level: LOG_LEVEL_DEBUG
`, dsn)

	err = afero.WriteFile(fs, "config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	application := app.NewApplication()

	// Run the server in a goroutine
	go func() {
		// We expect it to fail gracefully on context cancel or succeed startup
		_ = application.Run(
			ctx,
			fs,
			false,
			":0", // Random ports
			"",
			[]string{"config.yaml"},
			1*time.Second,
		)
	}()

	// Wait a bit for startup
	time.Sleep(2 * time.Second)

	// Verify DB was initialized
	db, err = postgres.NewDB(dsn)
	require.NoError(t, err)
	defer db.Close()

	var tableName string
	err = db.QueryRow("SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND tablename = 'upstream_services'").Scan(&tableName)
	assert.NoError(t, err, "upstream_services table should exist")
	assert.Equal(t, "upstream_services", tableName)
}
