// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"time"
)

// AuditEntry represents a single audit log entry.
type AuditEntry struct {
	Timestamp  time.Time       `json:"timestamp"`
	ToolName   string          `json:"tool_name"`
	UserID     string          `json:"user_id,omitempty"`
	ProfileID  string          `json:"profile_id,omitempty"`
	Arguments  json.RawMessage `json:"arguments,omitempty"`
	Result     any             `json:"result,omitempty"`
	Error      string          `json:"error,omitempty"`
	Duration   string          `json:"duration"`
	DurationMs int64           `json:"duration_ms"`
}

// AuditFilter defines the filters for reading audit logs.
type AuditFilter struct {
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	ToolName  string     `json:"tool_name,omitempty"`
	UserID    string     `json:"user_id,omitempty"`
	ProfileID string     `json:"profile_id,omitempty"`
	Limit     int        `json:"limit,omitempty"`
	Offset    int        `json:"offset,omitempty"`
}

// AuditStore defines the interface for audit log storage.
type AuditStore interface {
	// Write writes an audit entry to the store.
	//
	// ctx is the context for the request.
	// entry is the entry.
	//
	// Returns an error if the operation fails.
	Write(ctx context.Context, entry AuditEntry) error
	// Read reads audit entries from the store based on the filter.
	//
	// ctx is the context for the request.
	// filter is the filter to apply.
	//
	// Returns the entries and an error if the operation fails.
	Read(ctx context.Context, filter AuditFilter) ([]AuditEntry, error)
	// Close closes the store.
	//
	// Returns an error if the operation fails.
	Close() error
}
