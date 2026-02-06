// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import "time"

// LogPersister defines the interface for persisting logs.
type LogPersister interface {
	// SaveLog persists a log entry.
	SaveLog(entry LogEntry) error
}

// LogFilter defines filters for querying logs.
type LogFilter struct {
	StartTime *time.Time
	EndTime   *time.Time
	Level     string
	Source    string
	Search    string
	Limit     int
	Offset    int
}
