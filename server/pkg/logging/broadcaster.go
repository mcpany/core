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
	head        int
	full        bool
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
		history:     make([][]byte, 1000),
		limit:       1000,
	}
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

	count := b.limit
	if !b.full {
		count = b.head
	}
	result := make([][]byte, count)

	idx := 0
	if b.full {
		idx = b.head
	}

	for i := 0; i < count; i++ {
		// ⚡ BOLT: Zero-copy optimization.
		// We share the slice reference because b.history entries are immutable (only replaced).
		// This avoids allocations and copies under lock.
		// Randomized Selection from Top 5 High-Impact Targets
		result[i] = b.history[idx]

		idx++
		if idx >= b.limit {
			idx = 0
		}
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
		// ⚡ BOLT: Do not close channel here.
		// Since Broadcast releases the lock before sending, there is a race where Unsubscribe
		// could close the channel while Broadcast is trying to send to it, causing a panic.
		// We rely on the subscriber to stop reading or the channel to be GC'd.
		// close(ch)
	}
}

// Broadcast sends a message to all subscribers.
// This method is non-blocking; if a subscriber's channel is full, the message is dropped for that subscriber.
func (b *Broadcaster) Broadcast(msg []byte) {
	// We make a copy of msg to ensure history persists even if caller reuses buffer.
	// Doing this outside the lock reduces contention.
	msgCopy := make([]byte, len(msg))
	copy(msgCopy, msg)

	b.mu.Lock()

	// ⚡ BOLT: Ring Buffer Optimization
	// Randomized Selection from Top 5 High-Impact Targets
	b.history[b.head] = msgCopy
	b.head++
	if b.head >= b.limit {
		b.head = 0
		b.full = true
	}

	// ⚡ BOLT: Lock Contention Optimization
	// Snapshot subscribers to release lock early, preventing blocking on channel operations.
	// Randomized Selection from Top 5 High-Impact Targets
	subs := make([]chan []byte, 0, len(b.subscribers))
	for ch := range b.subscribers {
		subs = append(subs, ch)
	}
	b.mu.Unlock()

	for _, ch := range subs {
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

	count := b.limit
	if !b.full {
		count = b.head
	}
	result := make([][]byte, count)

	idx := 0
	if b.full {
		idx = b.head
	}

	for i := 0; i < count; i++ {
		// ⚡ BOLT: Zero-copy optimization.
		result[i] = b.history[idx]

		idx++
		if idx >= b.limit {
			idx = 0
		}
	}
	return result
}

// Hydrate populates the history buffer with messages.
// It is intended to be called at startup. Messages are NOT broadcasted to subscribers,
// as subscribers shouldn't exist yet, or shouldn't receive old history as "new" events.
func (b *Broadcaster) Hydrate(messages [][]byte) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, msg := range messages {
		// Copy msg to ensure ownership
		msgCopy := make([]byte, len(msg))
		copy(msgCopy, msg)

		b.history[b.head] = msgCopy
		b.head++
		if b.head >= b.limit {
			b.head = 0
			b.full = true
		}
	}
}
