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
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Source    string `json:"source,omitempty"`
	// Enhanced fields for Tool Executions
	Type     string `json:"type,omitempty"` // "general" or "tool_execution"
	ToolName string `json:"tool_name,omitempty"`
	Input    string `json:"input,omitempty"`
	Output   string `json:"output,omitempty"`
	Duration string `json:"duration,omitempty"`
	IsError  bool   `json:"is_error,omitempty"`
}

// BroadcastHandler implements slog.Handler and sends logs to the Broadcaster.
type BroadcastHandler struct {
	broadcaster *Broadcaster
	attrs       []slog.Attr
	group       string
	mu          sync.Mutex
}

// NewBroadcastHandler creates a new BroadcastHandler.
func NewBroadcastHandler(broadcaster *Broadcaster) *BroadcastHandler {
	return &BroadcastHandler{
		broadcaster: broadcaster,
	}
}

// Enabled always returns true as we want to capture all logs (or rely on parent logger level).
func (h *BroadcastHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

// Handle handles the log record by converting it to LogEntry and broadcasting it.
func (h *BroadcastHandler) Handle(_ context.Context, r slog.Record) error {
	entry := LogEntry{
		ID:        uuid.New().String(),
		Timestamp: r.Time.Format(time.RFC3339),
		Level:     r.Level.String(),
		Message:   r.Message,
		Type:      "general", // Default
	}

	// Try to find source in attributes or use default
	r.Attrs(func(a slog.Attr) bool {
		switch a.Key {
		case "source", "component":
			entry.Source = a.Value.String()
		case "log_type":
			entry.Type = a.Value.String()
		case "tool_name":
			entry.ToolName = a.Value.String()
		case "input":
			entry.Input = a.Value.String()
		case "output":
			entry.Output = a.Value.String()
		case "duration":
			entry.Duration = a.Value.String()
		case "is_error":
			entry.IsError = a.Value.Bool()
		}
		return true
	})

	// If message is empty but we have attributes, maybe format them?
	// For now, let's append attributes to message if it's debugging or specific keys
	// This is a simplification. Ideally we'd send structured data.

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
	return nil
}

// WithAttrs returns a new handler with the given attributes.
func (h *BroadcastHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.mu.Lock()
	defer h.mu.Unlock()
	return &BroadcastHandler{
		broadcaster: h.broadcaster,
		attrs:       append(h.attrs, attrs...),
		group:       h.group,
	}
}

// WithGroup returns a new handler with the given group.
func (h *BroadcastHandler) WithGroup(name string) slog.Handler {
	h.mu.Lock()
	defer h.mu.Unlock()
	return &BroadcastHandler{
		broadcaster: h.broadcaster,
		attrs:       h.attrs,
		group:       name,
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
