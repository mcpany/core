// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package audit

import (
	"context"
// Entry represents a single audit log entry.
//
// Summary: Represents a Entry.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
// Filter defines the filters for reading audit logs.
//
// Summary: Represents a Filter.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
type Filter struct {
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	ToolName  string     `json:"tool_name,omitempty"`
	UserID    string     `json:"user_id,omitempty"`
	ProfileID string     `json:"profile_id,omitempty"`
// Store defines the interface for audit log storage.
//
// Summary: Represents a Store.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
