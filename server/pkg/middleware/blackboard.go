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
// Summary: Represents a shared key-value store with agent-aware row-level security.
type BlackboardStore struct {
	db             *sql.DB
	isolationLevel string
}

// NewBlackboardStore creates a new SQLite Blackboard store.
//
// Summary: Creates a new SQLite-backed Blackboard store.
//
// Parameters:
//   - path (string): The file path to the SQLite database.
//   - isolationLevel (string): The isolation level, e.g., "agent_aware".
//
// Returns:
//   - *BlackboardStore: A new instance of the BlackboardStore.
//   - error: An error if the database connection or schema creation fails.
//
// Errors:
//   - Returns an error if the path is empty.
//   - Returns an error if opening the sqlite database fails.
//   - Returns an error if creating the blackboard table fails.
//
// Side Effects:
//   - Connects to the specified SQLite database.
//   - Executes schema creation queries (CREATE TABLE, CREATE INDEX).
func NewBlackboardStore(path string, isolationLevel string) (*BlackboardStore, error) {
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

	return &BlackboardStore{
		db:             db,
		isolationLevel: isolationLevel,
	}, nil
}

// Get retrieves a value from the blackboard for a specific agent.
//
// Summary: Retrieves a value from the blackboard for a specific agent.
//
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - agentID (string): The identifier of the agent.
//   - key (string): The key to lookup.
//
// Returns:
//   - string: The retrieved value.
//   - error: An error if the value cannot be retrieved.
//
// Errors:
//   - Returns "key not found" if the key does not exist for the specified agent.
//   - Returns an error if the database query fails.
//
// Side Effects:
//   - Executes a SELECT query on the database.
func (s *BlackboardStore) Get(ctx context.Context, agentID, key string) (string, error) {
	if s.isolationLevel == "agent_aware" && agentID == "" {
		return "", fmt.Errorf("agent_aware isolation requires a valid agent ID")
	}

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
// Summary: Stores a value in the blackboard for a specific agent.
//
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - agentID (string): The identifier of the agent.
//   - key (string): The key to store the value under.
//   - value (string): The value to store.
//
// Returns:
//   - error: An error if the storage operation fails.
//
// Errors:
//   - Returns an error if the database execution fails.
//
// Side Effects:
//   - Executes an INSERT OR REPLACE (UPSERT) query on the database.
func (s *BlackboardStore) Set(ctx context.Context, agentID, key, value string) error {
	if s.isolationLevel == "agent_aware" && agentID == "" {
		return fmt.Errorf("agent_aware isolation requires a valid agent ID")
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO blackboard (agent_id, key, value) VALUES (?, ?, ?)
		ON CONFLICT(agent_id, key) DO UPDATE SET value = excluded.value
	`, agentID, key, value)
	return err
}

// Close closes the database connection.
//
// Summary: Closes the underlying database connection.
//
// Parameters:
//   - None.
//
// Returns:
//   - error: An error if closing the connection fails.
//
// Errors:
//   - Returns an error if the database connection cannot be closed properly.
//
// Side Effects:
//   - Closes the active database connection, preventing further queries.
func (s *BlackboardStore) Close() error {
	return s.db.Close()
}
