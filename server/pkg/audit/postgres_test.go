package audit

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestPostgresAuditStore_New(t *testing.T) {
	// Mock DB
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// We can't inject the mock DB directly into NewPostgresAuditStore because it calls sql.Open.
	// So we'll test the struct directly or need to refactor NewPostgresAuditStore to accept a DB (which changes the pattern).
	// However, for unit testing the logic, we can construct the struct manually.

	// But wait, NewPostgresAuditStore does schema initialization.
	// If we want to test NewPostgresAuditStore, we need to mock sql.Open which is hard in Go without a wrapper.
	// For this exercise, we will test the methods Write and Verify by manually constructing the store with the mock DB.
}

func TestPostgresAuditStore_Write(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	store := &PostgresAuditStore{
		db: db,
	}

	entry := Entry{
		Timestamp:  time.Now(),
		ToolName:   "test_tool",
		UserID:     "user1",
		ProfileID:  "p1",
		Arguments:  []byte(`{"arg":"val"}`),
		Result:     map[string]any{"res": "val"},
		Error:      "",
		DurationMs: 100,
	}

	mock.ExpectBegin()
	mock.ExpectExec("LOCK TABLE audit_logs IN EXCLUSIVE MODE").WillReturnResult(sqlmock.NewResult(0, 0))

	// Mock getting previous hash
	mock.ExpectQuery("SELECT hash FROM audit_logs ORDER BY id DESC LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"hash"}).AddRow("prev_hash"))

	// Mock Insert
	mock.ExpectExec("INSERT INTO audit_logs").
		WithArgs(
			entry.Timestamp,
			entry.ToolName,
			entry.UserID,
			entry.ProfileID,
			string(entry.Arguments),
			`{"res":"val"}`,
			entry.Error,
			entry.DurationMs,
			"prev_hash",
			sqlmock.AnyArg(), // The new hash
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = store.Write(context.Background(), entry)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresAuditStore_Verify_Valid(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	store := &PostgresAuditStore{
		db: db,
	}

	ts := time.Now()
	tsStr := ts.Format(time.RFC3339Nano)
	toolName := "tool1"
	userID := "u1"
	profileID := "p1"
	args := "{}"
	result := "{}"
	errStr := ""
	dur := int64(10)
	prevHash := ""

	hash := computeHash(tsStr, toolName, userID, profileID, args, result, errStr, dur, prevHash)

	rows := sqlmock.NewRows([]string{"id", "timestamp", "tool_name", "user_id", "profile_id", "arguments", "result", "error", "duration_ms", "prev_hash", "hash"}).
		AddRow(1, ts, toolName, userID, profileID, args, result, errStr, dur, prevHash, hash)

	mock.ExpectQuery("SELECT .* FROM audit_logs ORDER BY id ASC").WillReturnRows(rows)

	valid, err := store.Verify()
	assert.NoError(t, err)
	assert.True(t, valid)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresAuditStore_Verify_Tampered(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	store := &PostgresAuditStore{
		db: db,
	}

	ts := time.Now()
	// Tampered data: tool name changed but hash matches original
	rows := sqlmock.NewRows([]string{"id", "timestamp", "tool_name", "user_id", "profile_id", "arguments", "result", "error", "duration_ms", "prev_hash", "hash"}).
		AddRow(1, ts, "hacked_tool", "u1", "p1", "{}", "{}", "", 10, "", "some_valid_hash_for_original_data")

	mock.ExpectQuery("SELECT .* FROM audit_logs ORDER BY id ASC").WillReturnRows(rows)

	valid, err := store.Verify()
	assert.Error(t, err) // Should return error about integrity
	assert.False(t, valid)
	assert.Contains(t, err.Error(), "hash mismatch")
}

func TestPostgresAuditStore_Verify_ChainBroken(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	store := &PostgresAuditStore{
		db: db,
	}

	ts := time.Now()
	tsStr := ts.Format(time.RFC3339Nano)
	h1 := computeHash(tsStr, "t1", "u1", "p1", "{}", "{}", "", 10, "")
	// Row 2 claims prev_hash is h1, but we pass "tampered"
	h2 := computeHash(tsStr, "t2", "u1", "p1", "{}", "{}", "", 10, "tampered")

	rows := sqlmock.NewRows([]string{"id", "timestamp", "tool_name", "user_id", "profile_id", "arguments", "result", "error", "duration_ms", "prev_hash", "hash"}).
		AddRow(1, ts, "t1", "u1", "p1", "{}", "{}", "", 10, "", h1).
		AddRow(2, ts, "t2", "u1", "p1", "{}", "{}", "", 10, "tampered", h2)

	mock.ExpectQuery("SELECT .* FROM audit_logs ORDER BY id ASC").WillReturnRows(rows)

	valid, err := store.Verify()
	assert.Error(t, err)
	assert.False(t, valid)
	assert.Contains(t, err.Error(), "prev_hash mismatch")
}

func TestPostgresAuditStore_New_Error(t *testing.T) {
	// Simple check for DSN validation
	_, err := NewPostgresAuditStore("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dsn is required")
}
