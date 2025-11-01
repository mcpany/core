/*
 * Copyright 2025 Author(s) of MCP Any
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
	"fmt"
	"sync"

	"github.com/mcpany/core/pkg/bus/memory"
	"github.com/mcpany/core/pkg/bus/redis"
	"github.com/mcpany/core/proto/bus"
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
	Publish(topic string, msg T) error

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

// BusProvider is a thread-safe container for managing multiple, type-safe bus
// instances, with each bus being dedicated to a specific topic. It ensures that
// for any given topic, there is only one bus instance, creating one on demand
// if it doesn't already exist.
//
// This allows different parts of the application to get a bus for a specific
// message type and topic without needing to manage the lifecycle of the bus
// instances themselves.
type BusProvider struct {
	buses  map[string]any
	mu     sync.RWMutex
	config *bus.MessageBus
}

// NewBusProvider creates and returns a new BusProvider, which is used to manage
// multiple topic-based bus instances.
func NewBusProvider(messageBus *bus.MessageBus) (*BusProvider, error) {
	provider := &BusProvider{
		buses:  make(map[string]any),
		config: messageBus,
	}

	if provider.config == nil {
		provider.config = &bus.MessageBus{}
	}

	if provider.config.GetInMemory() == nil && provider.config.GetRedis() == nil {
		provider.config.SetInMemory(&bus.InMemoryBus{})
	}

	if provider.config.GetInMemory() != nil {
		// In-memory bus requires no additional setup
	} else if provider.config.GetRedis() != nil {
		// Redis client is now created within the RedisBus
	} else {
		return nil, fmt.Errorf("unknown bus type")
	}

	return provider, nil
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

	var newBus Bus[T]
	if p.config.GetInMemory() != nil {
		newBus = memory.New[T]()
	} else if p.config.GetRedis() != nil {
		newBus = redis.New[T](p.config.GetRedis())
	}

	p.buses[topic] = newBus
	return newBus
}
