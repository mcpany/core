// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileAuditStore(t *testing.T) {
	t.Run("write to file", func(t *testing.T) {
		tmpDir := t.TempDir()
		logFile := filepath.Join(tmpDir, "audit.log")

		store, err := NewFileAuditStore(logFile)
		require.NoError(t, err)
		defer store.Close()

		entry := AuditEntry{
			Timestamp:  time.Now(),
			ToolName:   "test-tool",
			UserID:     "user-123",
			Duration:   "100ms",
			DurationMs: 100,
		}

		err = store.Write(entry)
		require.NoError(t, err)

		// Read file content
		content, err := os.ReadFile(logFile)
		require.NoError(t, err)

		var readEntry AuditEntry
		err = json.Unmarshal(content, &readEntry)
		require.NoError(t, err)

		assert.Equal(t, entry.ToolName, readEntry.ToolName)
		assert.Equal(t, entry.UserID, readEntry.UserID)
		assert.Equal(t, entry.Duration, readEntry.Duration)
		// Timestamp comparison might be tricky due to precision/marshaling, so skipping exact timestamp check
	})

	t.Run("write to stdout (nil file)", func(t *testing.T) {
		// NewFileAuditStore with empty path uses stdout (handled internally)
		// Testing stdout capture is tricky, but we can verify it doesn't crash or error.
		store, err := NewFileAuditStore("")
		require.NoError(t, err)
		defer store.Close() // Should be safe to call on nil file

		entry := AuditEntry{
			Timestamp:  time.Now(),
			ToolName:   "test-stdout",
		}

		err = store.Write(entry)
		require.NoError(t, err)
	})

	t.Run("invalid file path", func(t *testing.T) {
		// Using a directory as file path should fail
		tmpDir := t.TempDir()
		_, err := NewFileAuditStore(tmpDir)
		require.Error(t, err)
	})
}
