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

// LogStore interface for persisting logs.
type LogStore interface {
	SaveLog(ctx context.Context, entry *LogEntry) error
}

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
	groups      []string
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
// name is the name of the resource.
//
// Returns the result.
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

// DBHandler implements slog.Handler and writes logs to a LogStore.
type DBHandler struct {
	store  LogStore
	attrs  []slog.Attr
	groups []string
	mu     sync.Mutex
	level  slog.Level
	ch     chan *LogEntry
}

// NewDBHandler creates a new DBHandler.
func NewDBHandler(store LogStore, level slog.Level) *DBHandler {
	h := &DBHandler{
		store: store,
		level: level,
		ch:    make(chan *LogEntry, 1000), // Buffer
	}
	// Start worker
	go h.worker()
	return h
}

// SetStore updates the store.
func (h *DBHandler) SetStore(store LogStore) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.store = store
}

func (h *DBHandler) worker() {
	for entry := range h.ch {
		h.mu.Lock()
		store := h.store
		h.mu.Unlock()

		if store != nil {
			// Use background context for async save
			_ = store.SaveLog(context.Background(), entry)
		}
	}
}

// Enabled returns true if the level is greater than or equal to the handler's level.
func (h *DBHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

// Handle handles the log record.
func (h *DBHandler) Handle(_ context.Context, r slog.Record) error {
	entry := LogEntry{
		ID:        uuid.New().String(),
		Timestamp: r.Time.UTC().Format(time.RFC3339), // Store as UTC for consistent sorting
		Level:     r.Level.String(),
		Message:   r.Message,
		Metadata:  make(map[string]any),
	}

	// Reuse logic from BroadcastHandler manually or extract it?
	// For simplicity, duplicating the attribute merging logic here.
	mergeAttr := func(root map[string]any, groups []string, a slog.Attr) {
		targetMap := root
		for _, g := range groups {
			if _, ok := targetMap[g]; !ok {
				targetMap[g] = make(map[string]any)
			}
			if m, ok := targetMap[g].(map[string]any); ok {
				targetMap = m
			} else {
				m := make(map[string]any)
				targetMap[g] = m
				targetMap = m
			}
		}
		targetMap[a.Key] = a.Value.Any()
	}

	for _, a := range h.attrs {
		mergeAttr(entry.Metadata, nil, a)
	}

	r.Attrs(func(a slog.Attr) bool {
		mergeAttr(entry.Metadata, h.groups, a)
		return true
	})

	if tool, ok := entry.Metadata["toolName"].(string); ok && tool != "" {
		entry.Source = tool
	} else if src, ok := entry.Metadata["source"].(string); ok && src != "" {
		entry.Source = src
	} else if comp, ok := entry.Metadata["component"].(string); ok && comp != "" {
		entry.Source = comp
	}

	if entry.Source == "" && r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		entry.Source = f.Function
	}

	// Non-blocking send
	select {
	case h.ch <- &entry:
	default:
		// Drop if full
	}
	return nil
}

// WithAttrs returns a new handler with the given attributes.
func (h *DBHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.mu.Lock()
	defer h.mu.Unlock()

	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)

	// Note: We return a new handler but it shares the channel and worker!
	// Wait, if we return a new handler, Handle() will be called on the new handler.
	// We want to reuse the worker/channel.
	// So we pass the same `ch` and `store` (by ref?).
	// `store` can be updated on the original handler, but clones need to see it too?
	// If `store` is updated in `SetStore` on the *original* handler, how do clones see it?
	// `SetStore` should likely update a shared state object or the clones must share the pointer to store wrapper.
	// But `store` is an interface.
	// Actually `NewDBHandler` returns `*DBHandler`.
	// `WithAttrs` returns `slog.Handler` (interface).
	// If `WithAttrs` creates a NEW struct `&DBHandler{...}`, `SetStore` on the OLD struct won't affect new ones unless they share state.
	// The `ch` is shared (reference type). The `worker` reads from `ch`.
	// The `worker` is running on the original handler context? No, `go h.worker()` binds to `h`.
	// If `WithAttrs` creates a NEW `DBHandler` and sends to `h.ch`, the worker (bound to original `h`) reads it.
	// The worker uses `h.store`.
	// So as long as we send to the SAME channel, and the worker (on the original instance) processes it, it works!
	// And the new handler doesn't need a worker.
	// BUT `Handle` on the NEW handler sends to `ch`.
	// So we just need to copy `ch` reference.
	// What about `store`? The new handler's `Handle` method doesn't use `store` directly, it sends to `ch`.
	// Correct. `Handle` uses `ch`. `worker` uses `store`.
	// So we only need to ensure `worker` is running and has access to `store`.
	// The worker is started in `NewDBHandler`.
	// `WithAttrs` creates a new `DBHandler` sharing `ch`. It does NOT start a new worker.
	// The `store` field in the NEW handler is irrelevant if `Handle` only writes to `ch`.
	// However, `SetStore` updates `h.store`.
	// We only call `SetStore` on the global handler (root).
	// The worker is attached to the root handler (via closure or method receiver).
	// Yes, `go h.worker()` captures the `h` pointer.
	// So updating `h.store` works.

	return &DBHandler{
		store:  h.store, // Copied but unused in Handle
		ch:     h.ch,    // Shared channel
		attrs:  newAttrs,
		groups: h.groups,
		level:  h.level,
	}
}

// WithGroup returns a new handler with the given group.
func (h *DBHandler) WithGroup(name string) slog.Handler {
	h.mu.Lock()
	defer h.mu.Unlock()

	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = name

	return &DBHandler{
		store:  h.store,
		ch:     h.ch,
		attrs:  h.attrs,
		groups: newGroups,
		level:  h.level,
	}
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
