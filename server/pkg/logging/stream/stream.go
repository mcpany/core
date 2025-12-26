// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package stream

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// LogEntry represents a structured log entry for streaming.
type LogEntry struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Source    string `json:"source,omitempty"`
}

// Broadcaster manages WebSocket connections and broadcasts log entries.
type Broadcaster struct {
	mu          sync.RWMutex
	subscribers map[chan<- LogEntry]struct{}
}

var (
	globalBroadcaster *Broadcaster
	once              sync.Once
)

// GetBroadcaster returns the singleton Broadcaster instance.
func GetBroadcaster() *Broadcaster {
	once.Do(func() {
		globalBroadcaster = &Broadcaster{
			subscribers: make(map[chan<- LogEntry]struct{}),
		}
	})
	return globalBroadcaster
}

// Subscribe adds a channel to the subscribers list.
func (b *Broadcaster) Subscribe() chan LogEntry {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan LogEntry, 100) // Buffer to prevent blocking
	b.subscribers[ch] = struct{}{}
	return ch
}

// Unsubscribe removes a channel from the subscribers list.
func (b *Broadcaster) Unsubscribe(ch chan LogEntry) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.subscribers, ch)
	close(ch)
}

// Broadcast sends a log entry to all subscribers.
func (b *Broadcaster) Broadcast(record slog.Record) {
	entry := LogEntry{
		ID:        time.Now().String(), // Simple ID
		Timestamp: record.Time.Format(time.RFC3339),
		Level:     record.Level.String(),
		Message:   record.Message,
	}

	// Extract attributes if needed, or source
	record.Attrs(func(a slog.Attr) bool {
		if a.Key == "source" {
			entry.Source = a.Value.String()
			return false // Stop after finding source
		}
		return true
	})

	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.subscribers {
		select {
		case ch <- entry:
		default:
			// Drop message if channel is full to prevent blocking
		}
	}
}

// SlogHandler is a wrapper around slog.Handler that also broadcasts logs.
type SlogHandler struct {
	slog.Handler
	broadcaster *Broadcaster
}

// NewSlogHandler creates a new SlogHandler.
func NewSlogHandler(handler slog.Handler, broadcaster *Broadcaster) *SlogHandler {
	return &SlogHandler{
		Handler:     handler,
		broadcaster: broadcaster,
	}
}

// Handle handles the record by passing it to the underlying handler and broadcasting it.
func (h *SlogHandler) Handle(ctx context.Context, r slog.Record) error {
	// Broadcast first (non-blocking)
	h.broadcaster.Broadcast(r)
	return h.Handler.Handle(ctx, r)
}
