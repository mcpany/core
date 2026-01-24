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

// GetHistory returns the current log history.
func (b *Broadcaster) GetHistory() [][]byte {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := make([][]byte, len(b.history))
	copy(result, b.history)
	return result
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
// and the current history of messages.
func (b *Broadcaster) SubscribeWithHistory() (chan []byte, [][]byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan []byte, 100)
	b.subscribers[ch] = struct{}{}

	result := make([][]byte, len(b.history))
	copy(result, b.history)

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
	// We make a copy of the message to ensure it persists safely
	msgCopy := make([]byte, len(msg))
	copy(msgCopy, msg)
	b.history = append(b.history, msgCopy)

	if len(b.history) > b.limit {
		// Truncate efficiently
		// We could use a ring buffer, but for 1000 items slice operation is fine.
		// To avoid memory leak with slice underlying array, we could re-allocate if capacity grows too much,
		// but Go's append usually handles this reasonably well for small buffers.
		b.history = b.history[len(b.history)-b.limit:]
	}

	for ch := range b.subscribers {
		select {
		case ch <- msg:
		default:
			// Drop message if channel is full
		}
	}
}
