package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestEnsureColumn_Validation(t *testing.T) {
	// Create a temporary database file
	f, err := os.CreateTemp("", "audit_validation_*.db")
	require.NoError(t, err)
	dbPath := f.Name()
	f.Close()
	defer os.Remove(dbPath)

	// Allow the temp path
	validation.SetAllowedPaths([]string{os.TempDir()})
	defer validation.SetAllowedPaths(nil)

	// Open DB
	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)
	defer db.Close()

	// Initialize table
	_, err = db.Exec("CREATE TABLE audit_logs (id INTEGER PRIMARY KEY)")
	require.NoError(t, err)

	// Test valid column
	err = ensureColumn(db, "hash")
	assert.NoError(t, err)

	// Test invalid column
	err = ensureColumn(db, "invalid_col")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid column name")

	// Test injection attempt
	err = ensureColumn(db, "hash; DROP TABLE audit_logs;")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid column name")
}

func TestSQLiteAuditStore(t *testing.T) {
	// Create a temporary database file
	f, err := os.CreateTemp("", "audit_test_*.db")
	require.NoError(t, err)
	dbPath := f.Name()
	f.Close()
	defer os.Remove(dbPath)

	// Allow the temp path
	validation.SetAllowedPaths([]string{os.TempDir()})
	defer validation.SetAllowedPaths(nil)

	// Initialize store
	store, err := NewSQLiteAuditStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	// Create a sample entry
	entry := Entry{
		Timestamp:  time.Date(2023, 10, 27, 12, 0, 0, 0, time.UTC),
		ToolName:   "test_tool",
		UserID:     "user123",
		ProfileID:  "profileABC",
		Arguments:  json.RawMessage(`{"arg": "value"}`),
		Result:     map[string]any{"res": "ok"},
		Error:      "",
		Duration:   "100ms",
		DurationMs: 100,
	}

	// Write entry
	err = store.Write(context.Background(), entry)
	assert.NoError(t, err)

	// Verify data in DB
	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)
	defer db.Close()

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM audit_logs").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	var toolName, userID, args, result, hash, prevHash string
	err = db.QueryRow("SELECT tool_name, user_id, arguments, result, hash, prev_hash FROM audit_logs").Scan(&toolName, &userID, &args, &result, &hash, &prevHash)
	require.NoError(t, err)
	assert.Equal(t, "test_tool", toolName)
	assert.Equal(t, "user123", userID)
	assert.JSONEq(t, `{"arg": "value"}`, args)
	assert.JSONEq(t, `{"res": "ok"}`, result)
	assert.NotEmpty(t, hash)
	assert.Empty(t, prevHash) // First entry has empty prev_hash

	// Test Verify
	valid, err := store.Verify()
	assert.NoError(t, err)
	assert.True(t, valid)
}

func TestSQLiteAuditStore_TamperEvident(t *testing.T) {
	// Create a temporary database file
	f, err := os.CreateTemp("", "audit_tamper_*.db")
	require.NoError(t, err)
	dbPath := f.Name()
	f.Close()
	defer os.Remove(dbPath)

	// Allow the temp path
	validation.SetAllowedPaths([]string{os.TempDir()})
	defer validation.SetAllowedPaths(nil)

	// Initialize store
	store, err := NewSQLiteAuditStore(dbPath)
	require.NoError(t, err)

	// Write 3 entries
	for i := 0; i < 3; i++ {
		entry := Entry{
			Timestamp:  time.Now(),
			ToolName:   "test_tool",
			DurationMs: int64(i),
		}
		require.NoError(t, store.Write(context.Background(), entry))
	}

	// Verify - should be valid
	valid, err := store.Verify()
	assert.NoError(t, err)
	assert.True(t, valid)

	store.Close()

	// Tamper with the database
	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)
	defer db.Close()

	// Modify the second entry (id=2)
	_, err = db.Exec("UPDATE audit_logs SET tool_name = 'hacked_tool' WHERE id = 2")
	require.NoError(t, err)
	db.Close()

	// Re-open store and verify
	store, err = NewSQLiteAuditStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	valid, err = store.Verify()
	// Should return error or false
	if err == nil {
		assert.False(t, valid, "Verify should return false for tampered logs")
	} else {
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "integrity violation")
	}
}

