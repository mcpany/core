// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteLogStore(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := NewSQLiteLogStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	// 1. Test Write
	entry1 := LogEntry{
		ID:        uuid.New().String(),
		Timestamp: time.Now().Add(-2 * time.Minute).Format(time.RFC3339),
		Level:     "INFO",
		Message:   "Test message 1",
		Source:    "test",
		Metadata:  map[string]any{"key": "value"},
	}
	err = store.Write(entry1)
	require.NoError(t, err)

	entry2 := LogEntry{
		ID:        uuid.New().String(),
		Timestamp: time.Now().Add(-1 * time.Minute).Format(time.RFC3339),
		Level:     "ERROR",
		Message:   "Test message 2",
		Source:    "test",
	}
	err = store.Write(entry2)
	require.NoError(t, err)

	// 2. Test Read
	entries, err := store.Read(10)
	require.NoError(t, err)
	require.Len(t, entries, 2)

	// Check order (should be chronological: entry1 then entry2)
	assert.Equal(t, entry1.ID, entries[0].ID)
	assert.Equal(t, entry1.Message, entries[0].Message)
	// Metadata check
	// Note: JSON unmarshal might change types (e.g. numbers to float64), so direct map comparison might be tricky depending on "value" type.
	// Here "value" is string, so it should be fine.
	assert.Equal(t, "value", entries[0].Metadata["key"])

	assert.Equal(t, entry2.ID, entries[1].ID)

	// 3. Test Limit
	entries, err = store.Read(1)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	// Should return the LATEST entry (entry2) because Read sorts by timestamp DESC then limits, then reverses.
	// Wait, if I limit 1 on DESC sort, I get the NEWEST.
	// Then I reverse it. So I get [entry2].
	assert.Equal(t, entry2.ID, entries[0].ID)
}

func TestSQLiteLogStore_Persistence(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "persist.db")
	expectedLogPath := filepath.Join(tmpDir, "persist_logs.db")

	// Open first time
	store1, err := NewSQLiteLogStore(dbPath)
	require.NoError(t, err)

	entry := LogEntry{
		ID:        uuid.New().String(),
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "INFO",
		Message:   "Persisted message",
	}
	require.NoError(t, store1.Write(entry))
	require.NoError(t, store1.Close())

	// Verify file exists
	_, err = os.Stat(expectedLogPath)
	require.NoError(t, err)

	// Open second time
	store2, err := NewSQLiteLogStore(dbPath)
	require.NoError(t, err)
	defer store2.Close()

	entries, err := store2.Read(10)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, entry.Message, entries[0].Message)
}
