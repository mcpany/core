// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package uab

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// UniversalAgentBus represents the core logic for the Universal Agent Bus.
// It manages multi-agent sessions and transports.
type UniversalAgentBus struct {
	db *sql.DB
}

// NewUniversalAgentBus creates a new instance of the Universal Agent Bus.
// It initializes a dedicated SQLite database for UAB state.
func NewUniversalAgentBus(path string) (*UniversalAgentBus, error) {
	if path == "" {
		return nil, fmt.Errorf("sqlite path is required for UAB")
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database for UAB: %w", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS uab_sessions (
		id TEXT PRIMARY KEY,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS uab_transports (
		id TEXT PRIMARY KEY,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	ctxSchema, cancelSchema := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelSchema()

	if _, err := db.ExecContext(ctxSchema, schema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to create UAB tables: %w", err)
	}

	return &UniversalAgentBus{db: db}, nil
}

// RegisterSession registers a new multi-agent session.
func (uab *UniversalAgentBus) RegisterSession(ctx context.Context, id string) error {
	_, err := uab.db.ExecContext(ctx, "INSERT OR IGNORE INTO uab_sessions (id) VALUES (?)", id)
	return err
}

// RegisterTransport registers a new transport connection.
func (uab *UniversalAgentBus) RegisterTransport(ctx context.Context, id string) error {
	_, err := uab.db.ExecContext(ctx, "INSERT OR IGNORE INTO uab_transports (id) VALUES (?)", id)
	return err
}

// GetSessionCount returns the number of active multi-agent sessions.
func (uab *UniversalAgentBus) GetSessionCount(ctx context.Context) (int, error) {
	var count int
	err := uab.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM uab_sessions").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetTransportCount returns the number of active transports.
func (uab *UniversalAgentBus) GetTransportCount(ctx context.Context) (int, error) {
	var count int
	err := uab.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM uab_transports").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// Close gracefully closes the database connection.
func (uab *UniversalAgentBus) Close() error {
	return uab.db.Close()
}
