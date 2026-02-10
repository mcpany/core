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

func TestPersistenceAndHydration(t *testing.T) {
	// Setup temporary file for persistence
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "mcpany.log")

	// Reset logger state
	ForTestsOnlyResetLogger()
	// Reset broadcaster
	GlobalBroadcaster = NewBroadcaster()

	// Initialize logging with TEXT format but WITH persistence path
	var buf bytes.Buffer
	Init(slog.LevelInfo, &buf, "text", logFile)

	// Log a message
	logger := GetLogger()
	logger.Info("info message", "key", "value")
	logger.Warn("warn message")

	// Verify console output is TEXT
	consoleOutput := buf.String()
	assert.Contains(t, consoleOutput, "level=INFO")
	assert.Contains(t, consoleOutput, "msg=\"info message\"")
	assert.NotContains(t, consoleOutput, "{") // Should not look like JSON

	// Verify persistence file exists and contains JSON
	content, err := os.ReadFile(logFile)
	require.NoError(t, err)
	fileContent := string(content)
	assert.Contains(t, fileContent, "{\"") // Should start with JSON brace
	assert.Contains(t, fileContent, "\"level\":\"INFO\"")
	// slog uses "msg" by default, while LogEntry uses "message".
	// Since we use slog.JSONHandler for persistence, we expect "msg" in the file.
	assert.Contains(t, fileContent, "\"msg\":\"info message\"")

	// Verify Hydration
	// First, simulate a restart by resetting the Broadcaster
	GlobalBroadcaster = NewBroadcaster()
	// Broadcaster should be empty
	ch, history := GlobalBroadcaster.SubscribeWithHistory()
	assert.Empty(t, history)
	GlobalBroadcaster.Unsubscribe(ch)

	// Hydrate from file
	err = HydrateFromFile(logFile)
	require.NoError(t, err)

	// Verify history is populated
	ch, history = GlobalBroadcaster.SubscribeWithHistory()
	defer GlobalBroadcaster.Unsubscribe(ch)

	assert.Len(t, history, 2)

	// Check first message
	var entry1 LogEntry
	err = json.Unmarshal(history[0], &entry1)
	require.NoError(t, err)
	assert.Equal(t, "INFO", entry1.Level)
	assert.Equal(t, "info message", entry1.Message)

	// Check second message
	var entry2 LogEntry
	err = json.Unmarshal(history[1], &entry2)
	require.NoError(t, err)
	assert.Equal(t, "WARN", entry2.Level)
	assert.Equal(t, "warn message", entry2.Message)
}
