// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"sync"
)

// Broadcaster manages a set of subscribers and broadcasts messages to them.
type Broadcaster struct {
	mu          sync.RWMutex
	subscribers map[chan []byte]struct{}
	history     [][]byte
	limit       int
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

// Subscribe returns a channel that will receive broadcast messages.
// The channel has a small buffer to prevent slow consumers from blocking the broadcaster.
// It is the caller's responsibility to read from the channel promptly.
func (b *Broadcaster) Subscribe() chan []byte {
	ch, _ := b.SubscribeWithHistory()
	return ch
}

// SubscribeWithHistory returns a channel that will receive broadcast messages
// and the current history of messages atomically.
func (b *Broadcaster) SubscribeWithHistory() (chan []byte, [][]byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan []byte, 100)
	b.subscribers[ch] = struct{}{}

	// Copy history
	history := make([][]byte, len(b.history))
	copy(history, b.history)

	return ch, history
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
	// Make a copy of the message to ensure it persists correctly
	msgCopy := make([]byte, len(msg))
	copy(msgCopy, msg)

	b.history = append(b.history, msgCopy)
	if len(b.history) > b.limit {
		// Optimization: instead of slicing which might keep underlying array growing or holding refs,
		// since we are just appending, slicing [1:] is fine. Go GC handles this well.
		// For very high throughput, a ring buffer is better, but for logs [1:] is standard.
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

// GetHistory returns the current history of broadcasted messages.
//
// Returns the history.
func (b *Broadcaster) GetHistory() [][]byte {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Return a copy of the slice to avoid race conditions if the caller modifies it
	// (though they get byte slices which are refs, but we assume read-only usage or we'd copy those too)
	// We only copy the outer slice. The inner byte slices are considered immutable logs.
	history := make([][]byte, len(b.history))
	copy(history, b.history)
	return history
}
