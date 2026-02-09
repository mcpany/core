// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
)

// LogStore interface defines the methods for a log persistence layer.
type LogStore interface {
	Write(ctx context.Context, entry LogEntry) error
	Read(ctx context.Context, limit int, offset int) ([]LogEntry, error)
	Query(ctx context.Context, opts LogQueryOptions) ([]LogEntry, error)
}

// LogQueryOptions defines filters for querying logs.
type LogQueryOptions struct {
	Limit  int
	Offset int
	Level  string
	Source string
	Search string
}

// SQLiteLogStore implements LogStore using SQLite.
type SQLiteLogStore struct {
	db *sql.DB
}

// NewSQLiteLogStore creates a new SQLiteLogStore.
// It initializes the schema if necessary.
func NewSQLiteLogStore(db *sql.DB) (*SQLiteLogStore, error) {
	s := &SQLiteLogStore{db: db}
	if err := s.init(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *SQLiteLogStore) init() error {
	query := `
	CREATE TABLE IF NOT EXISTS logs (
		id TEXT PRIMARY KEY,
		timestamp TEXT NOT NULL,
		level TEXT NOT NULL,
		source TEXT,
		message TEXT,
		metadata TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp DESC);
    CREATE INDEX IF NOT EXISTS idx_logs_level ON logs(level);
    CREATE INDEX IF NOT EXISTS idx_logs_source ON logs(source);
	`
	_, err := s.db.Exec(query)
	return err
}

// Write inserts a log entry into the store.
func (s *SQLiteLogStore) Write(ctx context.Context, entry LogEntry) error {
	meta, err := json.Marshal(entry.Metadata)
	if err != nil {
		return err
	}
	// We use ExecContext for proper timeout handling if context is provided
	_, err = s.db.ExecContext(ctx, "INSERT INTO logs (id, timestamp, level, source, message, metadata) VALUES (?, ?, ?, ?, ?, ?)",
		entry.ID, entry.Timestamp, entry.Level, entry.Source, entry.Message, string(meta))
	return err
}

// Read retrieves logs with pagination.
func (s *SQLiteLogStore) Read(ctx context.Context, limit int, offset int) ([]LogEntry, error) {
	return s.Query(ctx, LogQueryOptions{Limit: limit, Offset: offset})
}

// Query retrieves logs with filtering options.
func (s *SQLiteLogStore) Query(ctx context.Context, opts LogQueryOptions) ([]LogEntry, error) {
	baseQuery := "SELECT id, timestamp, level, source, message, metadata FROM logs"
	var args []interface{}
	var conditions []string

	if opts.Level != "" && opts.Level != "ALL" {
		conditions = append(conditions, "level = ?")
		args = append(args, opts.Level)
	}
	if opts.Source != "" && opts.Source != "ALL" {
		conditions = append(conditions, "source = ?")
		args = append(args, opts.Source)
	}
	if opts.Search != "" {
		conditions = append(conditions, "(message LIKE ? OR source LIKE ?)")
		searchPattern := "%" + opts.Search + "%"
		args = append(args, searchPattern, searchPattern)
	}

	if len(conditions) > 0 {
		baseQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	// We order by timestamp DESC to get the most recent logs first when using LIMIT
	baseQuery += " ORDER BY timestamp DESC"

	if opts.Limit > 0 {
		baseQuery += " LIMIT ?"
		args = append(args, opts.Limit)
	}
	if opts.Offset > 0 {
		baseQuery += " OFFSET ?"
		args = append(args, opts.Offset)
	}

	rows, err := s.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []LogEntry
	for rows.Next() {
		var l LogEntry
		var metaStr string
		// Scan nullable source as sql.NullString? LogEntry.Source is string.
		// If DB has NULL, Scan into *string might fail if not careful, but we insert empty string usually.
		// However, let's be safe.
		var source sql.NullString
		if err := rows.Scan(&l.ID, &l.Timestamp, &l.Level, &source, &l.Message, &metaStr); err != nil {
			return nil, err
		}
		if source.Valid {
			l.Source = source.String
		}

		if metaStr != "" {
			if err := json.Unmarshal([]byte(metaStr), &l.Metadata); err != nil {
				// We ignore metadata parse errors to return partial data
				// but in a real app we might want to log this debug warning
			}
		}
		logs = append(logs, l)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Since we queried DESC (Newest First), we should reverse the slice to return Chronological order (Oldest First)
	// so the UI can append them naturally.
	for i, j := 0, len(logs)-1; i < j; i, j = i+1, j-1 {
		logs[i], logs[j] = logs[j], logs[i]
	}

	return logs, nil
}
