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

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/audit"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/tool"
	"google.golang.org/protobuf/proto"
)

// AuditMiddleware provides audit logging for tool executions.
type AuditMiddleware struct {
	mu          sync.RWMutex
	config      *configv1.AuditConfig
	store       audit.Store
	redactor    *Redactor
	broadcaster *logging.Broadcaster
}

// NewAuditMiddleware initializes a new AuditMiddleware with the provided configuration.
//
// Parameters:
//   - auditConfig: *configv1.AuditConfig. The configuration for audit logging.
//
// Returns:
//   - *AuditMiddleware: The initialized middleware.
//   - error: An error if initialization fails.
func NewAuditMiddleware(auditConfig *configv1.AuditConfig) (*AuditMiddleware, error) {
	m := &AuditMiddleware{
		config:      auditConfig,
		broadcaster: logging.NewBroadcaster(),
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
		var store audit.Store
		var err error

		// Determine storage type
		storageType := config.GetStorageType()
		if storageType == configv1.AuditConfig_STORAGE_TYPE_UNSPECIFIED {
			storageType = configv1.AuditConfig_STORAGE_TYPE_FILE
		}

		switch storageType {
		case configv1.AuditConfig_STORAGE_TYPE_POSTGRES:
			store, err = audit.NewPostgresAuditStore(config.GetOutputPath())
		case configv1.AuditConfig_STORAGE_TYPE_SQLITE:
			store, err = audit.NewSQLiteAuditStore(config.GetOutputPath())
		case configv1.AuditConfig_STORAGE_TYPE_FILE:
			store, err = audit.NewFileAuditStore(config.GetOutputPath())
		case configv1.AuditConfig_STORAGE_TYPE_WEBHOOK:
			store = audit.NewWebhookAuditStore(config.GetWebhookUrl(), config.GetWebhookHeaders())
		case configv1.AuditConfig_STORAGE_TYPE_SPLUNK:
			store = audit.NewSplunkAuditStore(config.GetSplunk())
		case configv1.AuditConfig_STORAGE_TYPE_DATADOG:
			store = audit.NewDatadogAuditStore(config.GetDatadog())
		default:
			store, err = audit.NewFileAuditStore(config.GetOutputPath())
		}

		if err != nil {
			return fmt.Errorf("failed to initialize audit store: %w", err)
		}
		m.store = store
	}
	return nil
}

// SetStore sets the audit store.
// This is primarily used for testing.
func (m *AuditMiddleware) SetStore(store audit.Store) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store = store
}

// UpdateConfig updates the audit middleware configuration at runtime.
//
// Parameters:
//   - auditConfig: *configv1.AuditConfig. The new configuration.
//
// Returns:
//   - error: An error if the update fails.
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
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - req: *tool.ExecutionRequest. The execution request.
//   - next: tool.ExecutionFunc. The next function in the execution chain.
//
// Returns:
//   - any: The result of the execution.
//   - error: An error if the execution fails.
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
	entry := audit.Entry{
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

func (m *AuditMiddleware) writeLog(ctx context.Context, store audit.Store, entry audit.Entry) {
	// Broadcast first for real-time updates
	if m.broadcaster != nil {
		b, err := json.Marshal(entry)
		if err == nil {
			m.broadcaster.Broadcast(b)
		}
	}

	if store == nil {
		return
	}
	if err := store.Write(ctx, entry); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write audit log: %v\n", err)
	}
}

// SubscribeWithHistory returns a channel that will receive broadcast messages,
// and the current history of messages.
func (m *AuditMiddleware) SubscribeWithHistory() (chan []byte, [][]byte) {
	return m.broadcaster.SubscribeWithHistory()
}

// GetHistory returns the current broadcast history.
func (m *AuditMiddleware) GetHistory() [][]byte {
	return m.broadcaster.GetHistory()
}

// Unsubscribe removes a subscriber channel.
func (m *AuditMiddleware) Unsubscribe(ch chan []byte) {
	m.broadcaster.Unsubscribe(ch)
}

// Read reads audit entries from the underlying store.
func (m *AuditMiddleware) Read(ctx context.Context, filter audit.Filter) ([]audit.Entry, error) {
	m.mu.RLock()
	store := m.store
	m.mu.RUnlock()

	if store == nil {
		return nil, fmt.Errorf("audit store not initialized")
	}
	return store.Read(ctx, filter)
}

// Close closes the underlying store.
//
// Returns an error if the operation fails.
func (m *AuditMiddleware) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.store != nil {
		return m.store.Close()
	}
	return nil
}
