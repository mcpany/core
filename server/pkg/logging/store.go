// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import "context"

// LogStore defines the interface for persisting logs.
type LogStore interface {
	// SaveLog saves a log entry to the store.
	SaveLog(ctx context.Context, entry LogEntry) error

	// GetRecentLogs retrieves the most recent log entries.
	// limit is the maximum number of logs to return.
	GetRecentLogs(ctx context.Context, limit int) ([]LogEntry, error)
}
