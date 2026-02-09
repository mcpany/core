// Copyright 2026 Author(s) of MCP Any
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
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	store, err := NewSQLiteLogStore(db)
	require.NoError(t, err)

	ctx := context.Background()

	// 1. Write Logs
	entry1 := LogEntry{
		Level:   "INFO",
		Source:  "test-source",
		Message: "test message 1",
		Metadata: map[string]any{
			"key": "value",
		},
	}
	err = store.Write(ctx, entry1)
	require.NoError(t, err)

	entry2 := LogEntry{
		Level:   "ERROR",
		Source:  "test-source",
		Message: "test error",
	}
	err = store.Write(ctx, entry2)
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond) // Ensure timestamp diff if needed, though they are sequential

	entry3 := LogEntry{
		Level:   "INFO",
		Source:  "other-source",
		Message: "other message",
	}
	err = store.Write(ctx, entry3)
	require.NoError(t, err)

	// 2. Read Logs (All)
	logs, err := store.Read(ctx, LogQueryOptions{Limit: 10})
	require.NoError(t, err)
	assert.Len(t, logs, 3)
	// Order is DESC
	assert.Equal(t, entry3.Message, logs[0].Message)

	// 3. Filter by Level
	logs, err = store.Read(ctx, LogQueryOptions{Level: "ERROR"})
	require.NoError(t, err)
	assert.Len(t, logs, 1)
	assert.Equal(t, entry2.Message, logs[0].Message)

	// 4. Filter by Source
	logs, err = store.Read(ctx, LogQueryOptions{Source: "test-source"})
	require.NoError(t, err)
	assert.Len(t, logs, 2)

	// 5. Search
	logs, err = store.Read(ctx, LogQueryOptions{Search: "error"})
	require.NoError(t, err)
	assert.Len(t, logs, 1)
	assert.Equal(t, entry2.Message, logs[0].Message)

	// 6. Pagination
	logs, err = store.Read(ctx, LogQueryOptions{Limit: 1, Offset: 0})
	require.NoError(t, err)
	assert.Len(t, logs, 1)
	assert.Equal(t, entry3.Message, logs[0].Message)

	logs, err = store.Read(ctx, LogQueryOptions{Limit: 1, Offset: 1})
	require.NoError(t, err)
	assert.Len(t, logs, 1)
	assert.Equal(t, entry2.Message, logs[0].Message)
}
