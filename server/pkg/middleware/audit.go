// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/audit"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/tool"
	"google.golang.org/protobuf/proto"
)

// AuditMiddleware represents the public AuditMiddleware entity.
//
// Summary: Defines the structured data model representing a middleware.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type AuditMiddleware struct {
	mu          sync.RWMutex
	config      *configv1.AuditConfig
	store       audit.Store
	redactor    *Redactor
	broadcaster *logging.Broadcaster
}

// NewAuditMiddleware serves as a public interface for interacting with NewAuditMiddleware.
//
// Summary: Constructs and returns an initialized audit middleware ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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
	} else {
		// Log that config was nil or disabled
		if config == nil {
			logging.GetLogger().Info("AuditMiddleware.initializeStore: config is nil")
		} else {
			logging.GetLogger().Info("AuditMiddleware.initializeStore: config.Enabled is false")
		}
	}
	return nil
}

// SetStore serves as a public interface for interacting with SetStore.
//
// Summary: Set the store appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *AuditMiddleware) SetStore(store audit.Store) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store = store
}

// UpdateConfig serves as a public interface for interacting with UpdateConfig.
//
// Summary: Update the config appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// Execute serves as a public interface for interacting with Execute.
//
// Summary: Execute the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

	// Trace Context
	traceID := GetTraceID(ctx)
	if traceID == "" {
		traceID = strings.ReplaceAll(uuid.New().String(), "-", "")
	}
	parentID := GetSpanID(ctx) // The parent is the current span in context
	spanID := strings.ReplaceAll(uuid.New().String(), "-", "")[:16]

	// Update context for downstream (Recursive Tracing)
	ctx = WithTraceContext(ctx, traceID, spanID, parentID)

	// Execute the tool
	result, err := next(ctx, req)

	duration := time.Since(start)

	// Prepare audit entry
	entry := audit.Entry{
		Timestamp:  start,
		ToolName:   req.ToolName,
		Duration:   duration.String(),
		DurationMs: duration.Milliseconds(),
		TraceID:    traceID,
		SpanID:     spanID,
		ParentID:   parentID,
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
		// ⚡ BOLT: Pass struct directly to avoid JSON marshaling.
		m.broadcaster.Broadcast(entry)
	}

	if store == nil {
		return
	}
	if err := store.Write(ctx, entry); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write audit log: %v\n", err)
	}
}

// ClearHistory serves as a public interface for interacting with ClearHistory.
//
// Summary: Clear the history appropriately based on current system conditions.
//
// Parameters:
//   - None.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *AuditMiddleware) ClearHistory() {
	if m.broadcaster != nil {
		m.broadcaster.ClearHistory()
	}
}

// Broadcast serves as a public interface for interacting with Broadcast.
//
// Summary: Broadcast the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *AuditMiddleware) Broadcast(entry audit.Entry) {
	if m.broadcaster != nil {
		m.broadcaster.Broadcast(entry)
	}
}

// SubscribeWithHistory serves as a public interface for interacting with SubscribeWithHistory.
//
// Summary: Subscribe the with history appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *AuditMiddleware) SubscribeWithHistory() (chan any, []any) {
	return m.broadcaster.SubscribeWithHistory()
}

// GetHistory serves as a public interface for interacting with GetHistory.
//
// Summary: Fetches and returns the underlying history from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *AuditMiddleware) GetHistory() []any {
	return m.broadcaster.GetHistory()
}

// Unsubscribe serves as a public interface for interacting with Unsubscribe.
//
// Summary: Unsubscribe the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *AuditMiddleware) Unsubscribe(ch chan any) {
	m.broadcaster.Unsubscribe(ch)
}

// Read serves as a public interface for interacting with Read.
//
// Summary: Read the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *AuditMiddleware) Read(ctx context.Context, filter audit.Filter) ([]audit.Entry, error) {
	m.mu.RLock()
	store := m.store
	m.mu.RUnlock()

	if store == nil {
		return nil, fmt.Errorf("audit store not initialized")
	}
	return store.Read(ctx, filter)
}

// Close serves as a public interface for interacting with Close.
//
// Summary: Close the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *AuditMiddleware) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.store != nil {
		return m.store.Close()
	}
	return nil
}

// Write serves as a public interface for interacting with Write.
//
// Summary: Write the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *AuditMiddleware) Write(ctx context.Context, entry audit.Entry) error {
	m.mu.RLock()
	store := m.store
	m.mu.RUnlock()

	if store == nil {
		return fmt.Errorf("audit store not initialized")
	}
	m.writeLog(ctx, store, entry)
	return nil
}
