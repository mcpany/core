// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// AuditHandler is a slog.Handler that exports logs to audit sinks.
type AuditHandler struct {
	next   slog.Handler
	config *configv1.AuditConfig
}

// NewAuditHandler creates a new AuditHandler.
func NewAuditHandler(next slog.Handler, config *configv1.AuditConfig) *AuditHandler {
	return &AuditHandler{
		next:   next,
		config: config,
	}
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
	}
}

// WithGroup returns a new generic Handler with the given group.
func (h *AuditHandler) WithGroup(name string) slog.Handler {
	return &AuditHandler{
		next:   h.next.WithGroup(name),
		config: h.config,
	}
}

// Export sends the log record to the configued sinks.
func (h *AuditHandler) Export(ctx context.Context, r slog.Record) error {
	if h.config == nil {
		return nil
	}
	// Stub implementation
	return nil
}
