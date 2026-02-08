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

// LogSaver defines the interface for saving logs.
type LogSaver interface {
	SaveLog(ctx context.Context, entry LogEntry) error
}

// BroadcastHandler implements slog.Handler and sends logs to the Broadcaster.
type BroadcastHandler struct {
	broadcaster *Broadcaster
	saver       LogSaver
	attrs       []slog.Attr
	group       string
	mu          sync.Mutex
	level       slog.Level
}

// SetSaver sets the log saver for the handler.
func (h *BroadcastHandler) SetSaver(saver LogSaver) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.saver = saver
}

// NewBroadcastHandler creates a new BroadcastHandler.
//
// broadcaster is the broadcaster.
// level is the minimum log level to broadcast.
//
// Returns the result.
func NewBroadcastHandler(broadcaster *Broadcaster, level slog.Level) *BroadcastHandler {
	return &BroadcastHandler{
		broadcaster: broadcaster,
		level:       level,
	}
}

// Enabled returns true if the level is greater than or equal to the handler's level.
//
// _ is an unused parameter.
// level is the log level.
//
// Returns true if successful.
func (h *BroadcastHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

// Handle handles the log record by converting it to LogEntry and broadcasting it.
//
// _ is an unused parameter.
// r is the r.
//
// Returns an error if the operation fails.
func (h *BroadcastHandler) Handle(_ context.Context, r slog.Record) error {
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

	// Add handler attributes
	for _, a := range h.attrs {
		entry.Metadata[a.Key] = a.Value.Any()
	}

	// Also handle source from record PC if available
	if entry.Source == "" && r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		entry.Source = f.Function
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	h.broadcaster.Broadcast(data)

	// Save to storage if configured
	// We read lock to access saver safely
	// Note: We don't hold the lock during SaveLog to avoid blocking
	h.mu.Lock()
	saver := h.saver
	h.mu.Unlock()

	if saver != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = saver.SaveLog(ctx, entry)
		}()
	}

	return nil
}

// WithAttrs returns a new handler with the given attributes.
//
// attrs is the attrs.
//
// Returns the result.
func (h *BroadcastHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.mu.Lock()
	defer h.mu.Unlock()
	return &BroadcastHandler{
		broadcaster: h.broadcaster,
		attrs:       append(h.attrs, attrs...),
		group:       h.group,
		level:       h.level,
	}
}

// WithGroup returns a new handler with the given group.
//
// name is the name of the resource.
//
// Returns the result.
func (h *BroadcastHandler) WithGroup(name string) slog.Handler {
	h.mu.Lock()
	defer h.mu.Unlock()
	return &BroadcastHandler{
		broadcaster: h.broadcaster,
		attrs:       h.attrs,
		group:       name,
		level:       h.level,
	}
}

// TeeHandler is a slog.Handler that writes to multiple handlers.
type TeeHandler struct {
	handlers []slog.Handler
}

// NewTeeHandler creates a new TeeHandler.
//
// handlers is the handlers.
//
// Returns the result.
func NewTeeHandler(handlers ...slog.Handler) *TeeHandler {
	return &TeeHandler{handlers: handlers}
}

// Enabled returns true if any of the handlers are enabled.
//
// ctx is the context for the request.
// level is the level.
//
// Returns true if successful.
func (h *TeeHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// Handle forwards the record to all enabled handlers.
//
// ctx is the context for the request.
// r is the r.
//
// Returns an error if the operation fails.
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
//
// attrs is the attrs.
//
// Returns the result.
func (h *TeeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}
	return NewTeeHandler(handlers...)
}

// WithGroup returns a new TeeHandler with the group applied to all handlers.
//
// name is the name of the resource.
//
// Returns the result.
func (h *TeeHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithGroup(name)
	}
	return NewTeeHandler(handlers...)
}
