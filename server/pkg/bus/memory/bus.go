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
//   - topic: The topic to publish the message to.
//   - msg: The message to be sent.
func (b *DefaultBus[T]) Publish(_ context.Context, topic string, msg T) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if subs, ok := b.subscribers[topic]; ok {
		// ⚡ BOLT: Reuse timer to avoid allocating a new timer and channel for every subscriber.
		// Randomized Selection from Top 5 High-Impact Targets
		timer := time.NewTimer(b.publishTimeout)
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		defer timer.Stop()

		for id, ch := range subs {
			timer.Reset(b.publishTimeout)

			// Use a non-blocking send with a timeout to avoid blocking the
			// publisher indefinitely.
			select {
			case ch <- msg:
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
			case <-timer.C:
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
//   - topic: The topic to subscribe to.
//   - handler: The function to execute when a message is received.
//
// Returns an `unsubscribe` function that can be called to remove the
// subscription. When called, it removes the subscriber from the bus and closes
// its channel, terminating the associated goroutine.
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
// This is useful for scenarios where a component needs to wait for a specific
// event to occur once and then stop listening.
//
// Parameters:
//   - topic: The topic to subscribe to.
//   - handler: The function to execute.
//
// Returns a function that can be used to unsubscribe before the handler is
// invoked.
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
