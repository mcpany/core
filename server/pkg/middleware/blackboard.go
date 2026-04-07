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

type contextKey string

const (
	AgentAwareKey  contextKey = "agent_aware"
	IntentScopeKey contextKey = "intent_scope"
)

// BlackboardStore represents a shared key-value store with agent-aware row-level security.
//
// Summary: Represents a shared key-value store with agent-aware row-level security.
type BlackboardStore struct {
	db *sql.DB
}

// NewBlackboardStore creates a new SQLite Blackboard store.
//
// Summary: Creates a new SQLite-backed Blackboard store.
//
// Parameters:
//   - path (string): The file path to the SQLite database.
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
func NewBlackboardStore(path string) (*BlackboardStore, error) {
	if path == "" {
		return nil, fmt.Errorf("sqlite path is required")
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	ctxSchema, cancelSchema := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelSchema()

	var tableExists int
	err = db.QueryRowContext(ctxSchema, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='blackboard'").Scan(&tableExists)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to check sqlite_master: %w", err)
	}

	if tableExists > 0 {
		var hasIntentScope int
		err = db.QueryRowContext(ctxSchema, "SELECT COUNT(*) FROM pragma_table_info('blackboard') WHERE name='intent_scope'").Scan(&hasIntentScope)
		if err == nil && hasIntentScope == 0 {
			migration := `
			BEGIN TRANSACTION;
			CREATE TABLE blackboard_new (
				agent_id TEXT,
				intent_scope TEXT,
				key TEXT,
				value TEXT,
				PRIMARY KEY(agent_id, intent_scope, key)
			);
			INSERT INTO blackboard_new (agent_id, intent_scope, key, value)
			SELECT agent_id, '', key, value FROM blackboard;
			DROP TABLE blackboard;
			ALTER TABLE blackboard_new RENAME TO blackboard;
			COMMIT;
			`
			if _, err := db.ExecContext(ctxSchema, migration); err != nil {
				db.ExecContext(ctxSchema, "ROLLBACK;")
				_ = db.Close()
				return nil, fmt.Errorf("failed to migrate blackboard table: %w", err)
			}
		}
	}

	schema := `
	CREATE TABLE IF NOT EXISTS blackboard (
		agent_id TEXT,
		intent_scope TEXT,
		key TEXT,
		value TEXT,
		PRIMARY KEY(agent_id, intent_scope, key)
	);
	CREATE INDEX IF NOT EXISTS idx_agent_id ON blackboard(agent_id);
	CREATE INDEX IF NOT EXISTS idx_intent_scope ON blackboard(intent_scope);
	`
	if _, err := db.ExecContext(ctxSchema, schema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to create blackboard table: %w", err)
	}

	return &BlackboardStore{db: db}, nil
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
	agentAware := false
	if a, ok := ctx.Value(AgentAwareKey).(bool); ok {
		agentAware = a
	}

	intentScope := ""
	if agentAware {
		if sc, ok := ctx.Value(IntentScopeKey).(string); ok {
			intentScope = sc
		}
	}

	var value string
	err := s.db.QueryRowContext(ctx, "SELECT value FROM blackboard WHERE agent_id = ? AND intent_scope = ? AND key = ?", agentID, intentScope, key).Scan(&value)

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
	agentAware := false
	if a, ok := ctx.Value(AgentAwareKey).(bool); ok {
		agentAware = a
	}

	intentScope := ""
	if agentAware {
		if sc, ok := ctx.Value(IntentScopeKey).(string); ok {
			intentScope = sc
		}
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO blackboard (agent_id, intent_scope, key, value) VALUES (?, ?, ?, ?)
		ON CONFLICT(agent_id, intent_scope, key) DO UPDATE SET value = excluded.value
	`, agentID, intentScope, key, value)
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
