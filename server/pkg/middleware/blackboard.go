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
// Summary: A store for agent-specific key-value data.
type BlackboardStore struct {
	db *sql.DB
}

// NewBlackboardStore creates a new SQLite Blackboard store.
//
// Summary: Initializes a BlackboardStore using SQLite.
//
// Parameters:
//   - path: string. The file path for the SQLite database.
//
// Returns:
//   - *BlackboardStore: The initialized blackboard store.
//   - error: An error if database opening or schema creation fails.
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
// Summary: Retrieves a specific key's value for a given agent.
//
// Parameters:
//   - ctx: context.Context. The context for the database query.
//   - agentID: string. The unique identifier of the agent.
//   - key: string. The key to retrieve.
//
// Returns:
//   - string: The retrieved value.
//   - error: An error if the key is not found or the query fails.
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
// Summary: Upserts a key-value pair for a given agent.
//
// Parameters:
//   - ctx: context.Context. The context for the database execution.
//   - agentID: string. The unique identifier of the agent.
//   - key: string. The key to set.
//   - value: string. The value to store.
//
// Returns:
//   - error: An error if the database operation fails.
func (s *BlackboardStore) Set(ctx context.Context, agentID, key, value string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO blackboard (agent_id, key, value) VALUES (?, ?, ?)
		ON CONFLICT(agent_id, key) DO UPDATE SET value = excluded.value
	`, agentID, key, value)
	return err
}

// Close closes the database connection.
//
// Summary: Closes the underlying SQLite database connection.
//
// Returns:
//   - error: An error if the closing operation fails.
func (s *BlackboardStore) Close() error {
	return s.db.Close()
}
