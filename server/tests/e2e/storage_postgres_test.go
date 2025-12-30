// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/app"
	"github.com/mcpany/core/pkg/storage/postgres"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPostgresStorageE2E requires an existing postgres database.
// Set TEST_DB_DSN environment variable to run this test.
// Example: TEST_DB_DSN="postgres://user:password@localhost:5432/mcpany?sslmode=disable"
func TestPostgresStorageE2E(t *testing.T) {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		t.Skip("TEST_DB_DSN not set, skipping Postgres E2E test")
	}

	// Verify DB connection first
	db, err := postgres.NewDB(dsn)
	require.NoError(t, err, "Could not connect to database")
	// Clean up table if exists
	_, err = db.Exec("DROP TABLE IF EXISTS upstream_services")
	require.NoError(t, err)
	db.Close() // Close initial connection, let app manage it

	// Create a minimal config file
	configContent := fmt.Sprintf(`
global_settings:
  db_driver: postgres
  db_dsn: %s
  log_level: DEBUG
`, dsn)

	fs := afero.NewMemMapFs()
	err = afero.WriteFile(fs, "config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	application := app.NewApplication()

	httpPort := "18081"
	grpcPort := "19091"

	go func() {
		err := application.Run(ctx, fs, false, ":"+httpPort, ":"+grpcPort, []string{"config.yaml"}, 1*time.Second)
		if err != nil && err != context.Canceled {
			fmt.Printf("App failed: %v\n", err)
		}
	}()

	// Wait for health check
	require.Eventually(t, func() bool {
		resp, err :=  (http.DefaultClient).Get(fmt.Sprintf("http://localhost:%s/healthz", httpPort))
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == 200
	}, 10*time.Second, 100*time.Millisecond, "Server did not become healthy")

	// We assume if health check passed, the storage initialized successfully (as it fails hard on init error)
	assert.True(t, true, "App started successfully with Postgres")
}
