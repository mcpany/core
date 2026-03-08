// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package blackboard

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	_ "modernc.org/sqlite" // SQLite driver

	"github.com/mcpany/core/server/pkg/logging"
)

// BlackboardStore represents a persistent SQLite key-value store.
type BlackboardStore struct {
	db   *sql.DB
	path string
	mu   sync.RWMutex
}

// NewBlackboardStore creates a new BlackboardStore backed by SQLite.
func NewBlackboardStore(path string) (*BlackboardStore, error) {
	if path == "" {
		return nil, fmt.Errorf("sqlite path is required")
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Optimize SQLite performance
	db.Exec("PRAGMA journal_mode=WAL;")
	db.Exec("PRAGMA synchronous=NORMAL;")

	store := &BlackboardStore{
		db:   db,
		path: path,
	}

	if err := store.initDB(); err != nil {
		db.Close()
		return nil, err
	}

	return store, nil
}

func (s *BlackboardStore) initDB() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
	CREATE TABLE IF NOT EXISTS blackboard_kv (
		namespace TEXT NOT NULL,
		key TEXT NOT NULL,
		value TEXT NOT NULL,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (namespace, key)
	);
	`
	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create blackboard table: %w", err)
	}
	return nil
}

// Set stores a key-value pair in the blackboard.
func (s *BlackboardStore) Set(ctx context.Context, namespace, key string, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	valBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	query := `
	INSERT INTO blackboard_kv (namespace, key, value, updated_at)
	VALUES (?, ?, ?, ?)
	ON CONFLICT(namespace, key) DO UPDATE SET
		value=excluded.value,
		updated_at=excluded.updated_at;
	`
	_, err = s.db.ExecContext(ctx, query, namespace, key, string(valBytes), time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to set value in blackboard: %w", err)
	}

	logging.GetLogger().Debug("Blackboard Set", "namespace", namespace, "key", key)
	return nil
}

// Get retrieves a value from the blackboard.
func (s *BlackboardStore) Get(ctx context.Context, namespace, key string) (interface{}, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT value FROM blackboard_kv WHERE namespace = ? AND key = ?`
	row := s.db.QueryRowContext(ctx, query, namespace, key)

	var valStr string
	err := row.Scan(&valStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("failed to get value from blackboard: %w", err)
	}

	var value interface{}
	if err := json.Unmarshal([]byte(valStr), &value); err != nil {
		return nil, true, fmt.Errorf("failed to unmarshal value: %w", err)
	}

	logging.GetLogger().Debug("Blackboard Get", "namespace", namespace, "key", key)
	return value, true, nil
}

// Delete removes a key from the blackboard.
func (s *BlackboardStore) Delete(ctx context.Context, namespace, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `DELETE FROM blackboard_kv WHERE namespace = ? AND key = ?`
	_, err := s.db.ExecContext(ctx, query, namespace, key)
	if err != nil {
		return fmt.Errorf("failed to delete value from blackboard: %w", err)
	}

	logging.GetLogger().Debug("Blackboard Delete", "namespace", namespace, "key", key)
	return nil
}

// Close closes the SQLite database connection.
func (s *BlackboardStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.Close()
}
