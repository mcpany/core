// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

// AuditMiddleware provides audit logging for tool executions.
type AuditMiddleware struct {
	config *configv1.AuditConfig
	store  AuditStore
}

// NewAuditMiddleware creates a new AuditMiddleware.
func NewAuditMiddleware(config *configv1.AuditConfig) (*AuditMiddleware, error) {
	m := &AuditMiddleware{
		config: config,
	}
	if config != nil && config.GetEnabled() {
		var store AuditStore
		var err error

		// Determine storage type
		storageType := config.GetStorageType()
		// Backward compatibility: if unspecified but output_path is set, default to FILE
		// Also if output_path is empty, default to FILE (stdout)
		if storageType == configv1.AuditConfig_STORAGE_TYPE_UNSPECIFIED {
			storageType = configv1.AuditConfig_STORAGE_TYPE_FILE
		}

		switch storageType {
		case configv1.AuditConfig_STORAGE_TYPE_POSTGRES:
			store, err = NewPostgresAuditStore(config.GetOutputPath())
		case configv1.AuditConfig_STORAGE_TYPE_SQLITE:
			store, err = NewSQLiteAuditStore(config.GetOutputPath())
		case configv1.AuditConfig_STORAGE_TYPE_FILE:
			store, err = NewFileAuditStore(config.GetOutputPath())
		default:
			store, err = NewFileAuditStore(config.GetOutputPath())
		}

		if err != nil {
			return nil, fmt.Errorf("failed to initialize audit store: %w", err)
		}
		m.store = store
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
	m.writeLog(ctx, entry)

	return result, err
}

func (m *AuditMiddleware) writeLog(ctx context.Context, entry AuditEntry) {
	if m.store == nil {
		return
	}
	if err := m.store.Write(ctx, entry); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write audit log: %v\n", err)
	}
}

// Close closes the underlying store.
func (m *AuditMiddleware) Close() error {
	if m.store != nil {
		return m.store.Close()
	}
	return nil
}
