
package bus

import (
	"context"
	"fmt"

	"github.com/mcpany/core/pkg/bus/memory"
	"github.com/mcpany/core/pkg/bus/nats"
	"github.com/mcpany/core/pkg/bus/redis"
	"github.com/mcpany/core/proto/bus"
	"github.com/puzpuzpuz/xsync/v4"
)

// Bus defines the interface for a generic, type-safe event bus that facilitates
// communication between different parts of the application. The type parameter T
// specifies the type of message that the bus will handle.
type Bus[T any] interface {
	// Publish sends a message to all subscribers of a given topic. The message
	// is sent to each subscriber's channel, and the handler is invoked by a
	// dedicated goroutine for that subscriber.
	//
	// ctx is the context for the publish operation.
	// topic is the topic to publish the message to.
	// msg is the message to be sent.
	Publish(ctx context.Context, topic string, msg T) error

	// Subscribe registers a handler function for a given topic. It starts a
	// dedicated goroutine for the subscription to process messages from a
	// channel.
	//
	// ctx is the context for the subscribe operation.
	// topic is the topic to subscribe to.
	// handler is the function to be called with the message.
	// It returns a function that can be called to unsubscribe the handler.
	Subscribe(ctx context.Context, topic string, handler func(T)) (unsubscribe func())

	// SubscribeOnce registers a handler function that will be invoked only once
	// for a given topic. After the handler is called, the subscription is
	// automatically removed.
	//
	// ctx is the context for the subscribe operation.
	// topic is the topic to subscribe to.
	// handler is the function to be called with the message.
	// It returns a function that can be called to unsubscribe the handler
	// before it has been invoked.
	SubscribeOnce(ctx context.Context, topic string, handler func(T)) (unsubscribe func())
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
	buses  *xsync.Map[string, any]
	config *bus.MessageBus
}

// NewBusProvider creates and returns a new BusProvider, which is used to manage
// multiple topic-based bus instances.
func NewBusProvider(messageBus *bus.MessageBus) (*BusProvider, error) {
	provider := &BusProvider{
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
	default:
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
// GetBusHook is a test hook for overriding the bus retrieval logic.
var GetBusHook func(p *BusProvider, topic string) any

func GetBus[T any](p *BusProvider, topic string) Bus[T] {
	if GetBusHook != nil {
		if bus := GetBusHook(p, topic); bus != nil {
			return bus.(Bus[T])
		}
	}

	if bus, ok := p.buses.Load(topic); ok {
		return bus.(Bus[T])
	}

	var newBus Bus[T]
	switch p.config.WhichBusType() {
	case bus.MessageBus_InMemory_case:
		newBus = memory.New[T]()
	case bus.MessageBus_Redis_case:
		newBus = redis.New[T](p.config.GetRedis())
	case bus.MessageBus_Nats_case:
		var err error
		newBus, err = nats.New[T](p.config.GetNats())
		if err != nil {
			panic(err)
		}
	}

	bus, _ := p.buses.LoadOrStore(topic, newBus)
	return bus.(Bus[T])
}
