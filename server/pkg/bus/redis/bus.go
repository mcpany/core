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
type Bus[T any] struct {
	client *redis.Client
}

// New creates a new RedisBus instance using the provided configuration.
//
// Parameters:
//   - redisConfig: *bus.RedisBus. The configuration object containing address, password, and DB index.
//
// Returns:
//   - *Bus[T]: A pointer to the initialized RedisBus.
//   - error: An error if the configuration causes initialization issues (always nil currently).
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
// Parameters:
//   - client: *redis.Client. The pre-configured Redis client instance.
//
// Returns:
//   - *Bus[T]: A pointer to the initialized RedisBus.
func NewWithClient[T any](client *redis.Client) *Bus[T] {
	return &Bus[T]{
		client: client,
	}
}

// Publish publishes a message to a Redis channel.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - topic: string. The topic to publish to.
//   - msg: T. The message payload to publish.
//
// Returns:
//   - error: An error if serialization fails or the Redis publish operation fails.
func (b *Bus[T]) Publish(ctx context.Context, topic string, msg T) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return b.client.Publish(ctx, topic, payload).Err()
}

// Subscribe subscribes to a Redis channel and invokes the handler for each received message.
// It manages the subscription and handles message unmarshaling.
//
// Parameters:
//   - ctx: context.Context. The context for the subscription.
//   - topic: string. The topic to subscribe to.
//   - handler: func(T). The callback function to invoke for each message.
//
// Returns:
//   - func(): A function to unsubscribe and close the subscription.
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

// SubscribeOnce subscribes to a Redis topic and invokes the handler exactly once for the first received message.
//
// Parameters:
//   - ctx: context.Context. The context for the subscription.
//   - topic: string. The topic to subscribe to.
//   - handler: func(T). The callback function to invoke.
//
// Returns:
//   - func(): A function to unsubscribe (if called before the message is received).
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

// Close closes the underlying Redis client and releases resources.
//
// Returns:
//   - error: An error if closing the client fails.
func (b *Bus[T]) Close() error {
	return b.client.Close()
}
