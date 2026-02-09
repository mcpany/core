// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestSQLiteLogStore(t *testing.T) {
	// Use in-memory DB
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	store, err := NewSQLiteLogStore(db)
	require.NoError(t, err)

	ctx := context.Background()

	// 1. Test Write
	entry1 := LogEntry{
		ID:        "1",
		Timestamp: time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
		Level:     "INFO",
		Message:   "Test log 1",
		Source:    "test",
		Metadata:  map[string]any{"foo": "bar"},
	}
	err = store.Write(ctx, entry1)
	require.NoError(t, err)

	entry2 := LogEntry{
		ID:        "2",
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "ERROR",
		Message:   "Test log 2",
		Source:    "test",
	}
	err = store.Write(ctx, entry2)
	require.NoError(t, err)

	// 2. Test Read All
	logs, err := store.Read(ctx, 10, 0)
	require.NoError(t, err)
	assert.Len(t, logs, 2)
	// Expect chronological order (oldest first)
	assert.Equal(t, "1", logs[0].ID)
	assert.Equal(t, "2", logs[1].ID)
	assert.Equal(t, "bar", logs[0].Metadata["foo"])

	// 3. Test Filter by Level
	logs, err = store.Query(ctx, LogQueryOptions{Level: "ERROR"})
	require.NoError(t, err)
	assert.Len(t, logs, 1)
	assert.Equal(t, "2", logs[0].ID)

	// 4. Test Filter by Search
	logs, err = store.Query(ctx, LogQueryOptions{Search: "log 1"})
	require.NoError(t, err)
	assert.Len(t, logs, 1)
	assert.Equal(t, "1", logs[0].ID)

	// 5. Test Pagination
	// We have 2 logs. Limit 1.
	// Query uses DESC order then reverses.
	// DB: 2 (Newest), 1 (Oldest).
	// Limit 1 -> Gets "2".
	// Reverse -> "2".
	logs, err = store.Read(ctx, 1, 0)
	require.NoError(t, err)
	assert.Len(t, logs, 1)
	assert.Equal(t, "2", logs[0].ID) // Most recent

	// Offset 1 -> Gets "1".
	logs, err = store.Read(ctx, 1, 1)
	require.NoError(t, err)
	assert.Len(t, logs, 1)
	assert.Equal(t, "1", logs[0].ID)
}
