// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"sync"
)

// Broadcaster manages a set of subscribers and broadcasts messages to them.
//
// Summary: Broadcaster manages a set of subscribers and broadcasts messages to them.
//
// Summary: Broadcaster manages a set of subscribers and broadcasts messages to them.
type Broadcaster struct {
	mu          sync.RWMutex
	subscribers map[chan any]struct{}
	history     []any
	head        int
	full        bool
	limit       int
}
// GlobalBroadcaster is the shared broadcaster instance for logs.
//
// Summary: GlobalBroadcaster is the shared broadcaster instance for logs.
var (
// GlobalBroadcaster is the shared broadcaster instance for logs.
//
// NewBroadcaster creates a new Broadcaster. Returns the result.
//
// Summary: NewBroadcaster creates a new Broadcaster. Returns the result.
//
// Parameters:
//   - None.
//
// Returns:
//   - *Broadcaster: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
// Reset clears the broadcaster history and subscribers. This is primarily for testing to ensure a clean state.
//
// Summary: Reset clears the broadcaster history and subscribers. This is primarily for testing to ensure a clean state.
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
//   - May modify internal state or perform external network calls.
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (b *Broadcaster) Reset() {
// Subscribe returns a channel that will receive broadcast messages. The channel has a small buffer to prevent slow consumers from blocking the broadcaster. It is the caller's responsibility to read from the channel promptly.
//
// Summary: Subscribe returns a channel that will receive broadcast messages. The channel has a small buffer to prevent slow consumers from blocking the broadcaster. It is the caller's responsibility to read from the channel promptly.
//
// Parameters:
//   - None.
//
// Returns:
//   - chan any: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//   - None.
//
// Returns:
//   - chan any: The resulting object or data structure.
// SubscribeBuffered returns a channel that will receive broadcast messages with a custom buffer size. The channel has a buffer to prevent slow consumers from blocking the broadcaster. It is the caller's responsibility to read from the channel promptly.
//
// Summary: SubscribeBuffered returns a channel that will receive broadcast messages with a custom buffer size. The channel has a buffer to prevent slow consumers from blocking the broadcaster. It is the caller's responsibility to read from the channel promptly.
//
// Parameters:
//   - size (int): The numeric value for size.
//
// Returns:
//   - chan any: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//
// Parameters:
//   - size (int): The numeric value for size.
//
// Returns:
//   - chan any: The resulting object or data structure.
//
// Errors:
// SubscribeWithHistory returns a channel that will receive broadcast messages, and the current history of messages. This is atomic to ensure no messages are missed or duplicated.
//
// Summary: SubscribeWithHistory returns a channel that will receive broadcast messages, and the current history of messages. This is atomic to ensure no messages are missed or duplicated.
//
// Parameters:
//   - None.
//
// Returns:
//   - chan any: The resulting object or data structure.
//   - []any: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
// Summary: SubscribeWithHistory returns a channel that will receive broadcast messages, and the current history of messages. This is atomic to ensure no messages are missed or duplicated.
//
// Parameters:
//   - None.
// SubscribeWithHistoryBuffered returns a channel that will receive broadcast messages with a custom buffer size, and the current history of messages. This is atomic to ensure no messages are missed or duplicated.
//
// Summary: SubscribeWithHistoryBuffered returns a channel that will receive broadcast messages with a custom buffer size, and the current history of messages. This is atomic to ensure no messages are missed or duplicated.
//
// Parameters:
//   - size (int): The numeric value for size.
//
// Returns:
//   - chan any: The resulting object or data structure.
//   - []any: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
// SubscribeWithHistoryBuffered returns a channel that will receive broadcast messages with a custom buffer size, and the current history of messages. This is atomic to ensure no messages are missed or duplicated.
//
// Summary: SubscribeWithHistoryBuffered returns a channel that will receive broadcast messages with a custom buffer size, and the current history of messages. This is atomic to ensure no messages are missed or duplicated.
//
// Parameters:
//   - size (int): The numeric value for size.
//
// Returns:
//   - chan any: The resulting object or data structure.
//   - []any: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
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

// Unsubscribe removes a subscriber channel. ch is the ch.
//
// Summary: Unsubscribe removes a subscriber channel. ch is the ch.
//
// Parameters:
//   - ch (chan any): The provided ch data.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
	return ch, result
}

// Unsubscribe removes a subscriber channel. ch is the ch.
//
// Summary: Unsubscribe removes a subscriber channel. ch is the ch.
//
// Parameters:
//   - ch (chan any): The provided ch data.
// Broadcast sends a message to all subscribers. This method is non-blocking; if a subscriber's channel is full, the message is dropped for that subscriber.
//
// Summary: Broadcast sends a message to all subscribers. This method is non-blocking; if a subscriber's channel is full, the message is dropped for that subscriber.
//
// Parameters:
//   - msg (any): The provided msg data.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
		delete(b.subscribers, ch)
		close(ch)
	}
}

// Broadcast sends a message to all subscribers. This method is non-blocking; if a subscriber's channel is full, the message is dropped for that subscriber.
//
// Summary: Broadcast sends a message to all subscribers. This method is non-blocking; if a subscriber's channel is full, the message is dropped for that subscriber.
//
// Parameters:
//   - msg (any): The provided msg data.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (b *Broadcaster) Broadcast(msg any) {
	// ⚡ BOLT: Optimized Broadcast to use interface{} (any) instead of []byte.
	// Randomized Selection from Top 5 High-Impact Targets
	// This avoids unnecessary JSON marshaling when just storing in history
	// or broadcasting to subscribers who will handle serialization.
	// We assume 'msg' is safe to store (e.g. value type struct or immutable).

// GetHistory returns the current log history.
//
// Summary: GetHistory returns the current log history.
//
// Parameters:
//   - None.
//
// Returns:
//   - []any: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
		case ch <- msg:
		default:
			// Drop message if channel is full
		}
	}
}

// GetHistory returns the current log history.
//
// Summary: GetHistory returns the current log history.
//
// Parameters:
//   - None.
//
// Returns:
//   - []any: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (b *Broadcaster) GetHistory() []any {
	b.mu.RLock()
	defer b.mu.RUnlock()

	count := b.limit
// Hydrate populates the history buffer with messages. It is intended to be called at startup. Messages are NOT broadcasted to subscribers, as subscribers shouldn't exist yet, or shouldn't receive old history as "new" events.
//
// Summary: Hydrate populates the history buffer with messages. It is intended to be called at startup. Messages are NOT broadcasted to subscribers, as subscribers shouldn't exist yet, or shouldn't receive old history as "new" events.
//
// Parameters:
//   - messages ([]any): The provided messages data.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.

		idx++
		if idx >= b.limit {
			idx = 0
		}
	}
	return result
}

// Hydrate populates the history buffer with messages. It is intended to be called at startup. Messages are NOT broadcasted to subscribers, as subscribers shouldn't exist yet, or shouldn't receive old history as "new" events.
//
// Summary: Hydrate populates the history buffer with messages. It is intended to be called at startup. Messages are NOT broadcasted to subscribers, as subscribers shouldn't exist yet, or shouldn't receive old history as "new" events.
//
// Parameters:
//   - messages ([]any): The provided messages data.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
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
