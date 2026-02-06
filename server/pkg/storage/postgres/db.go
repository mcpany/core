// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package postgres implements PostgreSQL storage for MCP Any configuration.
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // Register postgres driver
)

// DB wraps the sql.DB connection.
type DB struct {
	*sql.DB
}

// NewDB opens a PostgreSQL database connection.
//
// dsn is the dsn.
//
// Returns the result.
// Returns an error if the operation fails.
func NewDB(dsn string) (*DB, error) {
	return NewDBWithDriver("postgres", dsn)
}

// NewDBWithDriver opens a database connection with the specified driver.
//
// driver is the driver.
// dsn is the dsn.
//
// Returns the result.
// Returns an error if the operation fails.
func NewDBWithDriver(driver, dsn string) (*DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres db: %w", err)
	}

	// ⚡ Bolt Optimization: Set sensible connection pool defaults.
	// Default MaxOpenConns is 0 (unlimited), which can exhaust DB resources.
	// Default MaxIdleConns is 2, which causes high connection churn.
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.PingContext(context.Background()); err != nil {
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
//
// db is the db.
//
// Returns the result.
// Returns an error if the operation fails.
func NewDBFromSQLDB(db *sql.DB) (*DB, error) {
	if err := db.PingContext(context.Background()); err != nil {
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

	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		config_json TEXT NOT NULL,
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS global_settings (
		id INTEGER PRIMARY KEY CHECK (id = 1),
		config_json TEXT NOT NULL,
		updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS secrets (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		key TEXT NOT NULL,
		config_json TEXT NOT NULL,
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS profile_definitions (
		id TEXT PRIMARY KEY,
		name TEXT UNIQUE NOT NULL,
		config_json TEXT NOT NULL,
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS service_collections (
		id TEXT PRIMARY KEY,
		name TEXT UNIQUE NOT NULL,
		config_json TEXT NOT NULL,
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS user_tokens (
		user_id TEXT NOT NULL,
		service_id TEXT NOT NULL,
		config_json TEXT NOT NULL,
		updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (user_id, service_id)
	);

	CREATE TABLE IF NOT EXISTS logs (
		id TEXT PRIMARY KEY,
		timestamp TIMESTAMPTZ NOT NULL,
		level TEXT NOT NULL,
		message TEXT,
		source TEXT,
		metadata_json TEXT,
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp);
	CREATE INDEX IF NOT EXISTS idx_logs_level ON logs(level);
	CREATE INDEX IF NOT EXISTS idx_logs_source ON logs(source);
	`
	_, err := db.ExecContext(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}
	return nil
}
