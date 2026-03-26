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
// Summary: Audit store implementation that persists log entries in a PostgreSQL database with cryptographic hash chaining.
// PostgresAuditStore writes audit logs to a PostgreSQL database.
// Summary: Audit store implementation that persists log entries in a PostgreSQL database with cryptographic hash chaining.
	db *sql.DB
	mu sync.Mutex
}

// NewPostgresAuditStore creates a new PostgresAuditStore.
// Summary: Initializes a new PostgresAuditStore with the provided DSN and creates the schema if it does not exist.
// Parameters:
//   - dsn (string): The PostgreSQL connection string (DSN).
//
// Returns:
//   - *PostgresAuditStore: The initialized audit store.
//   - error: An error if the database connection or schema initialization fails.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
// NewPostgresAuditStore creates a new PostgresAuditStore.
// Summary: Initializes a new PostgresAuditStore with the provided DSN and creates the schema if it does not exist.
// Parameters:
//   - dsn (string): The PostgreSQL connection string (DSN).
//
// Returns:
//   - *PostgresAuditStore: The initialized audit store.
//   - error: An error if the database connection or schema initialization fails.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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
// Summary: Persists a single audit entry with cryptographic hash chaining, ensuring strict sequential consistency via table locking.
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - entry (Entry): The audit log entry to be persisted.
//
// Returns:
//   - error: An error if the transaction fails, hash calculation fails, or database write fails.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
// Write writes an audit entry to the database.
// Summary: Persists a single audit entry with cryptographic hash chaining, ensuring strict sequential consistency via table locking.
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - entry (Entry): The audit log entry to be persisted.
//
// Returns:
//   - error: An error if the transaction fails, hash calculation fails, or database write fails.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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

	// Lock the table to ensure strict sequential consistency for hash chaining.
	if _, err := tx.ExecContext(ctx, "LOCK TABLE audit_logs IN EXCLUSIVE MODE"); err != nil {
		return fmt.Errorf("failed to lock table: %w", err)
	}

	var prevHash string
	err = tx.QueryRowContext(ctx, "SELECT hash FROM audit_logs ORDER BY id DESC LIMIT 1").Scan(&prevHash)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to get previous hash: %w", err)
	}
	if err == sql.ErrNoRows {
		prevHash = "" // First entry
	}

	// Compute hash
	tsStr := entry.Timestamp.Format(time.RFC3339Nano)
	hash := computeHash(tsStr, entry.ToolName, entry.UserID, entry.ProfileID, argsJSON, resultJSON, entry.Error, entry.DurationMs, prevHash)

	query := `
	INSERT INTO audit_logs (
		timestamp, tool_name, user_id, profile_id, arguments, result, error, duration_ms, prev_hash, hash
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err = tx.ExecContext(ctx, query,
		entry.Timestamp,
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

// Read reads audit entries from the database based on the filter.
// Summary: Not implemented for PostgresAuditStore.
// Parameters:
//   - ctx (context.Context): The context for the request (unused).
//   - filter (Filter): Criteria for filtering audit logs (unused).
//
// Returns:
//   - []Entry: Always returns nil.
//   - error: Always returns a "not implemented" error.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
// Read reads audit entries from the database based on the filter.
// Summary: Not implemented for PostgresAuditStore.
// Parameters:
//   - ctx (context.Context): The context for the request (unused).
//   - filter (Filter): Criteria for filtering audit logs (unused).
//
// Returns:
//   - []Entry: Always returns nil.
//   - error: Always returns a "not implemented" error.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
	return nil, fmt.Errorf("read not implemented for postgres audit store")
}

// Verify checks the integrity of the audit logs.
// Summary: Validates the cryptographic hash chain of all entries in the audit_logs table.
// Parameters:
//   - None.
//
// Returns:
//   - bool: True if the entire audit log chain is valid.
//   - error: An error if a hash mismatch is detected or a database query fails.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
// Verify checks the integrity of the audit logs.
// Summary: Validates the cryptographic hash chain of all entries in the audit_logs table.
// Parameters:
//   - None.
//
// Returns:
//   - bool: True if the entire audit log chain is valid.
//   - error: An error if a hash mismatch is detected or a database query fails.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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
		var args, result sql.NullString
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
// Summary: Gracefully closes the connection to the PostgreSQL database.
// Returns:
//   - error: An error if the database connection fails to close.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
// Close closes the database connection.
// Summary: Gracefully closes the connection to the PostgreSQL database.
// Returns:
//   - error: An error if the database connection fails to close.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
// Parameters:
//   - None.
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.Close()
}
