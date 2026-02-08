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
// Summary: creates a new Broadcaster.
//
// Parameters:
//   None.
//
// Returns:
//   - *Broadcaster: The *Broadcaster.
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		subscribers: make(map[chan []byte]struct{}),
		history:     make([][]byte, 1000),
		limit:       1000,
	}
}

// Subscribe returns a channel that will receive broadcast messages.
//
// Summary: returns a channel that will receive broadcast messages.
//
// Parameters:
//   None.
//
// Returns:
//   - chan []byte: The chan []byte.
func (b *Broadcaster) Subscribe() chan []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan []byte, 100)
	b.subscribers[ch] = struct{}{}
	return ch
}

// SubscribeWithHistory returns a channel that will receive broadcast messages,.
//
// Summary: returns a channel that will receive broadcast messages,.
//
// Parameters:
//   None.
//
// Returns:
//   - chan []byte: The chan []byte.
//   - [][]byte: The [][]byte.
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
// Summary: removes a subscriber channel.
//
// Parameters:
//   - ch: chan []byte. The ch.
//
// Returns:
//   None.
func (b *Broadcaster) Unsubscribe(ch chan []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.subscribers[ch]; ok {
		delete(b.subscribers, ch)
		close(ch)
	}
}

// Broadcast sends a message to all subscribers.
//
// Summary: sends a message to all subscribers.
//
// Parameters:
//   - msg: []byte. The msg.
//
// Returns:
//   None.
func (b *Broadcaster) Broadcast(msg []byte) {
	// We make a copy of msg to ensure history persists even if caller reuses buffer.
	// Doing this outside the lock reduces contention.
	msgCopy := make([]byte, len(msg))
	copy(msgCopy, msg)

	b.mu.Lock()
	defer b.mu.Unlock()

	// ⚡ BOLT: Ring Buffer Optimization
	// Randomized Selection from Top 5 High-Impact Targets
	b.history[b.head] = msgCopy
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
// Parameters:
//   None.
//
// Returns:
//   - [][]byte: The [][]byte.
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
//
// Summary: populates the history buffer with messages.
//
// Parameters:
//   - messages: [][]byte. The messages.
//
// Returns:
//   None.
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