func TestSQLiteAuditStore_Migration(t *testing.T) {
	// Create a temporary database file
	f, err := os.CreateTemp("", "audit_migration_*.db")
	require.NoError(t, err)
	dbPath := f.Name()
	f.Close()
	defer os.Remove(dbPath)

	// Allow the temp path
	validation.SetAllowedPaths([]string{os.TempDir()})
	defer validation.SetAllowedPaths(nil)

	// Manually create legacy table
	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)
	schema := `
	CREATE TABLE audit_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp TEXT,
		tool_name TEXT,
		user_id TEXT,
		profile_id TEXT,
		arguments TEXT,
		result TEXT,
		error TEXT,
		duration_ms INTEGER
	);
	`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	// Insert a legacy record
	_, err = db.Exec("INSERT INTO audit_logs (timestamp, tool_name) VALUES (?, ?)", time.Now().Format(time.RFC3339Nano), "legacy_tool")
	require.NoError(t, err)
	db.Close()

	// Open store - should trigger migration
	store, err := NewSQLiteAuditStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	// Check if columns exist
	db, err = sql.Open("sqlite", dbPath)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec("SELECT hash, prev_hash FROM audit_logs LIMIT 1")
	assert.NoError(t, err, "Columns should exist after migration")

	// Verify behavior on legacy data
	// Since legacy data has empty hashes, verification should fail
	valid, err := store.Verify()
	// Verify should fail
	assert.False(t, valid)
	assert.Error(t, err)
}

