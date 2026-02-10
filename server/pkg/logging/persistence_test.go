// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogPersistenceAndHydration(t *testing.T) {
	// Create a temporary directory for logs
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	// 1. Initialize Logger with File Persistence
	ForTestsOnlyResetLogger()
	var buf1 bytes.Buffer
	Init(slog.LevelInfo, &buf1, logFile, "text")

	logger := GetLogger()
	logger.Info("Message 1", "key", "value1")
	logger.Warn("Message 2", "key", "value2")

	// Verify file was written
	content, err := os.ReadFile(logFile)
	require.NoError(t, err)
	assert.NotEmpty(t, content)
	assert.Contains(t, string(content), "Message 1")
	assert.Contains(t, string(content), "Message 2")

	// 2. Simulate Restart: Hydrate from File
	// Reset Broadcaster to clear in-memory history
	GlobalBroadcaster = NewBroadcaster()

	err = HydrateFromFile(logFile)
	require.NoError(t, err)

	// 3. Verify GlobalBroadcaster has history
	history := GlobalBroadcaster.GetHistory()
	require.Len(t, history, 2, "Should have 2 messages in history")

	// Parse history to verify content
	var entry1, entry2 LogEntry
	err = json.Unmarshal(history[0], &entry1)
	require.NoError(t, err)
	err = json.Unmarshal(history[1], &entry2)
	require.NoError(t, err)

	assert.Equal(t, "INFO", entry1.Level)
	assert.Equal(t, "Message 1", entry1.Message)
	assert.Equal(t, "value1", entry1.Metadata["key"])

	assert.Equal(t, "WARN", entry2.Level)
	assert.Equal(t, "Message 2", entry2.Message)
	assert.Equal(t, "value2", entry2.Metadata["key"])
}
