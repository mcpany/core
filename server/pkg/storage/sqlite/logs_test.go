// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package sqlite

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogStorage(t *testing.T) {
	dbPath := "test_logs.db"
	defer os.Remove(dbPath)

	db, err := NewDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	store := NewStore(db)

	entry := logging.LogEntry{
		ID:        uuid.New().String(),
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "INFO",
		Message:   "Test log message",
		Source:    "test-source",
		Metadata:  map[string]any{"key": "value"},
	}

	// Test SaveLog
	err = store.SaveLog(entry)
	require.NoError(t, err)

	// Test QueryLogs
	logs, total, err := store.QueryLogs(context.Background(), logging.LogFilter{
		Limit: 10,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, logs, 1)
	assert.Equal(t, entry.ID, logs[0].ID)
	assert.Equal(t, entry.Message, logs[0].Message)
	assert.Equal(t, "value", logs[0].Metadata["key"])

	// Test Filter
	logs, total, err = store.QueryLogs(context.Background(), logging.LogFilter{
		Search: "Test log",
		Limit:  10,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, total)

	logs, total, err = store.QueryLogs(context.Background(), logging.LogFilter{
		Search: "Non-existent",
		Limit:  10,
	})
	require.NoError(t, err)
	assert.Equal(t, 0, total)
}
