// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

// LogStore defines the interface for persisting logs.
type LogStore interface {
	// Write persists a log entry.
	Write(entry LogEntry) error

	// Read retrieves the last N log entries.
	Read(limit int) ([]LogEntry, error)

	// Close closes the store.
	Close() error
}
