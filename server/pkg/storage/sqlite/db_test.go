// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package sqlite

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDB_Success(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mcpany-db-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "mcp.db")

	// 1. Create DB
	db, err := NewDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	// 2. Verify File Exists
	_, err = os.Stat(dbPath)
	assert.NoError(t, err)

	// 3. Verify Tables
	expectedTables := []string{
		"upstream_services",
		"global_settings",
		"secrets",
		"users",
		"profile_definitions",
		"service_collections",
		"user_tokens",
		"credentials",
		"service_templates",
	}

	for _, table := range expectedTables {
		var name string
		// Check if table exists in sqlite_master
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		assert.NoError(t, err, "Table %s should exist", table)
		assert.Equal(t, table, name)
	}

	// 4. Verify PRAGMAs
	var journalMode string
	err = db.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	assert.NoError(t, err)
	assert.Equal(t, "wal", journalMode)

	var synchronous int
	err = db.QueryRow("PRAGMA synchronous").Scan(&synchronous)
	assert.NoError(t, err)
	// NORMAL is 1
	assert.Equal(t, 1, synchronous)

	var busyTimeout int
	err = db.QueryRow("PRAGMA busy_timeout").Scan(&busyTimeout)
	assert.NoError(t, err)
	assert.Equal(t, 5000, busyTimeout)
}

func TestNewDB_Idempotency(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mcpany-db-test-idem-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "mcp.db")

	// First initialization
	db1, err := NewDB(dbPath)
	require.NoError(t, err)
	db1.Close()

	// Second initialization (should not fail)
	db2, err := NewDB(dbPath)
	require.NoError(t, err)
	defer db2.Close()

	// Verify tables still exist
	var count int
	err = db2.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='upstream_services'").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestNewDB_Failure_InvalidPath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mcpany-db-test-fail-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a directory where the file should be
	dbPath := filepath.Join(tmpDir, "is_a_directory")
	err = os.Mkdir(dbPath, 0755)
	require.NoError(t, err)

	// Attempt to open DB on a path that is a directory
	// sql.Open might succeed, but initSchema or PRAGMAs might fail,
	// or the driver might return an error immediately upon connection.
	// Actually, sql.Open usually lazily connects.
	// But NewDB executes initSchema which runs a query, forcing connection.
	// Opening a directory as a DB usually fails in SQLite driver.

	db, err := NewDB(dbPath)
	if err == nil {
		db.Close()
	}
	assert.Error(t, err)
}

func TestNewDB_Failure_Mkdir(t *testing.T) {
	// This test tries to create a DB in a path where parent directory creation fails.
	// E.g. /dev/null/test.db or similar system protected paths.
	// But in a sandbox, we might not have root.
	// Let's use a file as the parent directory.

	tmpDir, err := os.MkdirTemp("", "mcpany-db-test-mkdir-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	parentFile := filepath.Join(tmpDir, "parent_is_file")
	err = os.WriteFile(parentFile, []byte("data"), 0644)
	require.NoError(t, err)

	// NewDB calls os.MkdirAll(filepath.Dir(path), 0750)
	// We want filepath.Dir(dbPath) to be parentFile.
	// So dbPath should be parentFile/mcp.db

	dbPath := filepath.Join(parentFile, "mcp.db")

	db, err := NewDB(dbPath)
	if err == nil {
		db.Close()
	}
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create db directory")
}
