// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestNewSQLiteAuditStore_EdgeCases(t *testing.T) {
	t.Run("empty path", func(t *testing.T) {
		store, err := NewSQLiteAuditStore("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "sqlite path is required")
		assert.Nil(t, store)
	})

	t.Run("invalid path", func(t *testing.T) {
		// Use a directory as the path which should fail for SQLite
		tempDir, err := os.MkdirTemp("", "audit_invalid")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		store, err := NewSQLiteAuditStore(tempDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create audit_logs table")
		assert.Nil(t, store)
	})

	t.Run("verify empty db", func(t *testing.T) {
		f, err := os.CreateTemp("", "audit_empty_*.db")
		require.NoError(t, err)
		dbPath := f.Name()
		f.Close()
		defer os.Remove(dbPath)

		store, err := NewSQLiteAuditStore(dbPath)
		require.NoError(t, err)
		defer store.Close()

		valid, err := store.Verify()
		assert.NoError(t, err)
		assert.True(t, valid)
	})
}

func TestEnsureColumns_Failure(t *testing.T) {
	// This test is tricky because we need to simulate a failure in ensureColumns.
	// Since NewSQLiteAuditStore calls ensureColumns internally, we can try to
	// create a table with a schema that conflicts or create a read-only db?
	// Or we can mock the DB, but ensureColumns takes *sql.DB which is hard to mock without a driver.

	// Let's try to make the DB read-only after creation but before ensureColumns runs?
	// Hard to do atomically.

	// Alternative: Create a table with a conflicting column type?
	f, err := os.CreateTemp("", "audit_col_fail_*.db")
	require.NoError(t, err)
	dbPath := f.Name()
	f.Close()
	defer os.Remove(dbPath)

	// Manually create table with a conflicting column type or name?
	// SQLite is very flexible with types, so type conflict might not work.
	// But if we create a table that is NOT audit_logs but then... no.

	// If we make the file read-only, NewSQLiteAuditStore might fail at CREATE TABLE or ensureColumns.
	// Let's create the table first, then make file read-only.

	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)
	_, err = db.Exec("CREATE TABLE audit_logs (id INTEGER PRIMARY KEY)")
	require.NoError(t, err)
	db.Close()

	// Make file read-only
	err = os.Chmod(dbPath, 0400)
	require.NoError(t, err)

	// NewSQLiteAuditStore should fail
	store, err := NewSQLiteAuditStore(dbPath)
	assert.Error(t, err)
	// The error could be at ensureColumns or earlier.
	// If it fails at ensureColumns (ALTER TABLE), it satisfies our coverage goal indirectly.
	assert.Nil(t, store)
}

func TestSQLiteAuditStore_Write_Errors(t *testing.T) {
	f, err := os.CreateTemp("", "audit_write_fail_*.db")
	require.NoError(t, err)
	dbPath := f.Name()
	f.Close()
	defer os.Remove(dbPath)

	store, err := NewSQLiteAuditStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	// Close the DB underneath
	store.db.Close()

	entry := AuditEntry{
		Timestamp: time.Now(),
		ToolName: "test",
	}

	err = store.Write(entry)
	assert.Error(t, err)
}

func TestSQLiteAuditStore_Verify_Errors(t *testing.T) {
	f, err := os.CreateTemp("", "audit_verify_fail_*.db")
	require.NoError(t, err)
	dbPath := f.Name()
	f.Close()
	defer os.Remove(dbPath)

	store, err := NewSQLiteAuditStore(dbPath)
	require.NoError(t, err)

	// Inject a row with missing columns or invalid data structure?
	// SQLite schema is flexible.
	// Let's try closing the DB to trigger query error.
	store.db.Close()

	valid, err := store.Verify()
	assert.Error(t, err)
	assert.False(t, valid)
}

func TestSQLiteAuditStore_ComplexWrite(t *testing.T) {
    // Tests writing with complex arguments and results (nil, non-nil, etc.)
    f, err := os.CreateTemp("", "audit_complex_*.db")
    require.NoError(t, err)
    dbPath := f.Name()
    f.Close()
    defer os.Remove(dbPath)

    store, err := NewSQLiteAuditStore(dbPath)
    require.NoError(t, err)
    defer store.Close()

    entry := AuditEntry{
        Timestamp:  time.Now(),
        ToolName:   "complex_tool",
        Arguments:  nil, // Should become "{}"
        Result:     nil, // Should become "{}"
        DurationMs: 50,
    }

    err = store.Write(entry)
    assert.NoError(t, err)

    // Verify it was stored as "{}"
    db, err := sql.Open("sqlite", dbPath)
    require.NoError(t, err)
    defer db.Close()

    var args, res string
    err = db.QueryRow("SELECT arguments, result FROM audit_logs").Scan(&args, &res)
    require.NoError(t, err)
    assert.Equal(t, "{}", args)
    assert.Equal(t, "{}", res)
}

func TestSQLiteAuditStore_IntegrityViolation_PrevHash(t *testing.T) {
     f, err := os.CreateTemp("", "audit_integrity_*.db")
    require.NoError(t, err)
    dbPath := f.Name()
    f.Close()
    defer os.Remove(dbPath)

    store, err := NewSQLiteAuditStore(dbPath)
    require.NoError(t, err)

    // Write 2 entries
    require.NoError(t, store.Write(AuditEntry{Timestamp: time.Now(), ToolName: "1"}))
    require.NoError(t, store.Write(AuditEntry{Timestamp: time.Now(), ToolName: "2"}))
    store.Close()

    // Manually tamper prev_hash of second entry
    db, err := sql.Open("sqlite", dbPath)
    require.NoError(t, err)
    _, err = db.Exec("UPDATE audit_logs SET prev_hash = 'tampered' WHERE id = 2")
    require.NoError(t, err)
    db.Close()

    store, err = NewSQLiteAuditStore(dbPath)
    require.NoError(t, err)
    defer store.Close()

    valid, err := store.Verify()
    assert.False(t, valid)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "prev_hash mismatch")
}

func TestEnsureColumns_AlreadyExists(t *testing.T) {
    // Test that ensureColumns doesn't fail if columns already exist
    f, err := os.CreateTemp("", "audit_exists_*.db")
    require.NoError(t, err)
    dbPath := f.Name()
    f.Close()
    defer os.Remove(dbPath)

    // Create DB and columns manually
    db, err := sql.Open("sqlite", dbPath)
    require.NoError(t, err)
    _, err = db.Exec("CREATE TABLE audit_logs (id INTEGER PRIMARY KEY, hash TEXT, prev_hash TEXT)")
    require.NoError(t, err)
    db.Close()

    // Should succeed
    store, err := NewSQLiteAuditStore(dbPath)
    assert.NoError(t, err)
    if store != nil {
        store.Close()
    }
}

func TestSQLiteAuditStore_ConcurrentWrites(t *testing.T) {
    f, err := os.CreateTemp("", "audit_concurrent_*.db")
    require.NoError(t, err)
    dbPath := f.Name()
    f.Close()
    defer os.Remove(dbPath)

    store, err := NewSQLiteAuditStore(dbPath)
    require.NoError(t, err)
    defer store.Close()

    concurrency := 10
    done := make(chan bool)

    for i := 0; i < concurrency; i++ {
        go func(idx int) {
            err := store.Write(AuditEntry{
                Timestamp: time.Now(),
                ToolName:  fmt.Sprintf("tool_%d", idx),
            })
            assert.NoError(t, err)
            done <- true
        }(i)
    }

    for i := 0; i < concurrency; i++ {
        <-done
    }

    valid, err := store.Verify()
    assert.NoError(t, err)
    assert.True(t, valid)
}
