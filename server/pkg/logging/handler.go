// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"context"
	"encoding/json"
	"log/slog"
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"
)

// LogEntry is the structure for logs sent over WebSocket.
// It matches the frontend expectation.
type LogEntry struct {
	ID        string         `json:"id"`
	Timestamp string         `json:"timestamp"`
	Level     string         `json:"level"`
	Message   string         `json:"message"`
	Source    string         `json:"source,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// RecordToEntry converts a slog.Record to a LogEntry.
func RecordToEntry(r slog.Record) LogEntry {
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
	return entry
}

// BroadcastHandler implements slog.Handler and sends logs to the Broadcaster.
// It also persists logs if a store provider is configured.
type BroadcastHandler struct {
	broadcaster   *Broadcaster
	storeProvider func() LogStore
	attrs         []slog.Attr
	group         string
	mu            sync.Mutex
	level         slog.Level
}

// NewBroadcastHandler creates a new BroadcastHandler.
//
// broadcaster is the broadcaster.
// storeProvider is a function that returns the current LogStore (or nil).
// level is the minimum log level to broadcast.
//
// Returns the result.
func NewBroadcastHandler(broadcaster *Broadcaster, storeProvider func() LogStore, level slog.Level) *BroadcastHandler {
	return &BroadcastHandler{
		broadcaster:   broadcaster,
		storeProvider: storeProvider,
		level:         level,
	}
}

// Enabled returns true if the level is greater than or equal to the handler's level.
func (h *BroadcastHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

// Handle handles the log record by converting it to LogEntry, persisting it, and broadcasting it.
func (h *BroadcastHandler) Handle(ctx context.Context, r slog.Record) error {
	entry := RecordToEntry(r)

	// Persist to store if available
	if h.storeProvider != nil {
		if s := h.storeProvider(); s != nil {
			// Best effort write
			_ = s.Write(ctx, entry)
		}
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	h.broadcaster.Broadcast(data)
	return nil
}

// WithAttrs returns a new handler with the given attributes.
func (h *BroadcastHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.mu.Lock()
	defer h.mu.Unlock()
	return &BroadcastHandler{
		broadcaster:   h.broadcaster,
		storeProvider: h.storeProvider,
		attrs:         append(h.attrs, attrs...),
		group:         h.group,
		level:         h.level,
	}
}

// WithGroup returns a new handler with the given group.
func (h *BroadcastHandler) WithGroup(name string) slog.Handler {
	h.mu.Lock()
	defer h.mu.Unlock()
	return &BroadcastHandler{
		broadcaster:   h.broadcaster,
		storeProvider: h.storeProvider,
		attrs:         h.attrs,
		group:         name,
		level:         h.level,
	}
}

// TeeHandler is a slog.Handler that writes to multiple handlers.
type TeeHandler struct {
	handlers []slog.Handler
}

// NewTeeHandler creates a new TeeHandler.
func NewTeeHandler(handlers ...slog.Handler) *TeeHandler {
	return &TeeHandler{handlers: handlers}
}

// Enabled returns true if any of the handlers are enabled.
func (h *TeeHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// Handle forwards the record to all enabled handlers.
func (h *TeeHandler) Handle(ctx context.Context, r slog.Record) error {
	var err error
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, r.Level) {
			if e := handler.Handle(ctx, r); e != nil {
				err = e
			}
		}
	}
	return err
}

// WithAttrs returns a new TeeHandler with the attributes applied to all handlers.
func (h *TeeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}
	return NewTeeHandler(handlers...)
}

// WithGroup returns a new TeeHandler with the group applied to all handlers.
func (h *TeeHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithGroup(name)
	}
	return NewTeeHandler(handlers...)
}
