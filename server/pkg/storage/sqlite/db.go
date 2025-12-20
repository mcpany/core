// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // Register sqlite driver
)

type DB struct {
	*sql.DB
}

func NewDB(path string) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	// Enable WAL mode for better concurrency
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping sqlite db: %w", err)
	}

	// Set pragmas
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return nil, fmt.Errorf("failed to set WAL mode: %w", err)
	}
	if _, err := db.Exec("PRAGMA busy_timeout=5000;"); err != nil {
		return nil, fmt.Errorf("failed to set busy_timeout: %w", err)
	}

	if err := initSchema(db); err != nil {
		db.Close()
		return nil, err
	}

	return &DB{db}, nil
}

func initSchema(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS upstream_services (
		id TEXT PRIMARY KEY,
		name TEXT UNIQUE NOT NULL,
		config_json TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}
	return nil
}
