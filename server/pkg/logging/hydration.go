// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/mcpany/core/server/pkg/util"
)

// HydrateFromFile hydrateFromFile hydrate from file.
//
// Summary: HydrateFromFile hydrate from file.
//
// Parameters: - None.
//   - path (string): The path.
//
// Returns: - None.
//   - error: An error if the operation fails.
func HydrateFromFile(path string) error {
	lines, err := util.ReadLastNLines(path, 1000)
	if err != nil {
		return err
	}

	// ⚡ BOLT: Store structs in history, not bytes.
	broadcastMessages := make([]any, 0, len(lines))
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

		broadcastMessages = append(broadcastMessages, entry)
	}

	if len(broadcastMessages) > 0 {
		GlobalBroadcaster.Hydrate(broadcastMessages)
	}

	return nil
}
