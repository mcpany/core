// Copyright 2025 Author(s) of MCP Any
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

// LogEntry is the structure for logs sent over WebSocket.
// It matches the frontend expectation.
//
// Summary: Data structure for log entries sent to the frontend.
type LogEntry struct {
	ID        string         `json:"id"`
	Timestamp string         `json:"timestamp"`
	Level     string         `json:"level"`
	Message   string         `json:"message"`
	Source    string         `json:"source,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// BroadcastHandler implements slog.Handler and sends logs to the Broadcaster.
//
// Summary: slog.Handler implementation that broadcasts log records.
type BroadcastHandler struct {
	broadcaster *Broadcaster
	attrs       []slog.Attr
	groups      []string
	mu          sync.Mutex
	level       slog.Leveler
}

// NewBroadcastHandler creates a new BroadcastHandler.
//
// Summary: Initializes NewBroadcastHandler operation.
//
// Parameters:
//   - broadcaster: *Broadcaster. The log broadcaster.
//   - level: slog.Leveler. The minimum log level to broadcast.
//
// Returns:
//   - *BroadcastHandler: The initialized broadcast handler.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewBroadcastHandler(broadcaster *Broadcaster, level slog.Leveler) *BroadcastHandler {
	return &BroadcastHandler{
		broadcaster: broadcaster,
		level:       level,
	}
}

// Enabled returns true if the level is greater than or equal to the handler's level.
//
// Summary: Executes Enabled operation.
//
// Parameters:
//   - ctx: context.Context. The context of the log.
//   - level: slog.Level. The log level.
//
// Returns:
//   - bool: True if the level is enabled, false otherwise.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (h *BroadcastHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

// Handle handles the log record by converting it to LogEntry and broadcasting it.
//
// Summary: Executes Handle operation.
//
// Parameters:
//   - ctx: context.Context. The context of the log.
//   - r: slog.Record. The log record.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (h *BroadcastHandler) Handle(_ context.Context, r slog.Record) error {
	entry := LogEntry{
		ID:        uuid.New().String(),
		Timestamp: r.Time.Format(time.RFC3339),
		Level:     r.Level.String(),
		Message:   r.Message,
		Metadata:  make(map[string]any),
	}

	// Helper to merge attribute into metadata, respecting groups
	mergeAttr := func(root map[string]any, groups []string, a slog.Attr) {
		targetMap := root

		// Navigate/Create nested maps for groups
		for _, g := range groups {
			if _, ok := targetMap[g]; !ok {
				targetMap[g] = make(map[string]any)
			}
			if m, ok := targetMap[g].(map[string]any); ok {
				targetMap = m
			} else {
				// Conflict: overwrite with new map
				m := make(map[string]any)
				targetMap[g] = m
				targetMap = m
			}
		}

		targetMap[a.Key] = a.Value.Any()
	}

	// 1. Process attributes from WithAttrs
	// Note: We currently do not apply h.groups to h.attrs because of the implementation structure.
	// h.attrs are treated as root-level or pre-scoped attributes.
	// This ensures 'WithAttrs' data is at least visible.
	for _, a := range h.attrs {
		// We pass nil for groups here to keep them at root (or as they were added)
		// If we wanted to support 'WithGroup().WithAttrs()', we'd need to associate groups with attrs at creation time.
		mergeAttr(entry.Metadata, nil, a)
	}

	// 2. Process attributes from Record
	r.Attrs(func(a slog.Attr) bool {
		mergeAttr(entry.Metadata, h.groups, a)
		return true
	})

	// Source detection
	// We prioritize source found at root level.
	// Priority: toolName (2) > source (1) > component (1)
	if tool, ok := entry.Metadata["toolName"].(string); ok && tool != "" {
		entry.Source = tool
	} else if src, ok := entry.Metadata["source"].(string); ok && src != "" {
		entry.Source = src
	} else if comp, ok := entry.Metadata["component"].(string); ok && comp != "" {
		entry.Source = comp
	}

	// Also handle source from record PC if available and not yet found
	if entry.Source == "" && r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		entry.Source = f.Function
	}

	// ⚡ BOLT: Avoid unnecessary JSON marshaling.
	// Randomized Selection from Top 5 High-Impact Targets
	// We pass the struct directly. The Broadcaster stores it efficiently,
	// and subscribers (WebSocket handler) will marshal it when writing to the wire.
	h.broadcaster.Broadcast(entry)
	return nil
}

// WithAttrs returns a new handler with the given attributes.
//
// Summary: Executes WithAttrs operation.
//
// Parameters:
//   - attrs: []slog.Attr. The attributes to add to the handler.
//
// Returns:
//   - slog.Handler: The new handler with the attributes.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (h *BroadcastHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.mu.Lock()
	defer h.mu.Unlock()

	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)

	return &BroadcastHandler{
		broadcaster: h.broadcaster,
		attrs:       newAttrs,
		groups:      h.groups,
		level:       h.level,
	}
}

// WithGroup returns a new handler with the given group.
//
// Summary: Executes WithGroup operation.
//
// Parameters:
//   - name: string. The name of the group.
//
// Returns:
//   - slog.Handler: The new handler with the group.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (h *BroadcastHandler) WithGroup(name string) slog.Handler {
	h.mu.Lock()
	defer h.mu.Unlock()

	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = name

	return &BroadcastHandler{
		broadcaster: h.broadcaster,
		attrs:       h.attrs,
		groups:      newGroups,
		level:       h.level,
	}
}

// TeeHandler is a slog.Handler that writes to multiple handlers.
//
// Summary: slog.Handler implementation that multiplexes log records.
type TeeHandler struct {
	handlers []slog.Handler
}

// NewTeeHandler creates a new TeeHandler.
//
// Summary: Initializes NewTeeHandler operation.
//
// Parameters:
//   - handlers: ...slog.Handler. The handlers to tee to.
//
// Returns:
//   - *TeeHandler: The initialized tee handler.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewTeeHandler(handlers ...slog.Handler) *TeeHandler {
	return &TeeHandler{handlers: handlers}
}

// Enabled returns true if any of the handlers are enabled.
//
// Summary: Executes Enabled operation.
//
// Parameters:
//   - ctx: context.Context. The context of the log.
//   - level: slog.Level. The log level.
//
// Returns:
//   - bool: True if any handler is enabled for the level, false otherwise.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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
// Summary: Executes Handle operation.
//
// Parameters:
//   - ctx: context.Context. The context of the log.
//   - r: slog.Record. The log record.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
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
// Summary: Executes WithAttrs operation.
//
// Parameters:
//   - attrs: []slog.Attr. The attributes to add to the handlers.
//
// Returns:
//   - slog.Handler: The new tee handler with the attributes.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (h *TeeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}
	return NewTeeHandler(handlers...)
}

// WithGroup returns a new TeeHandler with the group applied to all handlers.
//
// Summary: Executes WithGroup operation.
//
// Parameters:
//   - name: string. The name of the group.
//
// Returns:
//   - slog.Handler: The new tee handler with the group.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (h *TeeHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithGroup(name)
	}
	return NewTeeHandler(handlers...)
}
