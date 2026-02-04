// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mcpany/core/server/pkg/logging"
)

// SaveLog persists a log entry.
func (s *Store) SaveLog(entry logging.LogEntry) error {
	metadataJSON, err := json.Marshal(entry.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
	INSERT INTO logs (id, timestamp, level, source, message, metadata)
	VALUES (?, ?, ?, ?, ?, ?)
	`
	// SQLite timestamp format: RFC3339
	// The entry.Timestamp is string, typically RFC3339 already from BroadcastHandler.

	_, err = s.db.ExecContext(context.Background(), query, entry.ID, entry.Timestamp, entry.Level, entry.Source, entry.Message, string(metadataJSON))
	return err
}

// QueryLogs retrieves logs based on the filter.
func (s *Store) QueryLogs(ctx context.Context, filter logging.LogFilter) ([]logging.LogEntry, int, error) {
	var conditions []string
	var args []interface{}

	if filter.StartTime != nil {
		conditions = append(conditions, "timestamp >= ?")
		args = append(args, filter.StartTime.Format(time.RFC3339))
	}
	if filter.EndTime != nil {
		conditions = append(conditions, "timestamp <= ?")
		args = append(args, filter.EndTime.Format(time.RFC3339))
	}
	if filter.Level != "" && filter.Level != "ALL" {
		conditions = append(conditions, "level = ?")
		args = append(args, filter.Level)
	}
	if filter.Source != "" && filter.Source != "ALL" {
		conditions = append(conditions, "source = ?")
		args = append(args, filter.Source)
	}
	if filter.Search != "" {
		conditions = append(conditions, "(message LIKE ? OR source LIKE ?)")
		pattern := "%" + filter.Search + "%"
		args = append(args, pattern, pattern)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM logs " + whereClause
	var total int
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count logs: %w", err)
	}

	// Query logs
	query := fmt.Sprintf("SELECT id, timestamp, level, source, message, metadata FROM logs %s ORDER BY timestamp DESC LIMIT ? OFFSET ?", whereClause)
	args = append(args, filter.Limit)
	args = append(args, filter.Offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query logs: %w", err)
	}
	defer rows.Close()

	var logs []logging.LogEntry
	for rows.Next() {
		var entry logging.LogEntry
		var metadataJSON string
		var source sql.NullString // Handle null source

		if err := rows.Scan(&entry.ID, &entry.Timestamp, &entry.Level, &source, &entry.Message, &metadataJSON); err != nil {
			return nil, 0, err
		}
		if source.Valid {
			entry.Source = source.String
		}
		if metadataJSON != "" {
			if err := json.Unmarshal([]byte(metadataJSON), &entry.Metadata); err != nil {
				// Ignore malformed metadata in logs to avoid breaking the list
				_ = err
			}
		}
		logs = append(logs, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating logs: %w", err)
	}

	return logs, total, nil
}
