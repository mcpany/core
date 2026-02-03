// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import "time"

// LogEntry is the structure for logs sent over WebSocket and stored.
type LogEntry struct {
	ID        string         `json:"id"`
	Timestamp string         `json:"timestamp"`
	Level     string         `json:"level"`
	Message   string         `json:"message"`
	Source    string         `json:"source,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// LogFilter defines criteria for querying logs.
type LogFilter struct {
	Limit     int
	Offset    int
	Source    string
	Level     string
	StartTime *time.Time
	EndTime   *time.Time
	Search    string // Text search in message/source
}
