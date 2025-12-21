// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// SQLiteAuditStore writes audit logs to a SQLite database.
type SQLiteAuditStore struct {
	db *sql.DB
	mu sync.Mutex
}

// NewSQLiteAuditStore creates a new SQLiteAuditStore.
func NewSQLiteAuditStore(path string) (*SQLiteAuditStore, error) {
	if path == "" {
		return nil, fmt.Errorf("sqlite path is required")
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Create table if not exists
	schema := `
	CREATE TABLE IF NOT EXISTS audit_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp TEXT,
		tool_name TEXT,
		user_id TEXT,
		profile_id TEXT,
		arguments TEXT,
		result TEXT,
		error TEXT,
		duration_ms INTEGER
	);
	`
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create audit_logs table: %w", err)
	}

	return &SQLiteAuditStore{
		db: db,
	}, nil
}

// Write writes an audit entry to the database.
func (s *SQLiteAuditStore) Write(entry AuditEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Marshal complex types
	argsJSON := "{}"
	if len(entry.Arguments) > 0 {
		argsJSON = string(entry.Arguments)
	}

	resultJSON := "{}"
	if entry.Result != nil {
		if b, err := json.Marshal(entry.Result); err == nil {
			resultJSON = string(b)
		}
	}

	query := `
	INSERT INTO audit_logs (
		timestamp, tool_name, user_id, profile_id, arguments, result, error, duration_ms
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		entry.Timestamp.Format(time.RFC3339Nano),
		entry.ToolName,
		entry.UserID,
		entry.ProfileID,
		argsJSON,
		resultJSON,
		entry.Error,
		entry.DurationMs,
	)
	return err
}

// Close closes the database connection.
func (s *SQLiteAuditStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.Close()
}
