// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package sqlite

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewDB_Success verifies database file creation and PRAGMA setups.
func TestNewDB_Success(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := NewDB(dbPath)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer func() { _ = db.Close() }()

	// Verify database file exists
	_, err = os.Stat(dbPath)
	assert.NoError(t, err)

	// Verify PRAGMAs
	var journalMode string
	err = db.QueryRowContext(context.Background(), "PRAGMA journal_mode;").Scan(&journalMode)
	require.NoError(t, err)
	assert.Equal(t, "wal", journalMode)

	var synchronous int
	err = db.QueryRowContext(context.Background(), "PRAGMA synchronous;").Scan(&synchronous)
	require.NoError(t, err)
	// NORMAL maps to 1 in SQLite
	assert.Equal(t, 1, synchronous)

	var busyTimeout int
	err = db.QueryRowContext(context.Background(), "PRAGMA busy_timeout;").Scan(&busyTimeout)
	require.NoError(t, err)
	assert.Equal(t, 5000, busyTimeout)

	// MaxOpenConns cannot be directly verified through pure sql queries,
	// but we can ensure stats reflect what db sets
	stats := db.Stats()
	assert.Equal(t, 1, stats.MaxOpenConnections) // currently no active
}

// TestNewDB_Failure_Mkdir verifies failure behavior when directory cannot be created.
func TestNewDB_Failure_Mkdir(t *testing.T) {
	tempDir := t.TempDir()
	// Create a file where we want to create a directory
	badPath := filepath.Join(tempDir, "file_as_dir")
	err := os.WriteFile(badPath, []byte("content"), 0644)
	require.NoError(t, err)

	// Attempt to create DB inside the file
	dbPath := filepath.Join(badPath, "test.db")
	db, err := NewDB(dbPath)
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "failed to create db directory")
}

// TestInitSchema_TablesExist verifies that the initSchema function successfully creates all necessary tables.
func TestInitSchema_TablesExist(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "schema_test.db")

	db, err := NewDB(dbPath)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer func() { _ = db.Close() }()

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
		"logs",
	}

	for _, tableName := range expectedTables {
		t.Run("Table_"+tableName, func(t *testing.T) {
			var name string
			query := "SELECT name FROM sqlite_master WHERE type='table' AND name=?;"
			err := db.QueryRowContext(context.Background(), query, tableName).Scan(&name)
			require.NoError(t, err, "Table %s should exist", tableName)
			assert.Equal(t, tableName, name)
		})
	}

	t.Run("Index_idx_logs_timestamp", func(t *testing.T) {
		var name string
		query := "SELECT name FROM sqlite_master WHERE type='index' AND name=?;"
		err := db.QueryRowContext(context.Background(), query, "idx_logs_timestamp").Scan(&name)
		require.NoError(t, err, "Index idx_logs_timestamp should exist")
		assert.Equal(t, "idx_logs_timestamp", name)
	})
}
