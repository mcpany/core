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

	var toolName, userID, args, result string
	err = db.QueryRow("SELECT tool_name, user_id, arguments, result FROM audit_logs").Scan(&toolName, &userID, &args, &result)
	require.NoError(t, err)
	assert.Equal(t, "test_tool", toolName)
	assert.Equal(t, "user123", userID)
	assert.JSONEq(t, `{"arg": "value"}`, args)
	assert.JSONEq(t, `{"res": "ok"}`, result)
}
