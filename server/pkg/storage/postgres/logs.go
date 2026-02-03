// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package postgres

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
	VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err = s.db.Exec(query, entry.ID, entry.Timestamp, entry.Level, entry.Source, entry.Message, string(metadataJSON))
	return err
}

// QueryLogs retrieves logs based on the filter.
func (s *Store) QueryLogs(ctx context.Context, filter logging.LogFilter) ([]logging.LogEntry, int, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.StartTime != nil {
		conditions = append(conditions, fmt.Sprintf("timestamp >= $%d", argIdx))
		args = append(args, filter.StartTime.Format(time.RFC3339))
		argIdx++
	}
	if filter.EndTime != nil {
		conditions = append(conditions, fmt.Sprintf("timestamp <= $%d", argIdx))
		args = append(args, filter.EndTime.Format(time.RFC3339))
		argIdx++
	}
	if filter.Level != "" && filter.Level != "ALL" {
		conditions = append(conditions, fmt.Sprintf("level = $%d", argIdx))
		args = append(args, filter.Level)
		argIdx++
	}
	if filter.Source != "" && filter.Source != "ALL" {
		conditions = append(conditions, fmt.Sprintf("source = $%d", argIdx))
		args = append(args, filter.Source)
		argIdx++
	}
	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(message LIKE $%d OR source LIKE $%d)", argIdx, argIdx))
		pattern := "%" + filter.Search + "%"
		args = append(args, pattern)
		argIdx++
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
	query := fmt.Sprintf("SELECT id, timestamp, level, source, message, metadata FROM logs %s ORDER BY timestamp DESC LIMIT $%d OFFSET $%d", whereClause, argIdx, argIdx+1)
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
		var timestamp time.Time
		var source sql.NullString

		if err := rows.Scan(&entry.ID, &timestamp, &entry.Level, &source, &entry.Message, &metadataJSON); err != nil {
			return nil, 0, err
		}
		entry.Timestamp = timestamp.Format(time.RFC3339)
		if source.Valid {
			entry.Source = source.String
		}
		if metadataJSON != "" {
			if err := json.Unmarshal([]byte(metadataJSON), &entry.Metadata); err != nil {
				// log error?
			}
		}
		logs = append(logs, entry)
	}

	return logs, total, nil
}
