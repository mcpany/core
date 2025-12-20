// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

// AuditMiddleware provides audit logging for tool executions.
type AuditMiddleware struct {
	config *configv1.AuditConfig
	mu     sync.Mutex
	file   *os.File
}

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

// NewAuditMiddleware creates a new AuditMiddleware.
func NewAuditMiddleware(config *configv1.AuditConfig) (*AuditMiddleware, error) {
	m := &AuditMiddleware{
		config: config,
	}
	if config != nil && config.GetEnabled() && config.GetOutputPath() != "" {
		f, err := os.OpenFile(config.GetOutputPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644) //nolint:gosec
		if err != nil {
			return nil, fmt.Errorf("failed to open audit log file: %w", err)
		}
		m.file = f
	}
	return m, nil
}

// Execute intercepts tool execution to log audit events.
func (m *AuditMiddleware) Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error) {
	if m.config == nil || !m.config.GetEnabled() {
		return next(ctx, req)
	}

	start := time.Now()

	// Execute the tool
	result, err := next(ctx, req)

	duration := time.Since(start)

	// Prepare audit entry
	entry := AuditEntry{
		Timestamp:  start,
		ToolName:   req.ToolName,
		Duration:   duration.String(),
		DurationMs: duration.Milliseconds(),
	}

	if userID, ok := auth.UserFromContext(ctx); ok {
		entry.UserID = userID
	}
	if profileID, ok := auth.ProfileIDFromContext(ctx); ok {
		entry.ProfileID = profileID
	}

	if m.config.GetLogArguments() {
		// Try to marshal arguments to RawMessage to avoid double escaping if it's already structured
		argsBytes, marshalErr := json.Marshal(req.ToolInputs)
		if marshalErr == nil {
			entry.Arguments = json.RawMessage(argsBytes)
		}
	}

	if err != nil {
		entry.Error = err.Error()
	}

	if m.config.GetLogResults() && err == nil {
		entry.Result = result
	}

	// Write log
	m.writeLog(entry)

	return result, err
}

func (m *AuditMiddleware) writeLog(entry AuditEntry) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var w *os.File
	if m.file != nil {
		w = m.file
	} else {
		// Fallback to stdout if enabled but no path
		w = os.Stdout
	}

	// JSON encode
	if err := json.NewEncoder(w).Encode(entry); err != nil {
		// Last resort: print to stderr
		fmt.Fprintf(os.Stderr, "Failed to write audit log: %v\n", err)
	}
}

// Close closes the underlying file.
func (m *AuditMiddleware) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.file != nil {
		return m.file.Close()
	}
	return nil
}
