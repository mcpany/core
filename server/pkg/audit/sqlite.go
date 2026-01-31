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
type SQLiteAuditStore struct {
	db *sql.DB
	mu sync.Mutex

	queue     chan Entry
	wg        sync.WaitGroup
	done      chan struct{}
	closeOnce sync.Once
}

const (
	sqliteBatchSize = 100
	sqliteBatchWait = 100 * time.Millisecond
)

// NewSQLiteAuditStore creates a new SQLiteAuditStore.
//
// path is the path.
//
// Returns the result.
// Returns an error if the operation fails.
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

	store := &SQLiteAuditStore{
		db:    db,
		queue: make(chan Entry, 1000), // Buffer size 1000
		done:  make(chan struct{}),
	}
	store.wg.Add(1)
	go store.worker()

	return store, nil
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

func (s *SQLiteAuditStore) worker() {
	defer s.wg.Done()
	var batch []Entry
	ticker := time.NewTicker(sqliteBatchWait)
	defer ticker.Stop()

	for {
		select {
		case entry, ok := <-s.queue:
			if !ok {
				s.flushBatch(batch)
				return
			}
			batch = append(batch, entry)
			if len(batch) >= sqliteBatchSize {
				s.flushBatch(batch)
				batch = nil
			}
		case <-ticker.C:
			if len(batch) > 0 {
				s.flushBatch(batch)
				batch = nil
			}
		case <-s.done:
			// Drain queue
			for {
				select {
				case entry := <-s.queue:
					batch = append(batch, entry)
					if len(batch) >= sqliteBatchSize {
						s.flushBatch(batch)
						batch = nil
					}
				default:
					s.flushBatch(batch)
					return
				}
			}
		}
	}
}

func (s *SQLiteAuditStore) flushBatch(batch []Entry) {
	if len(batch) == 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		fmt.Printf("Failed to begin transaction: %v\n", err)
		return
	}
	defer func() { _ = tx.Rollback() }()

	// Get previous hash
	var prevHash string
	err = tx.QueryRowContext(ctx, "SELECT hash FROM audit_logs ORDER BY id DESC LIMIT 1").Scan(&prevHash)
	if err != nil && err != sql.ErrNoRows {
		fmt.Printf("Failed to get previous hash: %v\n", err)
		return
	}
	if err == sql.ErrNoRows {
		prevHash = ""
	}

	stmt, err := tx.PrepareContext(ctx, `
	INSERT INTO audit_logs (
		timestamp, tool_name, user_id, profile_id, arguments, result, error, duration_ms, prev_hash, hash
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		fmt.Printf("Failed to prepare statement: %v\n", err)
		return
	}
	defer func() { _ = stmt.Close() }()

	for _, entry := range batch {
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

		// Compute hash
		hash := computeHash(ts, entry.ToolName, entry.UserID, entry.ProfileID, argsJSON, resultJSON, entry.Error, entry.DurationMs, prevHash)

		_, err = stmt.ExecContext(ctx,
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
		if err != nil {
			fmt.Printf("Failed to insert entry: %v\n", err)
			return
		}
		prevHash = hash
	}

	if err := tx.Commit(); err != nil {
		fmt.Printf("Failed to commit transaction: %v\n", err)
	}
}

// Write writes an audit entry to the database.
//
// ctx is the context for the request.
// entry is the entry.
//
// Returns an error if the operation fails.
func (s *SQLiteAuditStore) Write(_ context.Context, entry Entry) error {
	select {
	case s.queue <- entry:
		return nil
	default:
		// ⚡ BOLT: Dropping log if queue is full to prevent blocking main thread.
		// Randomized Selection from Top 5 High-Impact Targets
		return fmt.Errorf("audit queue full")
	}
}

// Read reads audit entries from the database based on the filter.
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
// It returns true if the chain is valid, false otherwise.
// If an error occurs during reading, it returns false and the error.
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
// Returns an error if the operation fails.
func (s *SQLiteAuditStore) Close() error {
	var err error
	s.closeOnce.Do(func() {
		// Signal worker to stop
		close(s.done)
		// Do not close queue to avoid panics in concurrent Writes.
		// The worker will drain the queue via the done signal loop.

		// Wait for worker to finish flushing
		s.wg.Wait()

		s.mu.Lock()
		defer s.mu.Unlock()
		err = s.db.Close()
	})
	return err
}
