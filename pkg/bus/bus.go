/*
 * Copyright 2025 Author(s) of MCP-XY
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package bus

import (
	"sync"
	"time"

	"github.com/mcpxy/core/pkg/logging"
)

const (
	// defaultPublishTimeout is the default duration to wait for a subscriber to
	// accept a message before dropping it.
	defaultPublishTimeout = 1 * time.Second
)

// Bus defines the interface for a generic, type-safe event bus that facilitates
// communication between different parts of the application. The type parameter T
// specifies the type of message that the bus will handle.
type Bus[T any] interface {
	// Publish sends a message to all subscribers of a given topic. The message
	// is sent to each subscriber's channel, and the handler is invoked by a
	// dedicated goroutine for that subscriber.
	//
	// topic is the topic to publish the message to.
	// msg is the message to be sent.
	Publish(topic string, msg T)

	// Subscribe registers a handler function for a given topic. It starts a
	// dedicated goroutine for the subscription to process messages from a
	// channel.
	//
	// topic is the topic to subscribe to.
	// handler is the function to be called with the message.
	// It returns a function that can be called to unsubscribe the handler.
	Subscribe(topic string, handler func(T)) (unsubscribe func())

	// SubscribeOnce registers a handler function that will be invoked only once
	// for a given topic. After the handler is called, the subscription is
	// automatically removed.
	//
	// topic is the topic to subscribe to.
	// handler is the function to be called with the message.
	// It returns a function that can be called to unsubscribe the handler
	// before it has been invoked.
	SubscribeOnce(topic string, handler func(T)) (unsubscribe func())
}

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
func (b *DefaultBus[T]) Publish(topic string, msg T) {
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
func (b *DefaultBus[T]) Subscribe(topic string, handler func(T)) (unsubscribe func()) {
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
func (b *DefaultBus[T]) SubscribeOnce(topic string, handler func(T)) (unsubscribe func()) {
	var once sync.Once
	var unsub func()

	unsub = b.Subscribe(topic, func(msg T) {
		handler(msg)
		once.Do(unsub)
	})
	return unsub
}

// BusProvider is a thread-safe container for managing multiple, type-safe bus
// instances, with each bus being dedicated to a specific topic. It ensures that
// for any given topic, there is only one bus instance, creating one on demand
// if it doesn't already exist.
//
// This allows different parts of the application to get a bus for a specific
// message type and topic without needing to manage the lifecycle of the bus
// instances themselves.
type BusProvider struct {
	buses map[string]any
	mu    sync.RWMutex
}

// NewBusProvider creates and returns a new BusProvider, which is used to manage
// multiple topic-based bus instances.
func NewBusProvider() *BusProvider {
	return &BusProvider{
		buses: make(map[string]any),
	}
}

// GetBus retrieves or creates a bus for a specific topic and message type. If a
// bus for the given topic already exists, it is returned; otherwise, a new one
// is created and stored for future use.
//
// The type parameter T specifies the message type for the bus, ensuring
// type safety for each topic.
//
// Parameters:
//   - p: The BusProvider instance.
//   - topic: The name of the topic for which to get the bus.
//
// Returns a Bus instance for the specified message type and topic.
func GetBus[T any](p *BusProvider, topic string) Bus[T] {
	p.mu.Lock()
	defer p.mu.Unlock()

	if bus, ok := p.buses[topic]; ok {
		return bus.(Bus[T])
	}

	bus := New[T]()
	p.buses[topic] = bus
	return bus
}
