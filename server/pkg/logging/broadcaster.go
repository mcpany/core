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

	result := make([][]byte, len(b.history))
	for i, msg := range b.history {
		msgCopy := make([]byte, len(msg))
		copy(msgCopy, msg)
		result[i] = msgCopy
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
		close(ch)
	}
}

// Broadcast sends a message to all subscribers.
// This method is non-blocking; if a subscriber's channel is full, the message is dropped for that subscriber.
func (b *Broadcaster) Broadcast(msg []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Append to history
	// We make a copy of msg to ensure history persists even if caller reuses buffer
	msgCopy := make([]byte, len(msg))
	copy(msgCopy, msg)
	b.history = append(b.history, msgCopy)

	if len(b.history) > b.limit {
		// Simple slicing is efficient and sufficient.
		// Go's append will handle reallocation when capacity is reached.
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

// GetHistory returns the current log history.
func (b *Broadcaster) GetHistory() [][]byte {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := make([][]byte, len(b.history))
	for i, msg := range b.history {
		msgCopy := make([]byte, len(msg))
		copy(msgCopy, msg)
		result[i] = msgCopy
	}
	return result
}
