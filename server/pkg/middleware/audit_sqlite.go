// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
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
// path is the path.
// Returns the result, an error.
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
		duration_ms INTEGER,
		prev_hash TEXT,
		hash TEXT
	);
	`
	if _, err := db.Exec(schema); err != nil {
		_ = db.Close() // Best effort close
		return nil, fmt.Errorf("failed to create audit_logs table: %w", err)
	}

	// Ensure columns exist (for migration)
	if err := ensureColumns(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ensure columns: %w", err)
	}

	return &SQLiteAuditStore{
		db: db,
	}, nil
}

func ensureColumns(db *sql.DB) error {
	// Helper to check and add column
	addColumn := func(colName string) error {
		// Check if column exists
		//nolint:gosec // colName is trusted input from this function
		query := fmt.Sprintf("SELECT %s FROM audit_logs LIMIT 1", colName)
		if _, err := db.Exec(query); err == nil {
			return nil
		}
		// Add column

		query = fmt.Sprintf("ALTER TABLE audit_logs ADD COLUMN %s TEXT DEFAULT ''", colName)
		_, err := db.Exec(query)
		return err
	}

	if err := addColumn("prev_hash"); err != nil {
		return err
	}
	if err := addColumn("hash"); err != nil {
		return err
	}
	return nil
}

// Write writes an audit entry to the database.
// Returns an error.
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

	ts := entry.Timestamp.Format(time.RFC3339Nano)

	// Get previous hash
	var prevHash string
	// Order by ID desc to get the last entry
	err := s.db.QueryRow("SELECT hash FROM audit_logs ORDER BY id DESC LIMIT 1").Scan(&prevHash)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to get previous hash: %w", err)
	}
	if err == sql.ErrNoRows {
		prevHash = "" // First entry
	}

	// Compute hash
	hash := computeHash(ts, entry.ToolName, entry.UserID, entry.ProfileID, argsJSON, resultJSON, entry.Error, entry.DurationMs, prevHash)

	query := `
	INSERT INTO audit_logs (
		timestamp, tool_name, user_id, profile_id, arguments, result, error, duration_ms, prev_hash, hash
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.Exec(query,
		ts,
		entry.ToolName,
		entry.UserID,
		entry.ProfileID,
		argsJSON,
		resultJSON,
		entry.Error,
		entry.DurationMs,
		prevHash,
		hash,
	)
	return err
}

func computeHash(timestamp, toolName, userID, profileID, args, result, errorMsg string, durationMs int64, prevHash string) string {
	data := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%d|%s",
		timestamp, toolName, userID, profileID, args, result, errorMsg, durationMs, prevHash)
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:])
}

// Verify checks the integrity of the audit logs.
// It returns true if the chain is valid, false otherwise.
// If an error occurs during reading, it returns false and the error.
func (s *SQLiteAuditStore) Verify() (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.Query("SELECT id, timestamp, tool_name, user_id, profile_id, arguments, result, error, duration_ms, prev_hash, hash FROM audit_logs ORDER BY id ASC")
	if err != nil {
		return false, err
	}
	defer func() { _ = rows.Close() }()

	var expectedPrevHash string
	for rows.Next() {
		var id int64
		var ts, toolName, userID, profileID, args, result, errorMsg, prevHash, hash string
		var durationMs int64

		if err := rows.Scan(&id, &ts, &toolName, &userID, &profileID, &args, &result, &errorMsg, &durationMs, &prevHash, &hash); err != nil {
			return false, fmt.Errorf("scan error at id %d: %w", id, err)
		}

		if prevHash != expectedPrevHash {
			return false, fmt.Errorf("integrity violation at id %d: prev_hash mismatch (expected %q, got %q)", id, expectedPrevHash, prevHash)
		}

		calculatedHash := computeHash(ts, toolName, userID, profileID, args, result, errorMsg, durationMs, prevHash)
		if calculatedHash != hash {
			return false, fmt.Errorf("integrity violation at id %d: hash mismatch (calculated %q, got %q)", id, calculatedHash, hash)
		}

		expectedPrevHash = hash
	}
	return true, nil
}

// Close closes the database connection.
// Returns an error.
func (s *SQLiteAuditStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.Close()
}
