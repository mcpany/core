// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"sync"
)

// Broadcaster manages a set of subscribers and broadcasts messages to them.
//
// Summary: manages a set of subscribers and broadcasts messages to them.
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
	GlobalBroadcaster = NewBroadcaster()
)

// NewBroadcaster creates a new Broadcaster.
//
// Returns the result.
//
// Summary: creates a new Broadcaster.
//
// Returns:
//   - *Broadcaster: The result.
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
// This is primarily for testing to ensure a clean state.
//
// Summary: clears the broadcaster history and subscribers.
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
// The channel has a small buffer to prevent slow consumers from blocking the broadcaster.
// It is the caller's responsibility to read from the channel promptly.
//
// Summary: returns a channel that will receive broadcast messages.
//
// Returns:
//   - chan: The result.
//
// Side Effects:
//   - None.
func (b *Broadcaster) Subscribe() chan any {
	return b.SubscribeBuffered(100)
}

// SubscribeBuffered returns a channel that will receive broadcast messages with a custom buffer size.
// The channel has a buffer to prevent slow consumers from blocking the broadcaster.
// It is the caller's responsibility to read from the channel promptly.
//
// Summary: returns a channel that will receive broadcast messages with a custom buffer size.
//
// Parameters:
//   - size (int): The size.
//
// Returns:
//   - chan: The result.
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

// SubscribeWithHistory returns a channel that will receive broadcast messages,
// and the current history of messages. This is atomic to ensure no messages are missed or duplicated.
//
// Summary: returns a channel that will receive broadcast messages,.
//
// Returns:
//   - chan (any): The result.
//   - []any: The result.
//
// Side Effects:
//   - None.
func (b *Broadcaster) SubscribeWithHistory() (chan any, []any) {
	return b.SubscribeWithHistoryBuffered(100)
}

// SubscribeWithHistoryBuffered returns a channel that will receive broadcast messages with a custom buffer size,
// and the current history of messages. This is atomic to ensure no messages are missed or duplicated.
//
// Summary: returns a channel that will receive broadcast messages with a custom buffer size,.
//
// Parameters:
//   - size (int): The size.
//
// Returns:
//   - chan (any): The result.
//   - []any: The result.
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
// ch is the ch.
//
// Summary: removes a subscriber channel.
//
// Parameters:
//   - ch chan (any): The ch chan.
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
// This method is non-blocking; if a subscriber's channel is full, the message is dropped for that subscriber.
//
// Summary: sends a message to all subscribers.
//
// Parameters:
//   - msg (any): The msg.
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

// GetHistory returns the current log history.
//
// Summary: returns the current log history.
//
// Returns:
//   - []any: The result.
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
// It is intended to be called at startup. Messages are NOT broadcasted to subscribers,
// as subscribers shouldn't exist yet, or shouldn't receive old history as "new" events.
//
// Summary: populates the history buffer with messages.
//
// Parameters:
//   - messages ([]any): The messages.
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