func TestComputeHash_Collision(t *testing.T) {
	ts := time.Now().Format(time.RFC3339Nano)
	prevHash := "0000000000000000000000000000000000000000000000000000000000000000"

	// Entry 1: Tool="A", User="B|C"
	// Sprintf string was: "ts|A|B|C|..."
	hash1 := computeHash(ts, "A", "B|C", "profile", "{}", "{}", "", 100, prevHash)

	// Entry 2: Tool="A|B", User="C"
	// Sprintf string was: "ts|A|B|C|..."
	hash2 := computeHash(ts, "A|B", "C", "profile", "{}", "{}", "", 100, prevHash)

	// With JSON serialization, "A" and "B|C" becomes ["A", "B|C"]
	// "A|B" and "C" becomes ["A|B", "C"]
	// These are distinct JSON arrays, so hashes should differ.
	assert.NotEqual(t, hash1, hash2, "Hash collision detected! Different inputs produced the same hash.")
}
func TestSQLiteAuditStore_BackwardCompatibility(t *testing.T) {
	// Create a temporary database file
	f, err := os.CreateTemp("", "audit_compat_*.db")
	require.NoError(t, err)
	dbPath := f.Name()
	f.Close()
	defer os.Remove(dbPath)

	// Allow the temp path
	validation.SetAllowedPaths([]string{os.TempDir()})
	defer validation.SetAllowedPaths(nil)

	// Create store to init schema
	store, err := NewSQLiteAuditStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)
	defer db.Close()

	// Insert an entry manually using the LEGACY hashing logic
	ts := time.Date(2023, 10, 27, 12, 0, 0, 0, time.UTC).Format(time.RFC3339Nano)
	toolName := "old_tool"
	userID := "old_user"
	profileID := "old_prof"
	args := "{}"
	result := "{}"
	errorMsg := ""
	durationMs := int64(100)
	prevHash := "" // First entry

	legacyHash := computeHashV0(ts, toolName, userID, profileID, args, result, errorMsg, durationMs, prevHash)

	query := `
	INSERT INTO audit_logs (
		timestamp, tool_name, user_id, profile_id, arguments, result, error, duration_ms, prev_hash, hash
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = db.Exec(query, ts, toolName, userID, profileID, args, result, errorMsg, durationMs, prevHash, legacyHash)
	require.NoError(t, err)

	// Verify - should succeed using the fallback logic
	valid, err := store.Verify()
	assert.NoError(t, err)
	assert.True(t, valid, "Verification should pass for legacy hash")

	// Now add a NEW entry using the store (which uses new hashing)
	newEntry := Entry{
		Timestamp:  time.Now(),
		ToolName:   "new_tool",
		DurationMs: 200,
	}
	err = store.Write(context.Background(), newEntry)
	assert.NoError(t, err)

	// Verify again - both old and new should be valid
	valid, err = store.Verify()
	assert.NoError(t, err)
	assert.True(t, valid, "Verification should pass for mixed legacy and new hashes")
}

// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

func TestNewSQLiteAuditStore_EdgeCases(t *testing.T) {
	// Allow the temp path
	validation.SetAllowedPaths([]string{os.TempDir()})
	defer validation.SetAllowedPaths(nil)

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
	// Allow the temp path
	validation.SetAllowedPaths([]string{os.TempDir()})
	defer validation.SetAllowedPaths(nil)

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

	// Allow the temp path
	validation.SetAllowedPaths([]string{os.TempDir()})
	defer validation.SetAllowedPaths(nil)

	store, err := NewSQLiteAuditStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	// Close the DB underneath
	store.db.Close()

	entry := Entry{
		Timestamp: time.Now(),
		ToolName:  "test",
	}

	err = store.Write(context.Background(), entry)
	assert.Error(t, err)
}

func TestSQLiteAuditStore_Verify_Errors(t *testing.T) {
	f, err := os.CreateTemp("", "audit_verify_fail_*.db")
	require.NoError(t, err)
	dbPath := f.Name()
	f.Close()
	defer os.Remove(dbPath)

	// Allow the temp path
	validation.SetAllowedPaths([]string{os.TempDir()})
	defer validation.SetAllowedPaths(nil)

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

	// Allow the temp path
	validation.SetAllowedPaths([]string{os.TempDir()})
	defer validation.SetAllowedPaths(nil)

	store, err := NewSQLiteAuditStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	entry := Entry{
		Timestamp:  time.Now(),
		ToolName:   "complex_tool",
		Arguments:  nil, // Should become "{}"
		Result:     nil, // Should become "{}"
		DurationMs: 50,
	}

	err = store.Write(context.Background(), entry)
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

	// Allow the temp path
	validation.SetAllowedPaths([]string{os.TempDir()})
	defer validation.SetAllowedPaths(nil)

	store, err := NewSQLiteAuditStore(dbPath)
	require.NoError(t, err)

	// Write 2 entries
	require.NoError(t, store.Write(context.Background(), Entry{Timestamp: time.Now(), ToolName: "1"}))
	require.NoError(t, store.Write(context.Background(), Entry{Timestamp: time.Now(), ToolName: "2"}))
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

	// Allow the temp path
	validation.SetAllowedPaths([]string{os.TempDir()})
	defer validation.SetAllowedPaths(nil)

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

	// Allow the temp path
	validation.SetAllowedPaths([]string{os.TempDir()})
	defer validation.SetAllowedPaths(nil)

	store, err := NewSQLiteAuditStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	concurrency := 10
	done := make(chan bool)

	for i := 0; i < concurrency; i++ {
		go func(idx int) {
			err := store.Write(context.Background(), Entry{
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

func TestSQLiteAuditStore_Read(t *testing.T) {
	// Create a temporary database file
	f, err := os.CreateTemp("", "audit_read_*.db")
	require.NoError(t, err)
	dbPath := f.Name()
	f.Close()
	defer os.Remove(dbPath)

	// Allow the temp path
	validation.SetAllowedPaths([]string{os.TempDir()})
	defer validation.SetAllowedPaths(nil)

	// Initialize store
	store, err := NewSQLiteAuditStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	baseTime := time.Date(2023, 10, 27, 12, 0, 0, 0, time.UTC)

	// Write some entries
	entries := []Entry{
		{
			Timestamp:  baseTime,
			ToolName:   "tool1",
			UserID:     "user1",
			ProfileID:  "profile1",
			Arguments:  json.RawMessage(`{"arg": "1"}`),
			DurationMs: 10,
		},
		{
			Timestamp:  baseTime.Add(1 * time.Hour),
			ToolName:   "tool2",
			UserID:     "user2",
			ProfileID:  "profile2",
			Arguments:  json.RawMessage(`{"arg": "2"}`),
			DurationMs: 20,
		},
		{
			Timestamp:  baseTime.Add(2 * time.Hour),
			ToolName:   "tool1",
			UserID:     "user2",
			ProfileID:  "profile1",
			Arguments:  json.RawMessage(`{"arg": "3"}`),
			DurationMs: 30,
		},
	}

	for _, e := range entries {
		err := store.Write(context.Background(), e)
		require.NoError(t, err)
	}

	// Test Read All
	results, err := store.Read(context.Background(), Filter{})
	require.NoError(t, err)
	assert.Len(t, results, 3)
	// Results are ordered by timestamp DESC
	assert.Equal(t, "tool1", results[0].ToolName) // Last added
	assert.Equal(t, int64(30), results[0].DurationMs)

	// Test Filter by ToolName
	results, err = store.Read(context.Background(), Filter{ToolName: "tool1"})
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "tool1", results[0].ToolName)
	assert.Equal(t, "tool1", results[1].ToolName)

	// Test Filter by UserID
	results, err = store.Read(context.Background(), Filter{UserID: "user2"})
	require.NoError(t, err)
	assert.Len(t, results, 2)

	// Test Filter by Time Range
	startTime := baseTime.Add(30 * time.Minute)
	endTime := baseTime.Add(90 * time.Minute)
	results, err = store.Read(context.Background(), Filter{StartTime: &startTime, EndTime: &endTime})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "tool2", results[0].ToolName)

	// Test Limit and Offset
	results, err = store.Read(context.Background(), Filter{Limit: 1, Offset: 1})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	// Order DESC: 2h(tool1), 1h(tool2), 0h(tool1)
	// Offset 1 -> 1h(tool2)
	assert.Equal(t, "tool2", results[0].ToolName)
}
