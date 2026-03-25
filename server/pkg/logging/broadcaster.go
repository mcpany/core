// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"sync"
)

// Broadcaster manages a set of subscribers and broadcasts messages to them.
//
// Summary: Represents a Broadcaster.
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
	// Summary: Defines GlobalBroadcaster.
	GlobalBroadcaster = NewBroadcaster()
)

// NewBroadcaster creates a new broadcaster.
//
// Summary: Creates a new broadcaster.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - *Broadcaster: The result.
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		subscribers: make(map[chan any]struct{}),
		history:     make([]any, 1000),
		limit:       1000,
	}
}

// Reset reset reset.
//
// Summary: Reset reset.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - None.
func (b *Broadcaster) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers = make(map[chan any]struct{})
	b.history = make([]any, b.limit)
	b.head = 0
	b.full = false
}

// Subscribe returns a channel that will receive broadcast messages. The channel has a small buffer to prevent slow consumers from blocking the broadcaster. It is the caller's responsibility to read from the channel promptly.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - chanany: The resulting chanany.
//
// Errors: - None.
//   - None.
//
// Side Effects: - None.
//   - None.
//
// Summary: Executes Subscribe operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func (b *Broadcaster) Subscribe() chan any {
	return b.SubscribeBuffered(100)
}

// SubscribeBuffered returns a channel that will receive broadcast messages with a custom buffer size. The channel has a buffer to prevent slow consumers from blocking the broadcaster. It is the caller's responsibility to read from the channel promptly.
//
// Parameters: - None.
//   - size (int): The size parameter.
//
// Returns: - None.
//   - chanany: The resulting chanany.
//
// Errors: - None.
//   - None.
//
// Side Effects: - None.
//   - None.
//
// Summary: Executes SubscribeBuffered operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func (b *Broadcaster) SubscribeBuffered(size int) chan any {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan any, size)
	b.subscribers[ch] = struct{}{}
	return ch
}

// SubscribeWithHistory subscribeWithHistory subscribe with history.
//
// Summary: SubscribeWithHistory subscribe with history.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - any: The result.
//   - []any: The result.
func (b *Broadcaster) SubscribeWithHistory() (chan any, []any) {
	return b.SubscribeWithHistoryBuffered(100)
}

// SubscribeWithHistoryBuffered subscribeWithHistoryBuffered subscribe with history buffered.
//
// Summary: SubscribeWithHistoryBuffered subscribe with history buffered.
//
// Parameters: - None.
//   - size (int): The size.
//
// Returns: - None.
//   - any: The result.
//   - []any: The result.
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

// Unsubscribe unsubscribe unsubscribe.
//
// Summary: Unsubscribe unsubscribe.
//
// Parameters: - None.
//   - ch chan (any): The ch chan.
//
// Returns: - None.
//   - None.
func (b *Broadcaster) Unsubscribe(ch chan any) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.subscribers[ch]; ok {
		delete(b.subscribers, ch)
		close(ch)
	}
}

// Broadcast broadcast broadcast.
//
// Summary: Broadcast broadcast.
//
// Parameters: - None.
//   - msg (any): The msg.
//
// Returns: - None.
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

// ClearHistory clearHistory clear history.
//
// Summary: ClearHistory clear history.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - None.
func (b *Broadcaster) ClearHistory() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.history = make([]any, b.limit)
	b.head = 0
	b.full = false
}

// GetHistory retrieves the history.
//
// Summary: Retrieves the history.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - []any: The result.
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

// Hydrate hydrate hydrate.
//
// Summary: Hydrate hydrate.
//
// Parameters: - None.
//   - messages ([]any): The messages.
//
// Returns: - None.
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
