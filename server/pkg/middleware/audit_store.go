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

// AuditStore defines the interface for audit log storage.
type AuditStore interface {
	// Write writes an audit entry to the store.
	Write(ctx context.Context, entry AuditEntry) error
	// Close closes the store.
	Close() error
}
