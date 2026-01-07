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

	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

// AuditMiddleware provides audit logging for tool executions.
type AuditMiddleware struct {
	mu       sync.RWMutex
	config   *configv1.AuditConfig
	store    AuditStore
	redactor *Redactor
}

// NewAuditMiddleware creates a new AuditMiddleware.
func NewAuditMiddleware(auditConfig *configv1.AuditConfig) (*AuditMiddleware, error) {
	m := &AuditMiddleware{
		config: auditConfig,
	}
	if err := m.initializeStore(auditConfig); err != nil {
		return nil, err
	}
	// Initialize redactor with current global settings
	m.redactor = NewRedactor(config.GlobalSettings().GetDlp(), nil)
	return m, nil
}

func (m *AuditMiddleware) initializeStore(config *configv1.AuditConfig) error {
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
		case configv1.AuditConfig_STORAGE_TYPE_WEBHOOK:
			store = NewWebhookAuditStore(config.GetWebhookUrl(), config.GetWebhookHeaders())
		case configv1.AuditConfig_STORAGE_TYPE_SPLUNK:
			store = NewSplunkAuditStore(config.GetSplunk())
		case configv1.AuditConfig_STORAGE_TYPE_DATADOG:
			store = NewDatadogAuditStore(config.GetDatadog())
		default:
			store, err = NewFileAuditStore(config.GetOutputPath())
		}

		if err != nil {
			return fmt.Errorf("failed to initialize audit store: %w", err)
		}
		m.store = store
	}
	return nil
}

// UpdateConfig updates the audit configuration safely.
func (m *AuditMiddleware) UpdateConfig(auditConfig *configv1.AuditConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Update redactor on config update (it uses global DLP config, which might also change,
	// but UpdateConfig is usually called when config file changes, so good time to refresh)
	m.redactor = NewRedactor(config.GlobalSettings().GetDlp(), nil)

	// If config is nil, disable audit
	if auditConfig == nil {
		if m.store != nil {
			_ = m.store.Close()
			m.store = nil
		}
		m.config = nil
		return nil
	}

	// Check if storage config changed. If so, we need to re-initialize store.
	// For simplicity, we always re-initialize if enabled, or if we are enabling it.
	// Optimally, we check for diffs.
	needsReinit := false
	if m.config == nil {
		needsReinit = true
	} else if !proto.Equal(m.config, auditConfig) {
		needsReinit = true
	}

	if needsReinit {
		// Close old store
		if m.store != nil {
			_ = m.store.Close()
			m.store = nil
		}
		if err := m.initializeStore(auditConfig); err != nil {
			return err
		}
	}
	m.config = auditConfig
	return nil
}

// Execute intercepts tool execution to log audit events.
func (m *AuditMiddleware) Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error) {
	m.mu.RLock()
	auditConfig := m.config
	store := m.store
	redactor := m.redactor
	m.mu.RUnlock()

	if auditConfig == nil || !auditConfig.GetEnabled() {
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

	if auditConfig.GetLogArguments() {
		// Try to marshal arguments to RawMessage to avoid double escaping if it's already structured
		argsBytes, marshalErr := json.Marshal(req.ToolInputs)
		if marshalErr == nil {
			// Use Redactor to ensure no secrets are logged
			if redactor != nil {
				redactedBytes, err := redactor.RedactJSON(argsBytes)
				if err == nil {
					argsBytes = redactedBytes
				}
			}
			entry.Arguments = json.RawMessage(argsBytes)
		}
	}

	if err != nil {
		entry.Error = err.Error()
	}

	if auditConfig.GetLogResults() && err == nil {
		// Use Redactor for result too to ensure structs are handled correctly
		// and avoid side effects (modifying the result map if it's a map)
		// We marshal to JSON, redact, and then unmarshal or store as RawMessage if entry.Result supports it?
		// AuditEntry.Result is `any`. If we store redacted map, it's fine.
		// If we use RedactJSON, we get []byte.

		logResult := result
		if redactor != nil {
			// Best effort redaction
			jsonBytes, err := json.Marshal(result)
			if err == nil {
				redactedBytes, err := redactor.RedactJSON(jsonBytes)
				if err == nil {
					// We can store it as RawMessage if we change AuditEntry, but AuditEntry.Result is `any`.
					// Let's decode it back to generic interface to keep it compatible with whatever the store expects (usually JSON marshaling).
					var redactedResult interface{}
					if err := json.Unmarshal(redactedBytes, &redactedResult); err == nil {
						logResult = redactedResult
					}
				}
			}
		}
		entry.Result = logResult
	}

	// Write log
	m.writeLog(ctx, store, entry)

	return result, err
}

func (m *AuditMiddleware) writeLog(ctx context.Context, store AuditStore, entry AuditEntry) {
	if store == nil {
		return
	}
	if err := store.Write(ctx, entry); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write audit log: %v\n", err)
	}
}

// Close closes the underlying store.
func (m *AuditMiddleware) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.store != nil {
		return m.store.Close()
	}
	return nil
}
