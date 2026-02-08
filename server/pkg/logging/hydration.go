package logging

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/mcpany/core/server/pkg/util"
)

// HydrateFromFile reads the last N lines from the given log file,
// parses them (assuming JSON format), and populates the global broadcaster's history.
// This allows the UI to show historical logs upon connection, even after a server restart.
func HydrateFromFile(path string) error {
	lines, err := util.ReadLastNLines(path, 1000)
	if err != nil {
		return err
	}

	var broadcastMessages [][]byte
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		// Parse slog JSON
		var rawMap map[string]interface{}
		if err := json.Unmarshal(line, &rawMap); err != nil {
			continue // Skip malformed lines
		}

		// Map to LogEntry
		entry := LogEntry{
			ID:       uuid.New().String(),
			Metadata: make(map[string]any),
		}

		// Extract known fields
		if t, ok := rawMap["time"].(string); ok {
			entry.Timestamp = t
		} else {
			entry.Timestamp = time.Now().Format(time.RFC3339)
		}
		delete(rawMap, "time")

		if l, ok := rawMap["level"].(string); ok {
			entry.Level = l
		}
		delete(rawMap, "level")

		if m, ok := rawMap["msg"].(string); ok {
			entry.Message = m
		}
		delete(rawMap, "msg")

		if src, ok := rawMap["source"].(string); ok {
			entry.Source = src
		}
		delete(rawMap, "source")

		// Everything else goes to Metadata
		for k, v := range rawMap {
			entry.Metadata[k] = v
		}

		// Marshal to LogEntry JSON
		data, err := json.Marshal(entry)
		if err == nil {
			broadcastMessages = append(broadcastMessages, data)
		}
	}

	if len(broadcastMessages) > 0 {
		GlobalBroadcaster.Hydrate(broadcastMessages)
	}

	return nil
}
