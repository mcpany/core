// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"sync"
)

// Broadcaster manages a set of subscribers and broadcasts messages to them.
//
// Summary: Manager for broadcasting messages to multiple subscribers.
type Broadcaster struct {
	mu          sync.RWMutex
	subscribers map[chan any]struct{}
	history     []any
	head        int
	full        bool
	limit       int
}

var (
	// GlobalBroadcaster is the shared broadcaster instance for logs.
	// Summary: Singleton instance of the log broadcaster.
	GlobalBroadcaster = NewBroadcaster()
)

// NewBroadcaster creates a new Broadcaster.
//
// Summary: Initializes NewBroadcaster operation.
//
// Returns:
//   - *Broadcaster: The new Broadcaster instance.
//
// Parameters:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		subscribers: make(map[chan any]struct{}),
		history:     make([]any, 1000),
		limit:       1000,
	}
}

// Reset clears the broadcaster history and subscribers.
//
// Summary: Executes Reset operation.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (b *Broadcaster) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers = make(map[chan any]struct{})
	b.history = make([]any, b.limit)
	b.head = 0
	b.full = false
}

// Subscribe returns a channel that will receive broadcast messages.
//
// Summary: Executes Subscribe operation.
//
// Returns:
//   - chan any: The subscription channel.
//
// Parameters:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (b *Broadcaster) Subscribe() chan any {
	return b.SubscribeBuffered(100)
}

// SubscribeBuffered returns a channel that will receive broadcast messages with a custom buffer size.
//
// Summary: Executes SubscribeBuffered operation.
//
// Parameters:
//   - size (int): The buffer size.
//
// Returns:
//   - chan any: The subscription channel.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (b *Broadcaster) SubscribeBuffered(size int) chan any {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan any, size)
	b.subscribers[ch] = struct{}{}
	return ch
}

// SubscribeWithHistory returns a channel that will receive broadcast messages, and the current history of messages.
//
// Summary: Executes SubscribeWithHistory operation.
//
// Returns:
//   - chan any: The subscription channel.
//   - []any: The history of messages.
//
// Parameters:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (b *Broadcaster) SubscribeWithHistory() (chan any, []any) {
	return b.SubscribeWithHistoryBuffered(100)
}

// SubscribeWithHistoryBuffered returns a channel that will receive broadcast messages with a custom buffer size, and the current history of messages.
//
// Summary: Executes SubscribeWithHistoryBuffered operation.
//
// Parameters:
//   - size (int): The buffer size.
//
// Returns:
//   - chan any: The subscription channel.
//   - []any: The history of messages.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (b *Broadcaster) SubscribeWithHistoryBuffered(size int) (chan any, []any) {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan any, size)
	b.subscribers[ch] = struct{}{}

	count := b.limit
	if !b.full {
		count = b.head
	}
	result := make([]any, count)

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
// Summary: Executes Unsubscribe operation.
//
// Parameters:
//   - ch (chan any): The channel to unsubscribe.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (b *Broadcaster) Unsubscribe(ch chan any) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.subscribers[ch]; ok {
		delete(b.subscribers, ch)
		close(ch)
	}
}

// Broadcast sends a message to all subscribers.
//
// Summary: Executes Broadcast operation.
//
// Parameters:
//   - msg (any): The message to broadcast.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (b *Broadcaster) Broadcast(msg any) {
	// ⚡ BOLT: Optimized Broadcast to use interface{} (any) instead of []byte.
	// Randomized Selection from Top 5 High-Impact Targets
	// This avoids unnecessary JSON marshaling when just storing in history
	// or broadcasting to subscribers who will handle serialization.
	// We assume 'msg' is safe to store (e.g. value type struct or immutable).

	b.mu.Lock()
	defer b.mu.Unlock()

	// ⚡ BOLT: Ring Buffer Optimization
	b.history[b.head] = msg
	b.head++
	if b.head >= b.limit {
		b.head = 0
		b.full = true
	}

	for ch := range b.subscribers {
		select {
		case ch <- msg:
		default:
			// Drop message if channel is full
		}
	}
}

// ClearHistory clears the history of the broadcaster without removing subscribers.
//
// Summary: Executes ClearHistory operation.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (b *Broadcaster) ClearHistory() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.history = make([]any, b.limit)
	b.head = 0
	b.full = false
}

// GetHistory returns the current log history.
//
// Summary: Retrieves GetHistory operation.
//
// Returns:
//   - []any: The history of messages.
//
// Parameters:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (b *Broadcaster) GetHistory() []any {
	b.mu.RLock()
	defer b.mu.RUnlock()

	count := b.limit
	if !b.full {
		count = b.head
	}
	result := make([]any, count)

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
//
// Summary: Executes Hydrate operation.
//
// Parameters:
//   - messages ([]any): The messages to add to history.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (b *Broadcaster) Hydrate(messages []any) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, msg := range messages {
		b.history[b.head] = msg
		b.head++
		if b.head >= b.limit {
			b.head = 0
			b.full = true
		}
	}
}
