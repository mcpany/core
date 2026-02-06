// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite" // Register sqlite driver
)

// SQLiteLogStore implements LogStore using SQLite.
type SQLiteLogStore struct {
	db *sql.DB
}

// NewSQLiteLogStore creates a new SQLite-backed log store.
// If the path ends with .db, it appends _logs.db to the filename to avoid conflict with main config DB.
// e.g. data/mcpany.db -> data/mcpany_logs.db
func NewSQLiteLogStore(basePath string) (*SQLiteLogStore, error) {
	var logPath string
	ext := filepath.Ext(basePath)
	if ext != "" {
		logPath = strings.TrimSuffix(basePath, ext) + "_logs" + ext
	} else {
		logPath = basePath + "_logs.db"
	}

	if err := os.MkdirAll(filepath.Dir(logPath), 0750); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	db, err := sql.Open("sqlite", logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite db: %w", err)
	}

	if err := initLogSchema(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to init log schema: %w", err)
	}

	ctx := context.Background()
	// Set pragmas for performance
	if _, err := db.ExecContext(ctx, "PRAGMA journal_mode=WAL;"); err != nil {
		return nil, fmt.Errorf("failed to set WAL mode: %w", err)
	}
	if _, err := db.ExecContext(ctx, "PRAGMA synchronous=NORMAL;"); err != nil {
		return nil, fmt.Errorf("failed to set synchronous mode: %w", err)
	}
	// Busy timeout to handle concurrency
	if _, err := db.ExecContext(ctx, "PRAGMA busy_timeout=5000;"); err != nil {
		return nil, fmt.Errorf("failed to set busy_timeout: %w", err)
	}

	return &SQLiteLogStore{db: db}, nil
}

func initLogSchema(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS system_logs (
		id TEXT PRIMARY KEY,
		timestamp DATETIME NOT NULL,
		level TEXT NOT NULL,
		source TEXT,
		message TEXT,
		metadata TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_system_logs_timestamp ON system_logs(timestamp DESC);
	`
	_, err := db.ExecContext(context.Background(), query)
	return err
}

// Write persists a log entry.
func (s *SQLiteLogStore) Write(entry LogEntry) error {
	metadataBytes, err := json.Marshal(entry.Metadata)
	if err != nil {
		metadataBytes = []byte("{}")
	}

	// Ensure timestamp is parsed correctly or use now
	var ts time.Time
	if entry.Timestamp != "" {
		// Try parsing RFC3339
		parsed, err := time.Parse(time.RFC3339, entry.Timestamp)
		if err == nil {
			ts = parsed
		} else {
			// Try parsing RFC3339Nano
			parsed, err := time.Parse(time.RFC3339Nano, entry.Timestamp)
			if err == nil {
				ts = parsed
			} else {
				ts = time.Now()
			}
		}
	} else {
		ts = time.Now()
	}
	ts = ts.UTC()

	_, err = s.db.ExecContext(context.Background(), "INSERT INTO system_logs (id, timestamp, level, source, message, metadata) VALUES (?, ?, ?, ?, ?, ?)",
		entry.ID, ts, entry.Level, entry.Source, entry.Message, string(metadataBytes))
	return err
}

// Read retrieves the last N log entries.
func (s *SQLiteLogStore) Read(limit int) ([]LogEntry, error) {
	rows, err := s.db.QueryContext(context.Background(), "SELECT id, timestamp, level, source, message, metadata FROM system_logs ORDER BY timestamp DESC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var entries []LogEntry
	// Use a slice to store results in reverse order (oldest first) so they look correct in history
	// But query is DESC (newest first). So we need to reverse them.
	var reversedEntries []LogEntry

	for rows.Next() {
		var entry LogEntry
		var ts time.Time
		var metadataStr string
		var source sql.NullString

		if err := rows.Scan(&entry.ID, &ts, &entry.Level, &source, &entry.Message, &metadataStr); err != nil {
			return nil, err
		}

		if source.Valid {
			entry.Source = source.String
		}

		entry.Timestamp = ts.UTC().Format(time.RFC3339)
		if len(metadataStr) > 0 {
			var meta map[string]any
			if err := json.Unmarshal([]byte(metadataStr), &meta); err == nil {
				entry.Metadata = meta
			}
		}

		reversedEntries = append(reversedEntries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Reverse to chronological order (oldest -> newest)
	for i := len(reversedEntries) - 1; i >= 0; i-- {
		entries = append(entries, reversedEntries[i])
	}

	return entries, nil
}

// Close closes the store.
func (s *SQLiteLogStore) Close() error {
	return s.db.Close()
}
