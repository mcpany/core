// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"sync"
)

// Broadcaster represents the public Broadcaster entity.
//
// Summary: Defines the structured data model representing a .
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

// NewBroadcaster serves as a public interface for interacting with NewBroadcaster.
//
// Summary: Constructs and returns an initialized broadcaster ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		subscribers: make(map[chan any]struct{}),
		history:     make([]any, 1000),
		limit:       1000,
	}
}

// Reset serves as a public interface for interacting with Reset.
//
// Summary: Reset the  appropriately based on current system conditions.
//
// Parameters:
//   - None.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (b *Broadcaster) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers = make(map[chan any]struct{})
	b.history = make([]any, b.limit)
	b.head = 0
	b.full = false
}

// Subscribe serves as a public interface for interacting with Subscribe.
//
// Summary: Subscribe the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (b *Broadcaster) Subscribe() chan any {
	return b.SubscribeBuffered(100)
}

// SubscribeBuffered serves as a public interface for interacting with SubscribeBuffered.
//
// Summary: Subscribe the buffered appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (b *Broadcaster) SubscribeBuffered(size int) chan any {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan any, size)
	b.subscribers[ch] = struct{}{}
	return ch
}

// SubscribeWithHistory serves as a public interface for interacting with SubscribeWithHistory.
//
// Summary: Subscribe the with history appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (b *Broadcaster) SubscribeWithHistory() (chan any, []any) {
	return b.SubscribeWithHistoryBuffered(100)
}

// SubscribeWithHistoryBuffered serves as a public interface for interacting with SubscribeWithHistoryBuffered.
//
// Summary: Subscribe the with history buffered appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// Unsubscribe serves as a public interface for interacting with Unsubscribe.
//
// Summary: Unsubscribe the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (b *Broadcaster) Unsubscribe(ch chan any) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.subscribers[ch]; ok {
		delete(b.subscribers, ch)
		close(ch)
	}
}

// Broadcast serves as a public interface for interacting with Broadcast.
//
// Summary: Broadcast the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// ClearHistory serves as a public interface for interacting with ClearHistory.
//
// Summary: Clear the history appropriately based on current system conditions.
//
// Parameters:
//   - None.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (b *Broadcaster) ClearHistory() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.history = make([]any, b.limit)
	b.head = 0
	b.full = false
}

// GetHistory serves as a public interface for interacting with GetHistory.
//
// Summary: Fetches and returns the underlying history from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// Hydrate serves as a public interface for interacting with Hydrate.
//
// Summary: Hydrate the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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
