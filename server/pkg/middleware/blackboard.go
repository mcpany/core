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

// BlackboardStore represents the public BlackboardStore entity.
//
// Summary: Defines the structured data model representing a store.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type BlackboardStore struct {
	db *sql.DB
}

// NewBlackboardStore serves as a public interface for interacting with NewBlackboardStore.
//
// Summary: Constructs and returns an initialized blackboard store ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// Get serves as a public interface for interacting with Get.
//
// Summary: Fetches and returns the underlying  from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// Set serves as a public interface for interacting with Set.
//
// Summary: Set the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *BlackboardStore) Set(ctx context.Context, agentID, key, value string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO blackboard (agent_id, key, value) VALUES (?, ?, ?)
		ON CONFLICT(agent_id, key) DO UPDATE SET value = excluded.value
	`, agentID, key, value)
	return err
}

// Close serves as a public interface for interacting with Close.
//
// Summary: Close the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *BlackboardStore) Close() error {
	return s.db.Close()
}
