// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// AuditEntry represents a single audit log entry.
type AuditEntry struct {
	Timestamp    time.Time       `json:"timestamp"`
	ToolName     string          `json:"tool_name"`
	UserID       string          `json:"user_id,omitempty"`
	ProfileID    string          `json:"profile_id,omitempty"`
	Arguments    json.RawMessage `json:"arguments,omitempty"`
	Result       any             `json:"result,omitempty"`
	Error        string          `json:"error,omitempty"`
	Duration     string          `json:"duration"`
	DurationMs   int64           `json:"duration_ms"`
	PreviousHash string          `json:"prev_hash"`
	Hash         string          `json:"hash"`
}

// AuditStore defines the interface for audit log storage.
type AuditStore interface {
	// Write writes an audit entry to the store.
	Write(entry AuditEntry) error
	// Verify checks the integrity of the audit logs.
	Verify() (bool, error)
	// Close closes the store.
	Close() error
}

// ComputeAuditHash computes the SHA256 hash of an audit entry.
func ComputeAuditHash(entry AuditEntry, prevHash string) string {
	argsJSON := "{}"
	if len(entry.Arguments) > 0 {
		dst := new(bytes.Buffer)
		if err := json.Compact(dst, entry.Arguments); err == nil {
			argsJSON = dst.String()
		} else {
			argsJSON = string(entry.Arguments)
		}
	}

	resultJSON := "{}"
	if entry.Result != nil {
		if b, err := json.Marshal(entry.Result); err == nil {
			resultJSON = string(b)
		}
	}

	ts := entry.Timestamp.Format(time.RFC3339Nano)

	data := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%d|%s",
		ts, entry.ToolName, entry.UserID, entry.ProfileID, argsJSON, resultJSON, entry.Error, entry.DurationMs, prevHash)
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:])
}
