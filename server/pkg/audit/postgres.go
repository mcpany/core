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

	_ "github.com/lib/pq" // Register postgres driver
)

// PostgresAuditStore writes audit logs to a PostgreSQL database.
type PostgresAuditStore struct {
	db *sql.DB
	mu sync.Mutex
}

// NewPostgresAuditStore creates a new PostgresAuditStore.
//
// Parameters:
//   - dsn: string. The Data Source Name for the PostgreSQL connection.
//
// Returns:
//   - *PostgresAuditStore: The initialized audit store.
//   - error: An error if the connection fails or table creation fails.
func NewPostgresAuditStore(dsn string) (*PostgresAuditStore, error) {
	if dsn == "" {
		return nil, fmt.Errorf("postgres dsn is required")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres database: %w", err)
	}

	ctxPing, cancelPing := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelPing()
	if err := db.PingContext(ctxPing); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping postgres database: %w", err)
	}

	// Create table if not exists
	// Postgres uses standard SQL types.
	// BIGSERIAL for auto-incrementing 8-byte integer.
	// TIMESTAMPTZ for timestamps.
	// JSONB for arguments and result could be used, but we stick to TEXT/string to match the hashing logic exactly
	// and avoid issues with JSON normalization differences in DB.
	// We can cast to JSONB in queries if needed.
	schema := `
	CREATE TABLE IF NOT EXISTS audit_logs (
		id BIGSERIAL PRIMARY KEY,
		timestamp TIMESTAMPTZ NOT NULL,
		tool_name TEXT NOT NULL,
		user_id TEXT NOT NULL DEFAULT '',
		profile_id TEXT NOT NULL DEFAULT '',
		arguments TEXT,
		result TEXT,
		error TEXT,
		duration_ms BIGINT,
		prev_hash TEXT,
		hash TEXT
	);
	`
	ctxSchema, cancelSchema := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelSchema()
	if _, err := db.ExecContext(ctxSchema, schema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to create audit_logs table: %w", err)
	}

	return &PostgresAuditStore{
		db: db,
	}, nil
}

// Write writes an audit entry to the database.
//
// Parameters:
//   - ctx: context.Context. The context for the transaction.
//   - entry: Entry. The audit entry to write.
//
// Returns:
//   - error: An error if the write operation fails.
func (s *PostgresAuditStore) Write(ctx context.Context, entry Entry) error {
	// We don't need mutex here because we use database transaction for concurrency control.
	// s.mu.Lock() // removed

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

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback() // Safe to call even if committed
	}()

	// Get previous hash with locking to prevent concurrent writes branching the chain
	var prevHash string
	// Order by ID desc to get the last entry. FOR UPDATE locks the row(s).
	// Even if table is empty, we need to ensure we are the next one.
	// Since we can't lock a non-existent row, we might need a table lock or just rely on SERIALIZABLE if configured,
	// but explicit locking the "tip" is standard. If table is empty, there is no row to lock.
	// In that case, we might race on the very first insert.
	// To solve the "empty table race", we can use an advisory lock or lock the table.
	// Locking the table is heavy but safe for audit logs which are append-only.
	// Alternatively, we accept the race only on the very first record (id=1) which is rare.
	// Better approach: Lock the table in EXCLUSIVE mode which allows reads but blocks writes.
	// Or use `LOCK TABLE audit_logs IN SHARE ROW EXCLUSIVE MODE;`
	// Let's use `LOCK TABLE audit_logs IN EXCLUSIVE MODE` for safety to ensure strict ordering.
	if _, err := tx.ExecContext(ctx, "LOCK TABLE audit_logs IN EXCLUSIVE MODE"); err != nil {
		return fmt.Errorf("failed to lock table: %w", err)
	}

	err = tx.QueryRowContext(ctx, "SELECT hash FROM audit_logs ORDER BY id DESC LIMIT 1").Scan(&prevHash)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to get previous hash: %w", err)
	}
	if err == sql.ErrNoRows {
		prevHash = "" // First entry
	}

	// Compute hash
	// Use formatted timestamp string for hashing consistency (same as SQLite)
	tsStr := entry.Timestamp.Format(time.RFC3339Nano)
	hash := computeHash(tsStr, entry.ToolName, entry.UserID, entry.ProfileID, argsJSON, resultJSON, entry.Error, entry.DurationMs, prevHash)

	query := `
	INSERT INTO audit_logs (
		timestamp, tool_name, user_id, profile_id, arguments, result, error, duration_ms, prev_hash, hash
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err = tx.ExecContext(ctx, query,
		entry.Timestamp, // Postgres driver handles time.Time
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
		return fmt.Errorf("failed to insert audit log: %w", err)
	}

	return tx.Commit()
}

// Read reads audit entries from the database.
//
// Parameters:
//   - ctx: context.Context. Unused.
//   - filter: Filter. Unused.
//
// Returns:
//   - []Entry: Nil.
//   - error: An error indicating that read is not implemented for postgres audit store.
func (s *PostgresAuditStore) Read(_ context.Context, _ Filter) ([]Entry, error) {
	return nil, fmt.Errorf("read not implemented for postgres audit store")
}

// Verify checks the integrity of the audit logs by re-computing hashes.
//
// Returns:
//   - bool: True if all logs are valid and the chain is intact.
//   - error: An error if integrity verification fails or an operational error occurs.
func (s *PostgresAuditStore) Verify() (bool, error) {
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
		var ts time.Time
		var toolName, userID, profileID, errorMsg, prevHash, hash string
		var args, result sql.NullString // Can be null in some schemas, though we set defaults. Use NullString for safety.
		var durationMs int64

		if err := rows.Scan(&id, &ts, &toolName, &userID, &profileID, &args, &result, &errorMsg, &durationMs, &prevHash, &hash); err != nil {
			return false, fmt.Errorf("scan error at id %d: %w", id, err)
		}

		if prevHash != expectedPrevHash {
			return false, fmt.Errorf("integrity violation at id %d: prev_hash mismatch (expected %q, got %q)", id, expectedPrevHash, prevHash)
		}

		// Check hash version
		var calculatedHash string
		tsStr := ts.Format(time.RFC3339Nano)
		// Handle potential NULLs by treating as empty/default if that's how we wrote them,
		// but we wrote them as string literals "{}" or "".
		// If they come back as NULL from DB (legacy rows?), we treat as "".
		// Our Write method writes actual strings.
		argsStr := ""
		if args.Valid {
			argsStr = args.String
		}
		resultStr := ""
		if result.Valid {
			resultStr = result.String
		}

		if len(hash) > 3 && hash[:3] == "v1:" {
			calculatedHash = computeHash(tsStr, toolName, userID, profileID, argsStr, resultStr, errorMsg, durationMs, prevHash)
		} else {
			// Fallback to legacy
			calculatedHash = computeHashV0(tsStr, toolName, userID, profileID, argsStr, resultStr, errorMsg, durationMs, prevHash)
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
// Returns:
//   - error: An error if the connection cannot be closed.
func (s *PostgresAuditStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.Close()
}
