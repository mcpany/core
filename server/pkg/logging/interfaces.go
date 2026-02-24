// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import "context"

// LogEntry is the structure for logs.
// It matches the frontend expectation and is used for persistence.
type LogEntry struct {
	ID        string         `json:"id"`
	Timestamp string         `json:"timestamp"`
	Level     string         `json:"level"`
	Message   string         `json:"message"`
	Source    string         `json:"source,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// LogPersister defines the interface for persisting logs.
type LogPersister interface {
	// SaveLog saves a single log entry.
	SaveLog(ctx context.Context, entry LogEntry) error

	// ListLogs retrieves logs with pagination.
	// limit: max number of logs to return.
	// offset: number of logs to skip.
	ListLogs(ctx context.Context, limit, offset int) ([]LogEntry, error)
}
