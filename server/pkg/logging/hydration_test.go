// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHydrateFromFile(t *testing.T) {
	// Helper to reset GlobalBroadcaster to a clean state
	resetBroadcaster := func(t *testing.T) {
		old := GlobalBroadcaster
		GlobalBroadcaster = NewBroadcaster()
		t.Cleanup(func() {
			GlobalBroadcaster = old
		})
	}

	t.Run("File Not Found", func(t *testing.T) {
		resetBroadcaster(t)
		err := HydrateFromFile("non-existent-file.log")
		assert.Error(t, err)
	})

	t.Run("Empty File", func(t *testing.T) {
		resetBroadcaster(t)
		f, err := os.CreateTemp("", "empty-*.log")
		require.NoError(t, err)
		defer os.Remove(f.Name())
		f.Close()

		err = HydrateFromFile(f.Name())
		assert.NoError(t, err)
		assert.Empty(t, GlobalBroadcaster.GetHistory())
	})

	t.Run("Valid JSON Logs", func(t *testing.T) {
		resetBroadcaster(t)
		f, err := os.CreateTemp("", "valid-*.log")
		require.NoError(t, err)
		defer os.Remove(f.Name())

		// Write some logs
		logs := []map[string]interface{}{
			{
				"time":   "2023-10-27T10:00:00Z",
				"level":  "INFO",
				"msg":    "First message",
				"source": "component-a",
				"extra":  "value1",
			},
			{
				"time":    "2023-10-27T10:01:00Z",
				"level":   "ERROR",
				"msg":     "Second message",
				"details": 123,
			},
		}

		for _, l := range logs {
			data, err := json.Marshal(l)
			require.NoError(t, err)
			_, err = f.Write(data)
			require.NoError(t, err)
			_, err = f.WriteString("\n")
			require.NoError(t, err)
		}
		f.Close()

		err = HydrateFromFile(f.Name())
		assert.NoError(t, err)

		history := GlobalBroadcaster.GetHistory()
		require.Len(t, history, 2)

		var entry1 LogEntry
		err = json.Unmarshal(history[0], &entry1)
		require.NoError(t, err)
		assert.Equal(t, "First message", entry1.Message)
		assert.Equal(t, "INFO", entry1.Level)
		assert.Equal(t, "component-a", entry1.Source)
		assert.Equal(t, "value1", entry1.Metadata["extra"])

		var entry2 LogEntry
		err = json.Unmarshal(history[1], &entry2)
		require.NoError(t, err)
		assert.Equal(t, "Second message", entry2.Message)
		assert.Equal(t, "ERROR", entry2.Level)
		assert.Equal(t, float64(123), entry2.Metadata["details"]) // JSON numbers are float64
	})

	t.Run("Malformed JSON Ignored", func(t *testing.T) {
		resetBroadcaster(t)
		f, err := os.CreateTemp("", "malformed-*.log")
		require.NoError(t, err)
		defer os.Remove(f.Name())

		_, err = f.WriteString("Not JSON\n")
		require.NoError(t, err)
		_, err = f.WriteString(`{"valid": "json"}` + "\n")
		require.NoError(t, err)
		_, err = f.WriteString("{Unfinished JSON\n")
		require.NoError(t, err)
		f.Close()

		err = HydrateFromFile(f.Name())
		assert.NoError(t, err)

		history := GlobalBroadcaster.GetHistory()
		require.Len(t, history, 1) // Only one valid line

		var entry LogEntry
		err = json.Unmarshal(history[0], &entry)
		require.NoError(t, err)
		assert.Equal(t, "json", entry.Metadata["valid"])
	})

	t.Run("Field Mapping Defaults", func(t *testing.T) {
		resetBroadcaster(t)
		f, err := os.CreateTemp("", "defaults-*.log")
		require.NoError(t, err)
		defer os.Remove(f.Name())

		// Minimal JSON
		_, err = f.WriteString(`{"custom": "data"}` + "\n")
		require.NoError(t, err)
		f.Close()

		err = HydrateFromFile(f.Name())
		assert.NoError(t, err)

		history := GlobalBroadcaster.GetHistory()
		require.Len(t, history, 1)

		var entry LogEntry
		err = json.Unmarshal(history[0], &entry)
		require.NoError(t, err)
		assert.NotEmpty(t, entry.ID)
		assert.NotEmpty(t, entry.Timestamp) // Should default to now
		assert.Equal(t, "data", entry.Metadata["custom"])
	})
}
