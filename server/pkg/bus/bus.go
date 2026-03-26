// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package bus defines the message bus interface and implementations.
package bus

import (
	"context"
	"fmt"

	"github.com/mcpany/core/server/pkg/bus/kafka"
	"github.com/mcpany/core/server/pkg/bus/memory"
	"github.com/mcpany/core/server/pkg/bus/nats"
	"github.com/mcpany/core/server/pkg/bus/redis"
	"github.com/mcpany/core/proto/bus"
	xsync "github.com/puzpuzpuz/xsync/v4"
)

// Bus defines the interface for a generic, type-safe event bus that facilitates
// communication between different parts of the application. The type parameter T
// specifies the type of message that the bus will handle.
type Bus[T any] interface {
	// Publish sends a message to all subscribers of a given topic. The message
	// is sent to each subscriber's channel, and the handler is invoked by a
	// dedicated goroutine for that subscriber.
	//
	// Parameters:
	//   - ctx: context.Context. Controls the lifecycle of the publish operation (e.g. timeouts).
	//   - topic: string. The destination topic for the message.
	//   - msg: T. The payload message to be broadcasted.
	//
	// Returns:
	//   - error: An error if the publish operation fails (e.g. underlying transport error).
	Publish(ctx context.Context, topic string, msg T) error

	// Subscribe registers a handler function for a given topic. It starts a
	// dedicated goroutine for the subscription to process messages from a
	// channel.
	//
	// Parameters:
	//   - ctx: context.Context. Controls the setup of the subscription. Note that context cancellation
	//     may not automatically unsubscribe depending on implementation; use the returned unsubscribe function.
	//   - topic: string. The topic to listen to.
	//   - handler: func(T). The callback function invoked for each received message.
	//
	// Returns:
	//   - func(): A cleanup function that removes the subscription when called.
	Subscribe(ctx context.Context, topic string, handler func(T)) (unsubscribe func())

	// SubscribeOnce registers a handler function that will be invoked only once
	// for a given topic. After the handler is called, the subscription is
	// automatically removed.
	//
	// Parameters:
	//   - ctx: context.Context. Controls the setup of the subscription.
	//   - topic: string. The topic to listen to.
	//   - handler: func(T). The callback function invoked for the single received message.
	//
	// Returns:
	//   - func(): A cleanup function that removes the subscription if called before the first message.
	SubscribeOnce(ctx context.Context, topic string, handler func(T)) (unsubscribe func())
}

// Provider is a thread-safe container for managing multiple, type-safe bus
// instances, with each bus being dedicated to a specific topic. It ensures that
// for any given topic, there is only one bus instance, creating one on demand
// if it doesn't already exist.
//
// This allows different parts of the application to get a bus for a specific
// message type and topic without needing to manage the lifecycle of the bus
// instances themselves.
type Provider struct {
	buses  *xsync.Map[string, any]
	config *bus.MessageBus
}

// NewProviderHook is a test hook for overriding the NewProvider logic.
var NewProviderHook func(*bus.MessageBus) (*Provider, error)

// NewProvider creates and returns a new Provider, which is used to manage
// multiple topic-based bus instances.
//
// Parameters:
//   messageBus: The configuration for the message bus.
//
// Returns:
//   *Provider: The created Provider.
//   error: An error if creation fails.
func NewProvider(messageBus *bus.MessageBus) (*Provider, error) {
	if NewProviderHook != nil {
		return NewProviderHook(messageBus)
	}
	provider := &Provider{
		buses:  xsync.NewMap[string, any](),
		config: messageBus,
	}

	if provider.config == nil {
		provider.config = &bus.MessageBus{}
	}

	if !provider.config.HasBusType() {
		provider.config.SetInMemory(&bus.InMemoryBus{})
	}

	switch provider.config.WhichBusType() {
	case bus.MessageBus_InMemory_case:
		// In-memory bus requires no additional setup
	case bus.MessageBus_Redis_case:
		// Redis client is now created within the RedisBus
	case bus.MessageBus_Nats_case:
		// NATS client is now created within the NatsBus
	case bus.MessageBus_Kafka_case:
		// Kafka writer is now created within the KafkaBus
	default:
		return nil, fmt.Errorf("unknown bus type")
	}

	return provider, nil
}

// GetBusHook is a test hook for overriding the bus retrieval logic.
var GetBusHook func(p *Provider, topic string) (any, error)

// GetBus retrieves a bus for the given topic. If a bus for the given topic
// already exists, it is returned; otherwise, a new one is created and stored for
// future use.
//
// The type parameter T specifies the message type for the bus, ensuring
// type safety for each topic.
//
// Parameters:
//   p: The Provider instance.
//   topic: The topic name.
//
// Returns:
//   Bus[T]: The requested Bus instance.
//   error: An error if retrieval or creation fails.
func GetBus[T any](p *Provider, topic string) (Bus[T], error) {
	if GetBusHook != nil {
		bus, err := GetBusHook(p, topic)
		if err != nil {
			return nil, err
		}
		if bus != nil {
			return bus.(Bus[T]), nil
		}
	}

	if bus, ok := p.buses.Load(topic); ok {
		return bus.(Bus[T]), nil
	}

	var newBus Bus[T]
	var err error

	switch p.config.WhichBusType() {
	case bus.MessageBus_InMemory_case:
		newBus = memory.New[T]()
	case bus.MessageBus_Redis_case:
		newBus, err = redis.New[T](p.config.GetRedis())
		if err != nil {
			return nil, fmt.Errorf("failed to create Redis bus: %w", err)
		}
	case bus.MessageBus_Nats_case:
		newBus, err = nats.New[T](p.config.GetNats())
		if err != nil {
			return nil, fmt.Errorf("failed to create NATS bus: %w", err)
		}
	case bus.MessageBus_Kafka_case:
		newBus, err = kafka.New[T](p.config.GetKafka())
		if err != nil {
			return nil, fmt.Errorf("failed to create Kafka bus: %w", err)
		}
	default:
		return nil, fmt.Errorf("unknown bus type: %v", p.config.WhichBusType())
	}

	bus, _ := p.buses.LoadOrStore(topic, newBus)
	return bus.(Bus[T]), nil
}
