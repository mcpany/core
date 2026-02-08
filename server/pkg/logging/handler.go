// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"context"
	"encoding/json"
	"log/slog"
	"runtime"
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

// BroadcastHandler implements slog.Handler and sends logs to the Broadcaster.
type BroadcastHandler struct {
	broadcaster *Broadcaster
	level       slog.Level
	// modifiers is a chain of functions that build the metadata structure.
	// Each function takes the current map and returns the map where subsequent attributes should be added.
	modifiers []func(map[string]any) map[string]any
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
		modifiers:   make([]func(map[string]any) map[string]any, 0),
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

	// Apply modifiers to build the metadata structure and find the insertion point for record attributes
	currentMap := entry.Metadata
	for _, mod := range h.modifiers {
		currentMap = mod(currentMap)
	}

	// Track source priority to ensure we get the most specific source
	sourcePriority := 0 // 0: none, 1: component/source, 2: toolName

	// Try to find source in attributes or use default
	r.Attrs(func(a slog.Attr) bool {
		// Collect all attributes into Metadata (at the current nesting level)
		currentMap[a.Key] = a.Value.Any()

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
//
// attrs is the attrs.
//
// Returns the result.
func (h *BroadcastHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// Copy attributes to ensure immutability
	attrsCopy := make([]slog.Attr, len(attrs))
	copy(attrsCopy, attrs)

	// Create a modifier that adds attributes to the current map
	mod := func(m map[string]any) map[string]any {
		for _, a := range attrsCopy {
			m[a.Key] = a.Value.Any()
		}
		return m
	}

	// Copy existing modifiers and append the new one
	newModifiers := make([]func(map[string]any) map[string]any, len(h.modifiers)+1)
	copy(newModifiers, h.modifiers)
	newModifiers[len(h.modifiers)] = mod

	return &BroadcastHandler{
		broadcaster: h.broadcaster,
		level:       h.level,
		modifiers:   newModifiers,
	}
}

// WithGroup returns a new handler with the given group.
//
// name is the name of the resource.
//
// Returns the result.
func (h *BroadcastHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	// Create a modifier that creates a sub-map and returns it
	mod := func(m map[string]any) map[string]any {
		sub := make(map[string]any)
		m[name] = sub
		return sub
	}

	// Copy existing modifiers and append the new one
	newModifiers := make([]func(map[string]any) map[string]any, len(h.modifiers)+1)
	copy(newModifiers, h.modifiers)
	newModifiers[len(h.modifiers)] = mod

	return &BroadcastHandler{
		broadcaster: h.broadcaster,
		level:       h.level,
		modifiers:   newModifiers,
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
