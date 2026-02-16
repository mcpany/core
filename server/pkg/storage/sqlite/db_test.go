// Copyright 2026 Author(s) of MCP Any
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
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "mcpany_sqlite_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")

	// Call NewDB
	db, err := NewDB(dbPath)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()

	// Verify file exists
	_, err = os.Stat(dbPath)
	assert.NoError(t, err, "Database file should exist")

	// Verify default Pragma configuration (WAL mode for performance)
	var journalMode string
	err = db.QueryRow("PRAGMA journal_mode;").Scan(&journalMode)
	require.NoError(t, err)
	assert.Equal(t, "wal", journalMode, "Journal mode should be WAL")

	var synchronous int
	err = db.QueryRow("PRAGMA synchronous;").Scan(&synchronous)
	require.NoError(t, err)
	assert.Equal(t, 1, synchronous, "Synchronous should be NORMAL (1)")

	var busyTimeout int
	err = db.QueryRow("PRAGMA busy_timeout;").Scan(&busyTimeout)
	require.NoError(t, err)
	assert.Equal(t, 5000, busyTimeout, "Busy timeout should be 5000")

	// Verify Tables
	tables := []string{
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

	for _, table := range tables {
		var count int
		// Check if table exists in sqlite_master
		err = db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "Table %s should exist", table)
	}
}

func TestNewDB_Idempotency(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "mcpany_sqlite_test_idempotency")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")

	// First call
	db1, err := NewDB(dbPath)
	require.NoError(t, err)
	require.NotNil(t, db1)

	// Insert some data to ensure it persists
	_, err = db1.Exec("INSERT INTO global_settings (id, config_json) VALUES (1, '{}')")
	require.NoError(t, err)

	db1.Close()

	// Second call
	db2, err := NewDB(dbPath)
	require.NoError(t, err)
	require.NotNil(t, db2)
	defer db2.Close()

	// Verify tables still exist
	var count int
	err = db2.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='upstream_services'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Verify data persists
	var configJSON string
	err = db2.QueryRow("SELECT config_json FROM global_settings WHERE id=1").Scan(&configJSON)
	require.NoError(t, err)
	assert.Equal(t, "{}", configJSON)
}

func TestNewDB_InvalidPath(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "mcpany_sqlite_test_invalid")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a file where the directory should be
	blockerPath := filepath.Join(tempDir, "blocker")
	f, err := os.Create(blockerPath)
	require.NoError(t, err)
	f.Close()

	// Try to create DB inside the file (treat file as directory)
	dbPath := filepath.Join(blockerPath, "test.db")

	db, err := NewDB(dbPath)
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "failed to create db directory")
}

func TestNewDB_InitSchemaError(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpany_sqlite_test_schema_error")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")

	// Create a read-only file to simulate failure during initSchema (which tries to write)
	// But NewDB opens with "sqlite" driver which handles file creation.
	// If we create a directory where the file should be, NewDB fails early (tested in InvalidPath).

	// Let's try to corrupt the database file so it's not a valid SQLite file?
	// Then sql.Open succeeds, but Exec fails?
	f, err := os.Create(dbPath)
	require.NoError(t, err)
	f.WriteString("NOT A SQLITE DATABASE")
	f.Close()

	// Call NewDB
	// It should fail when trying to execute PRAGMA or CREATE TABLE on a corrupted file
	db, err := NewDB(dbPath)
	assert.Error(t, err)
	assert.Nil(t, db)
	if err != nil {
		// The error message might vary depending on where it fails (pragma or schema)
		// Usually "file is not a database"
		t.Logf("Got expected error: %v", err)
	}
}
