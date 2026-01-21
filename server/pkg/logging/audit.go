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

// AuditHandler is a slog.Handler that exports logs to audit sinks.
type AuditHandler struct {
	next   slog.Handler
	config *configv1.AuditConfig
	store  audit.Store
}

// NewAuditHandler creates a new AuditHandler.
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

// Enabled reports whether the handler handles records at the given level.
func (h *AuditHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

// Handle handles the Record.
func (h *AuditHandler) Handle(ctx context.Context, r slog.Record) error {
	// 1. Export the record
	if err := h.Export(ctx, r); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to export audit log: %v\n", err)
	}

	// 2. Delegate to next handler
	return h.next.Handle(ctx, r)
}

// WithAttrs returns a new generic Handler with the given attributes.
func (h *AuditHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &AuditHandler{
		next:   h.next.WithAttrs(attrs),
		config: h.config,
		store:  h.store,
	}
}

// WithGroup returns a new generic Handler with the given group.
func (h *AuditHandler) WithGroup(name string) slog.Handler {
	return &AuditHandler{
		next:   h.next.WithGroup(name),
		config: h.config,
		store:  h.store,
	}
}

// Export sends the log record to the configued sinks.
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
