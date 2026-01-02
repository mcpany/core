// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package postgres implements PostgreSQL storage for MCP Any configuration.
package postgres

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // Register postgres driver
)

// DB wraps the sql.DB connection.
type DB struct {
	*sql.DB
}

// NewDB opens a PostgreSQL database connection.
func NewDB(dsn string) (*DB, error) {
	return NewDBWithDriver("postgres", dsn)
}

// NewDBWithDriver opens a database connection with the specified driver.
func NewDBWithDriver(driver, dsn string) (*DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres db: %w", err)
	}

	if err := db.PingContext(context.TODO()); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping postgres db: %w", err)
	}

	if err := initSchema(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	return &DB{db}, nil
}

// NewDBFromSQLDB creates a new DB wrapper from an existing sql.DB connection.
func NewDBFromSQLDB(db *sql.DB) (*DB, error) {
	if err := db.PingContext(context.TODO()); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	if err := initSchema(db); err != nil {
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	return &DB{db}, nil
}

func initSchema(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS upstream_services (
		id TEXT PRIMARY KEY,
		name TEXT UNIQUE NOT NULL,
		config_json TEXT NOT NULL,
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := db.ExecContext(context.TODO(), query)
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}
	return nil
}
