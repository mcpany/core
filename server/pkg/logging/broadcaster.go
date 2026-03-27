// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"sync"
)

// Broadcaster manages a set of subscribers and broadcasts messages to them.
//
// Summary. Represents a Broadcaster.
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

// NewBroadcaster provides newbroadcaster functionality.
//
// Summary: NewBroadcaster.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		subscribers: make(map[chan any]struct{}),
		history:     make([]any, 1000),
		limit:       1000,
	}
}

// Reset provides reset functionality.
//
// Summary: Reset.
//
// Parameters.
//   - None.
//
// Returns.
//   - None.
func (b *Broadcaster) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers = make(map[chan any]struct{})
	b.history = make([]any, b.limit)
	b.head = 0
	b.full = false
}

// Subscribe provides subscribe functionality.
//
// Summary: Subscribe.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (b *Broadcaster) Subscribe() chan any {
	return b.SubscribeBuffered(100)
}

// SubscribeBuffered provides subscribebuffered functionality.
//
// Summary: SubscribeBuffered.
//
// Parameters.
//   - size: The parameter.
//
// Returns.
//   - result: The result.
func (b *Broadcaster) SubscribeBuffered(size int) chan any {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan any, size)
	b.subscribers[ch] = struct{}{}
	return ch
}

// SubscribeWithHistory provides subscribewithhistory functionality.
//
// Summary: SubscribeWithHistory.
//
// Parameters.
//   - ): The parameter.
//   - []any: The parameter.
//
// Returns.
//   - None.
func (b *Broadcaster) SubscribeWithHistory() (chan any, []any) {
	return b.SubscribeWithHistoryBuffered(100)
}

// SubscribeWithHistoryBuffered provides subscribewithhistorybuffered functionality.
//
// Summary: SubscribeWithHistoryBuffered.
//
// Parameters.
//   - size: The parameter.
//   - []any: The parameter.
//
// Returns.
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

// Unsubscribe provides unsubscribe functionality.
//
// Summary: Unsubscribe.
//
// Parameters.
//   - ch: The parameter.
//
// Returns.
//   - None.
func (b *Broadcaster) Unsubscribe(ch chan any) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.subscribers[ch]; ok {
		delete(b.subscribers, ch)
		close(ch)
	}
}

// Broadcast provides broadcast functionality.
//
// Summary: Broadcast.
//
// Parameters.
//   - msg: The parameter.
//
// Returns.
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

// ClearHistory provides clearhistory functionality.
//
// Summary: ClearHistory.
//
// Parameters.
//   - None.
//
// Returns.
//   - None.
func (b *Broadcaster) ClearHistory() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.history = make([]any, b.limit)
	b.head = 0
	b.full = false
}

// GetHistory provides gethistory functionality.
//
// Summary: GetHistory.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
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

// Hydrate provides hydrate functionality.
//
// Summary: Hydrate.
//
// Parameters.
//   - messages: The parameter.
//
// Returns.
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
