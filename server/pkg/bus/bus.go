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

// Bus defines the interface for a generic, type-safe event bus that facilitates.
//
// Summary: defines the interface for a generic, type-safe event bus that facilitates.
type Bus[T any] interface {
	// Publish sends a message to all subscribers of a given topic. The message.
	//
	// Summary: sends a message to all subscribers of a given topic. The message.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the operation.
	//   - topic: string. The string.
	//   - msg: T. The t.
	//
	// Returns:
	//   - error: An error if the operation fails.
	//
	// Throws/Errors:
	//   Returns an error if the operation fails.
	Publish(ctx context.Context, topic string, msg T) error

	// Subscribe registers a handler function for a given topic. It starts a.
	//
	// Summary: registers a handler function for a given topic. It starts a.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the operation.
	//   - topic: string. The string.
	//   - handler: func(T). The func( t).
	//
	// Returns:
	//   - unsubscribe: func(). The func().
	Subscribe(ctx context.Context, topic string, handler func(T)) (unsubscribe func())

	// SubscribeOnce registers a handler function that will be invoked only once.
	//
	// Summary: registers a handler function that will be invoked only once.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the operation.
	//   - topic: string. The string.
	//   - handler: func(T). The func( t).
	//
	// Returns:
	//   - unsubscribe: func(). The func().
	SubscribeOnce(ctx context.Context, topic string, handler func(T)) (unsubscribe func())
}

// Provider is a thread-safe container for managing multiple, type-safe bus.
//
// Summary: is a thread-safe container for managing multiple, type-safe bus.
type Provider struct {
	buses  *xsync.Map[string, any]
	config *bus.MessageBus
}

// NewProviderHook is a test hook for overriding the NewProvider logic.
var NewProviderHook func(*bus.MessageBus) (*Provider, error)

// NewProvider creates and returns a new Provider, which is used to manage.
//
// Summary: creates and returns a new Provider, which is used to manage.
//
// Parameters:
//   - messageBus: *bus.MessageBus. The messageBus.
//
// Returns:
//   - *Provider: The *Provider.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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

// GetBus retrieves a bus for the given topic. If a bus for the given topic.
//
// Summary: retrieves a bus for the given topic. If a bus for the given topic.
//
// Parameters:
//   - p: *Provider. The p.
//   - topic: string. The topic.
//
// Returns:
//   - Bus[T]: The Bus[T].
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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
