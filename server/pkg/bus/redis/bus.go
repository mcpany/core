// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package redis provides a Redis implementation of the bus.
package redis

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/mcpany/core/proto/bus"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/redis/go-redis/v9"
)

// Bus is a Redis-backed implementation of the Bus interface.
//
// Summary: Represents a Bus.
type Bus[T any] struct {
	client *redis.Client
}

// New creates and initializes a new RedisBus.
//
// Parameters: - None.
//   - redisConfig: *bus.RedisBus. The configuration settings for the Redis bus.
//
// Returns: - None.
//   - *Bus[T]: A pointer to the initialized Redis bus.
//   - error: An error if initialization fails (currently always nil).
//
// Summary: Initializes New operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func New[T any](redisConfig *bus.RedisBus) (*Bus[T], error) {
	options := redis.Options{
		Addr: "127.0.0.1:6379",
	}
	if redisConfig != nil {
		if addr := redisConfig.GetAddress(); addr != "" {
			options.Addr = addr
		}
		options.Password = redisConfig.GetPassword()
		options.DB = int(redisConfig.GetDb())
	}
	return NewWithClient[T](redis.NewClient(&options)), nil
}

// NewWithClient creates a new RedisBus with an existing Redis client.
//
// Parameters: - None.
//   - client: *redis.Client. The existing Redis client instance.
//
// Returns: - None.
//   - *Bus[T]: A pointer to the initialized Redis bus.
//
// Summary: Initializes NewWithClient operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func NewWithClient[T any](client *redis.Client) *Bus[T] {
	return &Bus[T]{
		client: client,
	}
}

// Publish publishes a message to a Redis channel.
//
// The message is marshaled to JSON before being published.
//
// Parameters: - None.
//   - ctx: context.Context. The context for the request.
//   - topic: string. The topic (channel) to publish to.
//   - msg: T. The message payload.
//
// Returns: - None.
//   - error: An error if marshaling or publishing fails.
//
// Summary: Executes Publish operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func (b *Bus[T]) Publish(ctx context.Context, topic string, msg T) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return b.client.Publish(ctx, topic, payload).Err()
}

// Subscribe subscribes to a Redis channel.
//
// It starts a goroutine that continuously receives messages from the channel
// and invokes the provided handler.
//
// Parameters: - None.
//   - ctx: context.Context. The context for the subscription.
//   - topic: string. The topic (channel) to subscribe to.
//   - handler: func(T). The callback function invoked for each message.
//
// Returns: - None.
//   - func(): A function that unsubscribes the handler when called.
//
// Summary: Executes Subscribe operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func (b *Bus[T]) Subscribe(ctx context.Context, topic string, handler func(T)) (unsubscribe func()) {
	if handler == nil {
		logging.GetLogger().Error("redis bus: handler cannot be nil")
		return func() {}
	}

	pubsub := b.client.Subscribe(ctx, topic)

	var unsubscribeOnce sync.Once
	unsubscribe = func() {
		unsubscribeOnce.Do(func() {
			_ = pubsub.Close()
		})
	}

	go func() {
		defer unsubscribe()
		log := logging.GetLogger()
		ch := pubsub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-ch:
				if msg == nil {
					return
				}
				var message T
				err := json.Unmarshal([]byte(msg.Payload), &message)
				if err != nil {
					log.Error("Failed to unmarshal message", "error", err)
					continue
				}

				func() {
					defer func() {
						if r := recover(); r != nil {
							log.Error("panic in handler", "error", r)
						}
					}()
					handler(message)
				}()
			}
		}
	}()

	return unsubscribe
}

// SubscribeOnce subscribes to a topic for a single message.
//
// It ensures that the handler is called only once for the next message received.
//
// Parameters: - None.
//   - ctx: context.Context. The context for the subscription.
//   - topic: string. The topic (channel) to subscribe to.
//   - handler: func(T). The callback function invoked for the single message.
//
// Returns: - None.
//   - func(): A function that unsubscribes the handler if called before the message is received.
//
// Summary: Executes SubscribeOnce operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func (b *Bus[T]) SubscribeOnce(ctx context.Context, topic string, handler func(T)) (unsubscribe func()) {
	if handler == nil {
		logging.GetLogger().Error("redis bus: handler cannot be nil")
		return func() {}
	}
	var once sync.Once
	// Use a channel to ensure regularUnsub is set before we try to call it
	ready := make(chan struct{})
	var regularUnsub func()

	// proxyUnsub waits for the real unsubscribe function to be available
	proxyUnsub := func() {
		<-ready
		if regularUnsub != nil {
			regularUnsub()
		}
	}

	regularUnsub = b.Subscribe(ctx, topic, func(msg T) {
		once.Do(func() {
			handler(msg)
			proxyUnsub()
		})
	})

	// Signal that regularUnsub is set
	close(ready)

	return proxyUnsub
}

// Close close close.
//
// Summary: Close close.
//
// Parameters:
//   - None.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (b *Bus[T]) Close() error {
	return b.client.Close()
}
