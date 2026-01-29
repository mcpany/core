// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package audit

import (
	"context"
	"encoding/json"
	"time"
)

// Entry represents a single audit log entry.
type Entry struct {
	// Timestamp is when the event occurred.
	Timestamp time.Time `json:"timestamp"`
	// ToolName is the name of the tool executed.
	ToolName string `json:"tool_name"`
	// UserID is the ID of the user who initiated the action.
	UserID string `json:"user_id,omitempty"`
	// ProfileID is the ID of the profile used.
	ProfileID string `json:"profile_id,omitempty"`
	// Arguments are the tool input arguments.
	Arguments json.RawMessage `json:"arguments,omitempty"`
	// Result is the output of the tool execution.
	Result any `json:"result,omitempty"`
	// Error is the error message if the execution failed.
	Error string `json:"error,omitempty"`
	// Duration is the execution duration as a string.
	Duration string `json:"duration"`
	// DurationMs is the execution duration in milliseconds.
	DurationMs int64 `json:"duration_ms"`
}

// Filter defines the filters for reading audit logs.
type Filter struct {
	// StartTime filters logs after this time.
	StartTime *time.Time `json:"start_time,omitempty"`
	// EndTime filters logs before this time.
	EndTime *time.Time `json:"end_time,omitempty"`
	// ToolName filters logs by tool name.
	ToolName string `json:"tool_name,omitempty"`
	// UserID filters logs by user ID.
	UserID string `json:"user_id,omitempty"`
	// ProfileID filters logs by profile ID.
	ProfileID string `json:"profile_id,omitempty"`
	// Limit is the maximum number of logs to return.
	Limit int `json:"limit,omitempty"`
	// Offset is the number of logs to skip.
	Offset int `json:"offset,omitempty"`
}

// Store defines the interface for audit log storage.
type Store interface {
	// Write writes an audit entry to the store.
	//
	// ctx is the context for the request.
	// entry is the entry.
	//
	// Returns an error if the operation fails.
	Write(ctx context.Context, entry Entry) error
	// Read reads audit entries from the store based on the filter.
	//
	// ctx is the context for the request.
	// filter is the filter to apply.
	//
	// Returns the entries and an error if the operation fails.
	Read(ctx context.Context, filter Filter) ([]Entry, error)
	// Close closes the store.
	//
	// Returns an error if the operation fails.
	Close() error
}
