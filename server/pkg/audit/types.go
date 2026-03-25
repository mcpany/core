// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package audit

import (
	"context"
	"encoding/json"
	"time"
)

// Entry represents a single audit log entry.
// Summary: Data structure representing a single recorded audit event.
type Entry struct {
	Timestamp  time.Time       `json:"timestamp"`
	ToolName   string          `json:"tool_name"`
	UserID     string          `json:"user_id,omitempty"`
	ProfileID  string          `json:"profile_id,omitempty"`
	TraceID    string          `json:"trace_id,omitempty"`
	SpanID     string          `json:"span_id,omitempty"`
	ParentID   string          `json:"parent_id,omitempty"`
	Arguments  json.RawMessage `json:"arguments,omitempty"`
	Result     any             `json:"result,omitempty"`
	Error      string          `json:"error,omitempty"`
	Duration   string          `json:"duration"`
	DurationMs int64           `json:"duration_ms"`
}

// Filter defines the filters for reading audit logs.
// Summary: Criteria for filtering and querying audit log entries.
type Filter struct {
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	ToolName  string     `json:"tool_name,omitempty"`
	UserID    string     `json:"user_id,omitempty"`
	ProfileID string     `json:"profile_id,omitempty"`
	Limit     int        `json:"limit,omitempty"`
	Offset    int        `json:"offset,omitempty"`
}

// Store defines the interface for audit log storage.
// Summary: Interface for audit log persistence and retrieval.
type Store interface {
	// Write writes an audit entry to the store.
	// Summary: Persists a single audit log entry.
	// Parameters:
	//   - ctx (context.Context): The context for the write operation.
	//   - entry (Entry): The audit entry to be recorded.
	// Returns:
	//   - error: An error if the write operation fails.
	Write(ctx context.Context, entry Entry) error

	// Read reads audit entries from the store based on the filter.
	// Summary: Retrieves a list of audit entries that match the provided filter.
	// Parameters:
	//   - ctx (context.Context): The context for the read operation.
	//   - filter (Filter): Criteria for selecting audit entries.
	// Returns:
	//   - []Entry: The list of audit entries matching the filter.
	//   - error: An error if the read operation fails.
	Read(ctx context.Context, filter Filter) ([]Entry, error)

	// Close closes the store.
	// Summary: Shuts down the audit store and releases associated resources.
	// Returns:
	//   - error: An error if closing the store fails.
	Close() error
}
