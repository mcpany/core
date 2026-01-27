// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mcpany/core/server/pkg/util"
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
	attrs       []slog.Attr
	group       string
	mu          sync.Mutex
	level       slog.Level
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

// RedactingHandler is a slog.Handler that deeply redacts sensitive information
// from Any attributes by marshaling them to JSON, checking for sensitive keys,
// and then unmarshaling them back. This ensures that text handlers don't leak
// secrets hidden in structs.
type RedactingHandler struct {
	inner slog.Handler
}

// NewRedactingHandler creates a new RedactingHandler.
func NewRedactingHandler(handler slog.Handler) *RedactingHandler {
	return &RedactingHandler{inner: handler}
}

// Enabled delegates to the inner handler.
func (h *RedactingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

// Handle redacts attributes and delegates to the inner handler.
func (h *RedactingHandler) Handle(ctx context.Context, r slog.Record) error {
	// We need to create a new record with redacted attributes.
	// Since we can't easily modify the existing record's attributes in place without
	// iterating and collecting them, we do exactly that.

	newR := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
	var newAttrs []slog.Attr

	r.Attrs(func(a slog.Attr) bool {
		newAttrs = append(newAttrs, h.redactAttr(a))
		return true
	})

	newR.AddAttrs(newAttrs...)
	return h.inner.Handle(ctx, newR)
}

// WithAttrs delegates to the inner handler with redacted attributes.
func (h *RedactingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	redactedAttrs := make([]slog.Attr, 0, len(attrs))
	for _, a := range attrs {
		redactedAttrs = append(redactedAttrs, h.redactAttr(a))
	}
	return NewRedactingHandler(h.inner.WithAttrs(redactedAttrs))
}

// WithGroup delegates to the inner handler.
func (h *RedactingHandler) WithGroup(name string) slog.Handler {
	return NewRedactingHandler(h.inner.WithGroup(name))
}

func (h *RedactingHandler) redactAttr(a slog.Attr) slog.Attr {
	// First check if the key itself is sensitive (should already be handled by ReplaceAttr in Init,
	// but good to have as backup/for direct use).
	if util.IsSensitiveKey(a.Key) {
		return slog.String(a.Key, "[REDACTED]")
	}

	// If the value is complex (Any or Group), we attempt deep redaction.
	// Note: Group is usually a list of Attrs, but Any can be a struct/map.
	if a.Value.Kind() == slog.KindAny {
		val := a.Value.Any()
		if val == nil {
			return a
		}

		// Attempt to marshal to JSON to catch JSON-tagged sensitive fields.
		b, err := json.Marshal(val)
		if err == nil {
			// Redact the JSON bytes
			redacted := util.RedactJSON(b)

			// Optimization: Check if modification happened
			if !bytes.Equal(b, redacted) {
				// If redacted, we need to return a value that represents the redacted structure.
				// Unmarshalling back to interface{} gives us a map/slice with redacted values.
				var v interface{}
				if err := json.Unmarshal(redacted, &v); err == nil {
					return slog.Any(a.Key, v)
				}
				// If unmarshal fails (unlikely for valid JSON), return the redacted JSON string
				// as a fallback.
				return slog.String(a.Key, string(redacted))
			}
		}
	} else if a.Value.Kind() == slog.KindGroup {
		// For groups, we need to recurse into the group's attributes.
		// slog.Group returns an Attr with a Value of KindGroup.
		// We can get the attributes via a.Value.Group().
		attrs := a.Value.Group()
		var newGroupAttrs []slog.Attr
		for _, subAttr := range attrs {
			newGroupAttrs = append(newGroupAttrs, h.redactAttr(subAttr))
		}
		return slog.Group(a.Key, asAny(newGroupAttrs)...)
	}

	return a
}

// asAny converts []slog.Attr to []any for slog.Group.
func asAny(attrs []slog.Attr) []any {
	args := make([]any, len(attrs))
	for i, a := range attrs {
		args[i] = a
	}
	return args
}
