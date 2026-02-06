// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package memory provides in-memory implementations of the bus interface.
package memory

import (
	"context"
	"sync"
	"time"

	"github.com/mcpany/core/server/pkg/logging"
)

const (
	// defaultPublishTimeout is the default duration to wait for a subscriber to
	// accept a message before dropping it.
	defaultPublishTimeout = 1 * time.Second
)

// DefaultBus is the default, thread-safe implementation of the Bus interface.
// It uses channels to deliver messages to subscribers, with each subscriber
// having its own dedicated goroutine for message processing.
type DefaultBus[T any] struct {
	mu             sync.RWMutex
	subscribers    map[string]map[uintptr]chan T
	nextID         uintptr
	publishTimeout time.Duration
}

// New creates and returns a new instance of DefaultBus, which is the default,
// thread-safe implementation of the Bus interface. It is initialized with the
// default publish timeout.
//
// The type parameter T specifies the type of message that the bus will handle.
//
// Returns:
//   - *DefaultBus[T]: A pointer to the initialized DefaultBus.
func New[T any]() *DefaultBus[T] {
	return &DefaultBus[T]{
		subscribers:    make(map[string]map[uintptr]chan T),
		publishTimeout: defaultPublishTimeout,
	}
}

// Publish sends a message to all handlers subscribed to the specified topic.
// It sends the message to a channel for each subscriber, where it will be
// processed by the subscriber's dedicated goroutine.
//
// To prevent a slow subscriber from blocking the publisher indefinitely, this
// call will time out after a configurable duration if a subscriber's channel is
// full. If a timeout occurs, the message is dropped for that subscriber, and a
// warning is logged.
//
// Parameters:
//   - _: context.Context. Unused.
//   - topic: string. The topic to publish to.
//   - msg: T. The message to send.
//
// Returns:
//   - error: Always nil for in-memory bus.
func (b *DefaultBus[T]) Publish(_ context.Context, topic string, msg T) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if subs, ok := b.subscribers[topic]; ok {
		for id, ch := range subs {
			// Use a non-blocking send with a timeout to avoid blocking the
			// publisher indefinitely.
			select {
			case ch <- msg:
			case <-time.After(b.publishTimeout):
				// It's important to have a logging strategy for dropped messages.
				log := logging.GetLogger()
				log.Warn("Message dropped on topic", "topic", topic, "subscriber_id", id, "timeout", b.publishTimeout)
			}
		}
	}
	return nil
}

// Subscribe registers a handler function for a given topic. It starts a new
// goroutine for each subscription to process messages from a buffered channel,
// ensuring that subscribers handle messages independently and do not block each
// other.
//
// Each subscriber is assigned a unique ID, and its channel is added to the list
// of subscribers for the given topic.
//
// Parameters:
//   - _: context.Context. Unused.
//   - topic: string. The topic to subscribe to.
//   - handler: func(T). The callback function.
//
// Returns:
//   - func(): A function to unsubscribe.
func (b *DefaultBus[T]) Subscribe(_ context.Context, topic string, handler func(T)) (unsubscribe func()) {
	b.mu.Lock()
	defer b.mu.Unlock()

	id := b.nextID
	b.nextID++

	if _, ok := b.subscribers[topic]; !ok {
		b.subscribers[topic] = make(map[uintptr]chan T)
	}

	// Create a buffered channel for the subscriber to prevent blocking the publisher.
	ch := make(chan T, 128)
	b.subscribers[topic][id] = ch

	// Start a dedicated goroutine for this subscriber to process messages.
	// This goroutine will exit when the channel is closed.
	go func() {
		for msg := range ch {
			handler(msg)
		}
	}()

	// Return a function to unsubscribe.
	return func() {
		b.mu.Lock()
		defer b.mu.Unlock()

		if subs, ok := b.subscribers[topic]; ok {
			if subCh, ok := subs[id]; ok {
				// Remove the subscriber from the map.
				delete(subs, id)
				if len(subs) == 0 {
					delete(b.subscribers, topic)
				}
				// Close the channel to terminate the subscriber's goroutine.
				close(subCh)
			}
		}
	}
}

// SubscribeOnce registers a handler for a topic that will be executed only
// once. After the handler is invoked for the first time, the subscription is
// automatically removed.
//
// Parameters:
//   - ctx: context.Context. The context passed to Subscribe.
//   - topic: string. The topic to subscribe to.
//   - handler: func(T). The callback function.
//
// Returns:
//   - func(): A function to unsubscribe.
func (b *DefaultBus[T]) SubscribeOnce(ctx context.Context, topic string, handler func(T)) (unsubscribe func()) {
	var once sync.Once
	var unsub func()

	unsub = b.Subscribe(ctx, topic, func(msg T) {
		once.Do(func() {
			unsub()
			handler(msg)
		})
	})
	return unsub
}
