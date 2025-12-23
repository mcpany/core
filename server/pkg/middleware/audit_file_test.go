// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileAuditStore(t *testing.T) {
	// Create a temporary file
	f, err := os.CreateTemp("", "audit_test_*.log")
	require.NoError(t, err)
	logPath := f.Name()
	f.Close()
	defer os.Remove(logPath)

	// Initialize store
	store, err := NewFileAuditStore(logPath)
	require.NoError(t, err)

	// Create a sample entry
	entry := AuditEntry{
		Timestamp:  time.Date(2023, 10, 27, 12, 0, 0, 0, time.UTC),
		ToolName:   "test_tool",
		UserID:     "user123",
		Arguments:  json.RawMessage(`{"arg": "value"}`),
		Result:     map[string]any{"res": "ok"},
		DurationMs: 100,
	}

	// Write entry
	err = store.Write(entry)
	assert.NoError(t, err)
	store.Close()

	// Verify file content
	content, err := os.ReadFile(logPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), `"hash":`)
	assert.Contains(t, string(content), `"prev_hash":""`)

	// Re-open store to test chaining
	store, err = NewFileAuditStore(logPath)
	require.NoError(t, err)

	entry2 := AuditEntry{
		Timestamp:  time.Date(2023, 10, 27, 12, 0, 1, 0, time.UTC),
		ToolName:   "test_tool_2",
		DurationMs: 200,
	}
	err = store.Write(entry2)
	assert.NoError(t, err)
	store.Close()

	// Verify chain
	valid, err := store.Verify()
	assert.NoError(t, err)
	assert.True(t, valid)
}

func TestFileAuditStore_TamperEvident(t *testing.T) {
	// Create a temporary file
	f, err := os.CreateTemp("", "audit_tamper_*.log")
	require.NoError(t, err)
	logPath := f.Name()
	f.Close()
	defer os.Remove(logPath)

	// Initialize store
	store, err := NewFileAuditStore(logPath)
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
	store.Close()

	// Verify - should be valid
	store, err = NewFileAuditStore(logPath)
	require.NoError(t, err)
	valid, err := store.Verify()
	assert.NoError(t, err)
	assert.True(t, valid)
	store.Close()

	// Tamper with the file
	// Read lines
	content, err := os.ReadFile(logPath)
	require.NoError(t, err)
	// Replace "test_tool" with "hacked_tool" in the second line (index 1)
	// Simple string replacement might affect hash if hash contains "test_tool" (it does)
	// But simply modifying the text invalidates the hash signature on that line.
	// And if we recalculate that hash, it invalidates the prev_hash of the next line.

	// Let's just modify the content blindly.
	tamperedContent := []byte(nil)
	// We need to find the newline
	nl1 := -1
	nl2 := -1
	for i, b := range content {
		if b == '\n' {
			if nl1 == -1 {
				nl1 = i
			} else if nl2 == -1 {
				nl2 = i
				break
			}
		}
	}

	// Inject garbage in the middle
	tamperedContent = append(tamperedContent, content[:nl1+1]...)
	tamperedContent = append(tamperedContent, []byte(`{"invalid": "json"}`)...)
	tamperedContent = append(tamperedContent, '\n')
	tamperedContent = append(tamperedContent, content[nl2+1:]...)

	err = os.WriteFile(logPath, tamperedContent, 0644)
	require.NoError(t, err)

	// Verify
	store, err = NewFileAuditStore(logPath)
	require.NoError(t, err) // Open might succeed even if file is bad
	valid, err = store.Verify()
	assert.False(t, valid)
	assert.Error(t, err)
}
