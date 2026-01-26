// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package sqlite

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mcpany/core/server/pkg/logging"
)

// SaveLog saves a log entry to the store.
//
// ctx is the context for the request.
// entry is the log entry.
//
// Returns an error if the operation fails.
func (s *Store) SaveLog(ctx context.Context, entry logging.LogEntry) error {
	metadataJSON, err := json.Marshal(entry.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal log metadata: %w", err)
	}

	query := `
	INSERT INTO system_logs (id, timestamp, level, source, message, metadata_json)
	VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err = s.db.ExecContext(ctx, query, entry.ID, entry.Timestamp, entry.Level, entry.Source, entry.Message, string(metadataJSON))
	if err != nil {
		return fmt.Errorf("failed to save log: %w", err)
	}
	return nil
}

// GetRecentLogs retrieves the most recent log entries.
//
// ctx is the context for the request.
// limit is the maximum number of logs to return.
//
// Returns the result.
// Returns an error if the operation fails.
func (s *Store) GetRecentLogs(ctx context.Context, limit int) ([]logging.LogEntry, error) {
	query := `
	SELECT id, timestamp, level, source, message, metadata_json
	FROM (
		SELECT id, timestamp, level, source, message, metadata_json
		FROM system_logs
		ORDER BY timestamp DESC
		LIMIT ?
	)
	ORDER BY timestamp ASC
	`

	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query system_logs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var logs []logging.LogEntry
	for rows.Next() {
		var entry logging.LogEntry
		var metadataJSON string
		var timestampStr string

		if err := rows.Scan(&entry.ID, &timestampStr, &entry.Level, &entry.Source, &entry.Message, &metadataJSON); err != nil {
			return nil, fmt.Errorf("failed to scan log entry: %w", err)
		}
		entry.Timestamp = timestampStr

		if metadataJSON != "" {
			if err := json.Unmarshal([]byte(metadataJSON), &entry.Metadata); err != nil {
				// Just continue with empty metadata if unmarshal fails
			}
		}
		logs = append(logs, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return logs, nil
}
