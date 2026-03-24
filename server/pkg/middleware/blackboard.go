// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// BlackboardStore represents a shared key-value store with agent-aware row-level security.
//
// Summary: Represents an in-memory or persisted shared blackboard storage component.
type BlackboardStore struct {
	db *sql.DB
}

// NewBlackboardStore creates a new SQLite Blackboard store.
//
// Summary: Initializes and sets up a new blackboard SQLite-backed data store at the given path.
//
// Parameters:
//   - path (string): The filesystem path where the SQLite database will be stored.
//
// Returns:
//   - *BlackboardStore: The created blackboard store instance.
//   - error: An error if database connection or initialization fails.
//
// Errors:
//   - Returns an error if the path is empty.
//   - Returns an error if SQLite fails to open the database.
//   - Returns an error if the table creation fails.
func NewBlackboardStore(path string) (*BlackboardStore, error) {
	if path == "" {
		return nil, fmt.Errorf("sqlite path is required")
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS blackboard (
		agent_id TEXT,
		key TEXT,
		value TEXT,
		PRIMARY KEY(agent_id, key)
	);
	CREATE INDEX IF NOT EXISTS idx_agent_id ON blackboard(agent_id);
	`
	ctxSchema, cancelSchema := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelSchema()

	if _, err := db.ExecContext(ctxSchema, schema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to create blackboard table: %w", err)
	}

	return &BlackboardStore{db: db}, nil
}

// Get retrieves a value from the blackboard for a specific agent.
//
// Summary: Fetches the value for a specific agent and key from the blackboard store.
//
// Parameters:
//   - ctx (context.Context): The context for database operations.
//   - agentID (string): The identifier of the agent trying to access the value.
//   - key (string): The key associated with the stored value.
//
// Returns:
//   - string: The fetched value associated with the key.
//   - error: An error if the key doesn't exist or a database operation fails.
//
// Errors:
//   - Returns sql.ErrNoRows if the key doesn't exist for the agent.
//   - Returns an error if the underlying database query fails.
func (s *BlackboardStore) Get(ctx context.Context, agentID, key string) (string, error) {
	var value string
	err := s.db.QueryRowContext(ctx, "SELECT value FROM blackboard WHERE agent_id = ? AND key = ?", agentID, key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("key not found")
		}
		return "", err
	}
	return value, nil
}

// Set stores a value in the blackboard for a specific agent.
//
// Summary: Inserts or updates a value for a specific agent and key in the blackboard store.
//
// Parameters:
//   - ctx (context.Context): The context for database operations.
//   - agentID (string): The identifier of the agent storing the value.
//   - key (string): The key to associate with the value.
//   - value (string): The value to store.
//
// Returns:
//   - error: An error if the database operation fails, nil on success.
//
// Errors:
//   - Returns an error if the underlying database execution fails.
func (s *BlackboardStore) Set(ctx context.Context, agentID, key, value string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO blackboard (agent_id, key, value) VALUES (?, ?, ?)
		ON CONFLICT(agent_id, key) DO UPDATE SET value = excluded.value
	`, agentID, key, value)
	return err
}

// Close closes the database connection.
//
// Summary: Closes the underlying database connection for the blackboard store.
//
// Returns:
//   - error: An error if closing the database fails.
//
// Errors:
//   - Returns an error if the underlying sqlite database could not be closed.
func (s *BlackboardStore) Close() error {
	return s.db.Close()
}
