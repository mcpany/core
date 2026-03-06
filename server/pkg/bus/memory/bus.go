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

// DefaultBus - Auto-generated documentation.
//
// Summary: DefaultBus is the default, thread-safe implementation of the Bus interface.
//
// Fields:
//   - Various fields for DefaultBus.
type DefaultBus[T any] struct {
	mu             sync.RWMutex
	subscribers    map[string]map[uintptr]chan T
	nextID         uintptr
	publishTimeout time.Duration
}

// New creates and returns a new instance of DefaultBus, which is the default, thread-safe implementation of the Bus interface. It is initialized with the default publish timeout. The type parameter T specifies the type of message that the bus will handle.
//
// Parameters:
//   - None
//
// Returns:
//   - *DefaultBus[T]: The resulting *DefaultBus[T].
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func New[T any]() *DefaultBus[T] {
	return &DefaultBus[T]{
		subscribers:    make(map[string]map[uintptr]chan T),
		publishTimeout: defaultPublishTimeout,
	}
}

// Publish - Auto-generated documentation.
//
// Summary: Publish sends a message to all handlers subscribed to the specified topic.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func (b *DefaultBus[T]) Publish(_ context.Context, topic string, msg T) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if subs, ok := b.subscribers[topic]; ok {
		var wg sync.WaitGroup
		for id, ch := range subs {
			wg.Add(1)
			go func(id uintptr, ch chan T) {
				defer wg.Done()

				// ⚡ BOLT: Parallelized publish to avoid head-of-line blocking.
				// Randomized Selection from Top 5 High-Impact Targets
				timer := time.NewTimer(b.publishTimeout)
				defer timer.Stop()

				select {
				case ch <- msg:
					// Success
				case <-timer.C:
					// It's important to have a logging strategy for dropped messages.
					log := logging.GetLogger()
					log.Warn("Message dropped on topic", "topic", topic, "subscriber_id", id, "timeout", b.publishTimeout)
				}
			}(id, ch)
		}
		wg.Wait()
	}
	return nil
}

// Subscribe - Auto-generated documentation.
//
// Summary: Subscribe registers a handler function for a given topic. It starts a new
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
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

// SubscribeOnce - Auto-generated documentation.
//
// Summary: SubscribeOnce registers a handler for a topic that will be executed only
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
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
