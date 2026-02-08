// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/mcpany/core/server/pkg/validation"

	// modernc.org/sqlite is a pure Go SQLite driver.
	_ "modernc.org/sqlite"
)

// SQLiteAuditStore writes audit logs to a SQLite database.
//
// Summary: writes audit logs to a SQLite database.
type SQLiteAuditStore struct {
	db *sql.DB
	mu sync.Mutex
}

// NewSQLiteAuditStore creates a new SQLiteAuditStore.
//
// Summary: creates a new SQLiteAuditStore.
//
// Parameters:
//   - path: string. The path.
//
// Returns:
//   - *SQLiteAuditStore: The *SQLiteAuditStore.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func NewSQLiteAuditStore(path string) (*SQLiteAuditStore, error) {
	if path == "" {
		return nil, fmt.Errorf("sqlite path is required")
	}

	if err := validation.IsAllowedPath(path); err != nil {
		return nil, fmt.Errorf("sqlite audit path not allowed: %w", err)
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
	ctxSchema, cancelSchema := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelSchema()
	if _, err := db.ExecContext(ctxSchema, schema); err != nil {
		_ = db.Close() // Best effort close
		return nil, fmt.Errorf("failed to create audit_logs table: %w", err)
	}

	// Set pragmas for performance
	ctxPragma, cancelPragma := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelPragma()
	if _, err := db.ExecContext(ctxPragma, "PRAGMA journal_mode=WAL;"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to set WAL mode: %w", err)
	}
	if _, err := db.ExecContext(ctxPragma, "PRAGMA synchronous=NORMAL;"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to set synchronous mode: %w", err)
	}
	if _, err := db.ExecContext(ctxPragma, "PRAGMA busy_timeout=5000;"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to set busy_timeout: %w", err)
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
	if err := ensureColumn(db, "prev_hash"); err != nil {
		return err
	}
	if err := ensureColumn(db, "hash"); err != nil {
		return err
	}
	return nil
}

func ensureColumn(db *sql.DB, colName string) error {
	// Whitelist valid column names to prevent SQL injection even from internal calls
	switch colName {
	case "prev_hash", "hash":
		// Allowed
	default:
		return fmt.Errorf("invalid column name: %s", colName)
	}

	// Check if column exists
	//nolint:gosec // colName is validated above
	query := fmt.Sprintf("SELECT %s FROM audit_logs LIMIT 1", colName)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := db.ExecContext(ctx, query); err == nil {
		return nil
	}
	// Add column

	query = fmt.Sprintf("ALTER TABLE audit_logs ADD COLUMN %s TEXT DEFAULT ''", colName)
	ctxAlter, cancelAlter := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelAlter()
	_, err := db.ExecContext(ctxAlter, query)
	return err
}

// Write writes an audit entry to the database.
//
// Summary: writes an audit entry to the database.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - entry: Entry. The entry.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (s *SQLiteAuditStore) Write(ctx context.Context, entry Entry) error {
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
	err := s.db.QueryRowContext(ctx, "SELECT hash FROM audit_logs ORDER BY id DESC LIMIT 1").Scan(&prevHash)
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

	_, err = s.db.ExecContext(ctx, query,
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

// Read reads audit entries from the database based on the filter.
//
// Summary: reads audit entries from the database based on the filter.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - filter: Filter. The filter.
//
// Returns:
//   - []Entry: The []Entry.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (s *SQLiteAuditStore) Read(ctx context.Context, filter Filter) ([]Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := "SELECT timestamp, tool_name, user_id, profile_id, arguments, result, error, duration_ms FROM audit_logs WHERE 1=1"
	var args []any

	if filter.StartTime != nil {
		query += " AND timestamp >= ?"
		args = append(args, filter.StartTime.Format(time.RFC3339Nano))
	}
	if filter.EndTime != nil {
		query += " AND timestamp <= ?"
		args = append(args, filter.EndTime.Format(time.RFC3339Nano))
	}
	if filter.ToolName != "" {
		query += " AND tool_name = ?"
		args = append(args, filter.ToolName)
	}
	if filter.UserID != "" {
		query += " AND user_id = ?"
		args = append(args, filter.UserID)
	}
	if filter.ProfileID != "" {
		query += " AND profile_id = ?"
		args = append(args, filter.ProfileID)
	}

	query += " ORDER BY timestamp DESC"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}
	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var entries []Entry
	for rows.Next() {
		var entry Entry
		var tsStr, argsStr, resultStr string
		if err := rows.Scan(&tsStr, &entry.ToolName, &entry.UserID, &entry.ProfileID, &argsStr, &resultStr, &entry.Error, &entry.DurationMs); err != nil {
			return nil, err
		}

		entry.Timestamp, _ = time.Parse(time.RFC3339Nano, tsStr)
		entry.Arguments = json.RawMessage(argsStr)
		if resultStr != "" && resultStr != "{}" {
			_ = json.Unmarshal([]byte(resultStr), &entry.Result)
		}
		entry.Duration = fmt.Sprintf("%dms", entry.DurationMs)

		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

// Verify checks the integrity of the audit logs.
//
// Summary: checks the integrity of the audit logs.
//
// Parameters:
//   None.
//
// Returns:
//   - bool: The bool.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (s *SQLiteAuditStore) Verify() (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	rows, err := s.db.QueryContext(ctx, "SELECT id, timestamp, tool_name, user_id, profile_id, arguments, result, error, duration_ms, prev_hash, hash FROM audit_logs ORDER BY id ASC")
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

		// Check hash version
		var calculatedHash string
		if len(hash) > 3 && hash[:3] == "v1:" {
			calculatedHash = computeHash(ts, toolName, userID, profileID, args, result, errorMsg, durationMs, prevHash)
		} else {
			// Fallback to legacy
			calculatedHash = computeHashV0(ts, toolName, userID, profileID, args, result, errorMsg, durationMs, prevHash)
		}

		if calculatedHash != hash {
			return false, fmt.Errorf("integrity violation at id %d: hash mismatch (calculated %q, got %q)", id, calculatedHash, hash)
		}

		expectedPrevHash = hash
	}
	if err := rows.Err(); err != nil {
		return false, err
	}
	return true, nil
}

// Close closes the database connection.
//
// Summary: closes the database connection.
//
// Parameters:
//   None.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (s *SQLiteAuditStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.Close()
}
