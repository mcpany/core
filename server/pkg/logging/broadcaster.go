// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"context"
	"encoding/json"
	"sync"
)

// Broadcaster manages a set of subscribers and broadcasts messages to them.
type Broadcaster struct {
	mu          sync.RWMutex
	subscribers map[chan []byte]struct{}
	history     [][]byte
	limit       int
	store       LogStore
}

var (
	// GlobalBroadcaster is the shared broadcaster instance for logs.
	GlobalBroadcaster = NewBroadcaster()
)

// NewBroadcaster creates a new Broadcaster.
//
// Returns the result.
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		subscribers: make(map[chan []byte]struct{}),
		history:     make([][]byte, 0, 1000),
		limit:       1000,
	}
}

// SetStore sets the persistent store for the broadcaster and loads recent history.
func (b *Broadcaster) SetStore(store LogStore) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.store = store

	// Load recent history from store
	// We use a background context as this is typically called during initialization
	entries, err := store.GetRecentLogs(context.Background(), b.limit)
	if err == nil && len(entries) > 0 {
		var dbHistory [][]byte
		for _, entry := range entries {
			if data, err := json.Marshal(entry); err == nil {
				dbHistory = append(dbHistory, data)
			}
		}

		// Prepend DB history to existing in-memory history (startup logs)
		// Result: [Oldest DB Log ... Newest DB Log, Startup Log 1 ... Startup Log N]
		b.history = append(dbHistory, b.history...)

		if len(b.history) > b.limit {
			// Keep the last N logs
			start := len(b.history) - b.limit
			b.history = b.history[start:]
		}
	}
}

// BroadcastLog broadcasts a structured log entry and persists it.
func (b *Broadcaster) BroadcastLog(entry LogEntry) {
	data, err := json.Marshal(entry)
	if err != nil {
		return
	}

	// Persist asynchronously to avoid blocking
	b.mu.RLock()
	store := b.store
	b.mu.RUnlock()

	if store != nil {
		go func() {
			_ = store.SaveLog(context.Background(), entry)
		}()
	}

	b.Broadcast(data)
}

// Subscribe returns a channel that will receive broadcast messages.
// The channel has a small buffer to prevent slow consumers from blocking the broadcaster.
// It is the caller's responsibility to read from the channel promptly.
func (b *Broadcaster) Subscribe() chan []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan []byte, 100)
	b.subscribers[ch] = struct{}{}
	return ch
}

// SubscribeWithHistory returns a channel that will receive broadcast messages,
// and the current history of messages. This is atomic to ensure no messages are missed or duplicated.
func (b *Broadcaster) SubscribeWithHistory() (chan []byte, [][]byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan []byte, 100)
	b.subscribers[ch] = struct{}{}

	result := make([][]byte, len(b.history))
	for i, msg := range b.history {
		msgCopy := make([]byte, len(msg))
		copy(msgCopy, msg)
		result[i] = msgCopy
	}

	return ch, result
}

// Unsubscribe removes a subscriber channel.
//
// ch is the ch.
func (b *Broadcaster) Unsubscribe(ch chan []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.subscribers[ch]; ok {
		delete(b.subscribers, ch)
		close(ch)
	}
}

// Broadcast sends a message to all subscribers.
// This method is non-blocking; if a subscriber's channel is full, the message is dropped for that subscriber.
func (b *Broadcaster) Broadcast(msg []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Append to history
	// We make a copy of msg to ensure history persists even if caller reuses buffer
	msgCopy := make([]byte, len(msg))
	copy(msgCopy, msg)
	b.history = append(b.history, msgCopy)

	if len(b.history) > b.limit {
		// Simple slicing is efficient and sufficient.
		// Go's append will handle reallocation when capacity is reached.
		b.history = b.history[1:]
	}

	for ch := range b.subscribers {
		select {
		case ch <- msg:
		default:
			// Drop message if channel is full
		}
	}
}

// GetHistory returns the current log history.
func (b *Broadcaster) GetHistory() [][]byte {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := make([][]byte, len(b.history))
	for i, msg := range b.history {
		msgCopy := make([]byte, len(msg))
		copy(msgCopy, msg)
		result[i] = msgCopy
	}
	return result
}
