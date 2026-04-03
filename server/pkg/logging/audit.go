// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/audit"
)

// AuditHandler represents the public AuditHandler entity.
//
// Summary: Defines the structured data model representing a handler.
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
type AuditHandler struct {
	next   slog.Handler
	config *configv1.AuditConfig
	store  audit.Store
}

// NewAuditHandler serves as a public interface for interacting with NewAuditHandler.
//
// Summary: Constructs and returns an initialized audit handler ready for consumption.
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
func NewAuditHandler(next slog.Handler, config *configv1.AuditConfig) *AuditHandler {
	h := &AuditHandler{
		next:   next,
		config: config,
	}
	if config != nil && config.GetEnabled() {
		h.initializeStore(config)
	}
	return h
}

func (h *AuditHandler) initializeStore(config *configv1.AuditConfig) {
	storageType := config.GetStorageType()
	if storageType == configv1.AuditConfig_STORAGE_TYPE_UNSPECIFIED {
		storageType = configv1.AuditConfig_STORAGE_TYPE_FILE
	}

	var store audit.Store
	var err error

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
		fmt.Fprintf(os.Stderr, "Failed to initialize audit handler store: %v\n", err)
		return
	}
	h.store = store
}

// Enabled serves as a public interface for interacting with Enabled.
//
// Summary: Enabled the  appropriately based on current system conditions.
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
func (h *AuditHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

// Handle serves as a public interface for interacting with Handle.
//
// Summary: Handle the  appropriately based on current system conditions.
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
func (h *AuditHandler) Handle(ctx context.Context, r slog.Record) error {
	// 1. Export the record
	if err := h.Export(ctx, r); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to export audit log: %v\n", err)
	}

	// 2. Delegate to next handler
	return h.next.Handle(ctx, r)
}

// WithAttrs serves as a public interface for interacting with WithAttrs.
//
// Summary: With the attrs appropriately based on current system conditions.
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
func (h *AuditHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &AuditHandler{
		next:   h.next.WithAttrs(attrs),
		config: h.config,
		store:  h.store,
	}
}

// WithGroup serves as a public interface for interacting with WithGroup.
//
// Summary: With the group appropriately based on current system conditions.
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
func (h *AuditHandler) WithGroup(name string) slog.Handler {
	return &AuditHandler{
		next:   h.next.WithGroup(name),
		config: h.config,
		store:  h.store,
	}
}

// Export serves as a public interface for interacting with Export.
//
// Summary: Export the  appropriately based on current system conditions.
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
func (h *AuditHandler) Export(ctx context.Context, r slog.Record) error {
	if h.store == nil {
		return nil
	}

	// Convert slog.Record to AuditEntry
	entry := audit.Entry{
		Timestamp: r.Time,
		ToolName:  "log:" + r.Message, // Use ToolName field to hold the message
	}

	attrs := make(map[string]any)
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.Any()
		return true
	})

	if len(attrs) > 0 {
		data, err := json.Marshal(attrs)
		if err == nil {
			entry.Arguments = json.RawMessage(data)
		}
	}

	return h.store.Write(ctx, entry)
}
