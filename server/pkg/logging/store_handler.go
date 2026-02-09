// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"context"
	"log/slog"
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"
)

// StoreHandler implements slog.Handler and writes logs to the LogStore.
type StoreHandler struct {
	store LogStore
	attrs []slog.Attr
	group string
	mu    sync.Mutex
	level slog.Level
}

// NewStoreHandler creates a new StoreHandler.
func NewStoreHandler(store LogStore, level slog.Level) *StoreHandler {
	return &StoreHandler{
		store: store,
		level: level,
	}
}

// Enabled returns true if the level is greater than or equal to the handler's level.
func (h *StoreHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

// Handle handles the log record by converting it to LogEntry and writing it to the store.
func (h *StoreHandler) Handle(ctx context.Context, r slog.Record) error {
	entry := LogEntry{
		ID:        uuid.New().String(),
		Timestamp: r.Time.Format(time.RFC3339),
		Level:     r.Level.String(),
		Message:   r.Message,
		Metadata:  make(map[string]any),
	}

	// Track source priority to ensure we get the most specific source
	sourcePriority := 0 // 0: none, 1: component/source, 2: toolName

	// Try to find source in attributes or use default
	r.Attrs(func(a slog.Attr) bool {
		// Collect all attributes into Metadata
		entry.Metadata[a.Key] = a.Value.Any()

		if a.Key == "source" || a.Key == "component" {
			if sourcePriority < 1 {
				entry.Source = a.Value.String()
				sourcePriority = 1
			}
		}
		if a.Key == "toolName" {
			entry.Source = a.Value.String()
			sourcePriority = 2
		}
		return true
	})

	// Also handle source from record PC if available
	if entry.Source == "" && r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		entry.Source = f.Function
	}

	// Write to store (synchronous for now)
	return h.store.Write(ctx, entry)
}

// WithAttrs returns a new handler with the given attributes.
func (h *StoreHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.mu.Lock()
	defer h.mu.Unlock()
	return &StoreHandler{
		store: h.store,
		attrs: append(h.attrs, attrs...),
		group: h.group,
		level: h.level,
	}
}

// WithGroup returns a new handler with the given group.
func (h *StoreHandler) WithGroup(name string) slog.Handler {
	h.mu.Lock()
	defer h.mu.Unlock()
	return &StoreHandler{
		store: h.store,
		attrs: h.attrs,
		group: name,
		level: h.level,
	}
}
