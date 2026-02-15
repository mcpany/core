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

type broadcastAction struct {
	isGroup   bool
	groupName string
	attrs     []slog.Attr
}

// BroadcastHandler implements slog.Handler and sends logs to the Broadcaster.
type BroadcastHandler struct {
	broadcaster *Broadcaster
	actions     []broadcastAction
	level       slog.Level
	mu          sync.Mutex
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

	// Helper stack to manage nested maps
	stack := []map[string]any{entry.Metadata}
	currentMap := func() map[string]any {
		return stack[len(stack)-1]
	}

	// Replay actions (groups and attributes from handler)
	for _, action := range h.actions {
		if action.isGroup {
			if action.groupName == "" {
				continue
			}
			newMap := make(map[string]any)
			curr := currentMap()
			if existing, ok := curr[action.groupName]; ok {
				if existingMap, ok := existing.(map[string]any); ok {
					newMap = existingMap
				}
			}
			curr[action.groupName] = newMap
			stack = append(stack, newMap)
		} else {
			curr := currentMap()
			for _, a := range action.attrs {
				addAttr(curr, a)
			}
		}
	}

	// Add record attributes
	r.Attrs(func(a slog.Attr) bool {
		addAttr(currentMap(), a)
		return true
	})

	// Extract source if available in Metadata
	if entry.Source == "" {
		if v, ok := entry.Metadata["toolName"]; ok {
			if s, ok := v.(string); ok {
				entry.Source = s
			}
		} else if v, ok := entry.Metadata["source"]; ok {
			if s, ok := v.(string); ok {
				entry.Source = s
			}
		} else if v, ok := entry.Metadata["component"]; ok {
			if s, ok := v.(string); ok {
				entry.Source = s
			}
		}
	}

	// Also handle source from record PC if available
	if entry.Source == "" && r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		entry.Source = f.Function
	}

	data, err := json.Marshal(entry)
	if err != nil {
		// Fallback: Try to broadcast an error entry to avoid silent loss of logs
		errorEntry := LogEntry{
			ID:        entry.ID,
			Timestamp: entry.Timestamp,
			Level:     "ERROR",
			Message:   "Failed to marshal log entry: " + err.Error(),
			Source:    "logging.BroadcastHandler",
		}
		// Best effort
		if errorData, e := json.Marshal(errorEntry); e == nil {
			h.broadcaster.Broadcast(errorData)
		}
		// Return the original error to be compliant with the Handler interface
		return err
	}

	h.broadcaster.Broadcast(data)
	return nil
}

// addAttr adds an attribute to the map, handling nested groups.
func addAttr(m map[string]any, a slog.Attr) {
	val := a.Value.Resolve()
	if val.Kind() == slog.KindGroup {
		// Group attribute
		if a.Key == "" {
			// Inline group
			for _, child := range val.Group() {
				addAttr(m, child)
			}
			return
		}

		subMap := make(map[string]any)
		if existing, ok := m[a.Key]; ok {
			if existingMap, ok := existing.(map[string]any); ok {
				subMap = existingMap
			}
		}
		m[a.Key] = subMap

		for _, child := range val.Group() {
			addAttr(subMap, child)
		}
	} else {
		m[a.Key] = val.Any()
	}
}

// WithAttrs returns a new handler with the given attributes.
//
// attrs is the attrs.
//
// Returns the result.
func (h *BroadcastHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.mu.Lock()
	defer h.mu.Unlock()

	newActions := make([]broadcastAction, len(h.actions)+1)
	copy(newActions, h.actions)
	newActions[len(h.actions)] = broadcastAction{isGroup: false, attrs: attrs}

	return &BroadcastHandler{
		broadcaster: h.broadcaster,
		actions:     newActions,
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

	newActions := make([]broadcastAction, len(h.actions)+1)
	copy(newActions, h.actions)
	newActions[len(h.actions)] = broadcastAction{isGroup: true, groupName: name}

	return &BroadcastHandler{
		broadcaster: h.broadcaster,
		actions:     newActions,
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
