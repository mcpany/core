// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"database/sql"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteAuditStore(t *testing.T) {
	// Create a temporary database file
	f, err := os.CreateTemp("", "audit_test_*.db")
	require.NoError(t, err)
	dbPath := f.Name()
	f.Close()
	defer os.Remove(dbPath)

	// Initialize store
	store, err := NewSQLiteAuditStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	// Create a sample entry
	entry := AuditEntry{
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
	err = store.Write(entry)
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

	// Initialize store
	store, err := NewSQLiteAuditStore(dbPath)
	require.NoError(t, err)

	// Write 3 entries
	for i := 0; i < 3; i++ {
		entry := AuditEntry{
			Timestamp:  time.Now(),
			ToolName:   "test_tool",
			DurationMs: int64(i),
		}
		require.NoError(t, store.Write(entry))
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
	newEntry := AuditEntry{
		Timestamp:  time.Now(),
		ToolName:   "new_tool",
		DurationMs: 200,
	}
	err = store.Write(newEntry)
	assert.NoError(t, err)

	// Verify again - both old and new should be valid
	valid, err = store.Verify()
	assert.NoError(t, err)
	assert.True(t, valid, "Verification should pass for mixed legacy and new hashes")
}
