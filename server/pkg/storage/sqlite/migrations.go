// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package sqlite

import (
	"database/sql"
	"fmt"
	"log/slog"
)

// Migration represents a single database migration.
type Migration struct {
	Version     int
	Description string
	Up          func(tx *sql.Tx) error
}

var migrations = []Migration{
	{
		Version:     1,
		Description: "Initial schema",
		Up: func(tx *sql.Tx) error {
			query := `
			CREATE TABLE IF NOT EXISTS upstream_services (
				id TEXT PRIMARY KEY,
				name TEXT UNIQUE NOT NULL,
				config_json TEXT NOT NULL,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
			);
			`
			_, err := tx.Exec(query)
			return err
		},
	},
}

// Migrate applies all pending migrations to the database.
func Migrate(db *sql.DB) error {
	// Create schema_versions table if it doesn't exist
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_versions (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create schema_versions table: %w", err)
	}

	// Get current version
	var currentVersion int
	err = db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_versions").Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("failed to get current schema version: %w", err)
	}

	slog.Info("Current database schema version", "version", currentVersion)

	for _, m := range migrations {
		if m.Version > currentVersion {
			slog.Info("Applying migration", "version", m.Version, "description", m.Description)
			if err := applyMigration(db, m); err != nil {
				return fmt.Errorf("failed to apply migration %d: %w", m.Version, err)
			}
			currentVersion = m.Version
		}
	}

	slog.Info("Database migration completed", "version", currentVersion)
	return nil
}

func applyMigration(db *sql.DB, m Migration) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := m.Up(tx); err != nil {
		return err
	}

	_, err = tx.Exec("INSERT INTO schema_versions (version) VALUES (?)", m.Version)
	if err != nil {
		return err
	}

	return tx.Commit()
}
