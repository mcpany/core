// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite" // Ensure driver is loaded
)

// LogStore defines the interface for storing and retrieving logs.
type LogStore interface {
	Write(ctx context.Context, entry LogEntry) error
	Read(ctx context.Context, opts LogQueryOptions) ([]LogEntry, error)
}

// LogQueryOptions defines filters for querying logs.
type LogQueryOptions struct {
	Limit  int
	Offset int
	Level  string
	Source string
	Search string
	// Since/Until could be added
}

// SQLiteLogStore implements LogStore using SQLite.
type SQLiteLogStore struct {
	db *sql.DB
}

// NewSQLiteLogStore creates a new SQLiteLogStore.
// It creates the logs table if it doesn't exist.
func NewSQLiteLogStore(db *sql.DB) (*SQLiteLogStore, error) {
	s := &SQLiteLogStore{db: db}
	if err := s.initSchema(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *SQLiteLogStore) initSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS logs (
		id TEXT PRIMARY KEY,
		timestamp DATETIME NOT NULL,
		level TEXT NOT NULL,
		source TEXT,
		message TEXT,
		metadata TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp);
	CREATE INDEX IF NOT EXISTS idx_logs_level ON logs(level);
	CREATE INDEX IF NOT EXISTS idx_logs_source ON logs(source);
	`
	_, err := s.db.Exec(query)
	return err
}

// Write inserts a log entry into the database.
func (s *SQLiteLogStore) Write(ctx context.Context, entry LogEntry) error {
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}
	if entry.Timestamp == "" {
		entry.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}

	metadataJSON, err := json.Marshal(entry.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
	INSERT INTO logs (id, timestamp, level, source, message, metadata)
	VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err = s.db.ExecContext(ctx, query, entry.ID, entry.Timestamp, entry.Level, entry.Source, entry.Message, string(metadataJSON))
	return err
}

// Read queries logs from the database.
func (s *SQLiteLogStore) Read(ctx context.Context, opts LogQueryOptions) ([]LogEntry, error) {
	query := "SELECT id, timestamp, level, source, message, metadata FROM logs WHERE 1=1"
	var args []interface{}

	if opts.Level != "" && opts.Level != "ALL" {
		query += " AND level = ?"
		args = append(args, opts.Level)
	}
	if opts.Source != "" && opts.Source != "ALL" {
		query += " AND source = ?"
		args = append(args, opts.Source)
	}
	if opts.Search != "" {
		query += " AND (message LIKE ? OR source LIKE ?)"
		pattern := "%" + opts.Search + "%"
		args = append(args, pattern, pattern)
	}

	query += " ORDER BY timestamp DESC"

	if opts.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, opts.Limit)
	}
	if opts.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, opts.Offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []LogEntry
	for rows.Next() {
		var entry LogEntry
		var metadataStr string
		if err := rows.Scan(&entry.ID, &entry.Timestamp, &entry.Level, &entry.Source, &entry.Message, &metadataStr); err != nil {
			return nil, err
		}
		if metadataStr != "" {
			_ = json.Unmarshal([]byte(metadataStr), &entry.Metadata)
		}
		logs = append(logs, entry)
	}

	return logs, nil
}
