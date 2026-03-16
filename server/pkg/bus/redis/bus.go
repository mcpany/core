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
// Summary: Bus is a Redis-backed implementation of the Bus interface.
//
// Summary: Bus is a Redis-backed implementation of the Bus interface.
type Bus[T any] struct {
	client *redis.Client
// New creates and initializes a new RedisBus.
//
// Summary: New creates and initializes a new RedisBus.
//
// Parameters:
//   - redisConfig (*bus.RedisBus): The provided redisconfig data.
//
// Returns:
//   - *Bus[T]: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//
// Returns:
//   - *Bus[T]: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func New[T any](redisConfig *bus.RedisBus) (*Bus[T], error) {
	options := redis.Options{
		Addr: "127.0.0.1:6379",
	}
// NewWithClient creates a new RedisBus with an existing Redis client.
//
// Summary: NewWithClient creates a new RedisBus with an existing Redis client.
//
// Parameters:
//   - client (*redis.Client): The provided client data.
//
// Returns:
//   - *Bus[T]: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
	return NewWithClient[T](redis.NewClient(&options)), nil
}

// NewWithClient creates a new RedisBus with an existing Redis client.
//
// Summary: NewWithClient creates a new RedisBus with an existing Redis client.
// Publish publishes a message to a Redis channel.
//
// Summary: Publish publishes a message to a Redis channel.
//
// Parameters:
//   - ctx (context.Context): The cancellation and deadline context.
//   - topic (string): The textual representation of topic.
//   - msg (T): The provided msg data.
//
// Returns:
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//   - May modify internal state or perform external network calls.
func NewWithClient[T any](client *redis.Client) *Bus[T] {
	return &Bus[T]{
		client: client,
	}
}

// Publish publishes a message to a Redis channel.
// Subscribe subscribes to a Redis channel.
//
// Summary: Subscribe subscribes to a Redis channel.
//
// Parameters:
//   - ctx (context.Context): The cancellation and deadline context.
//   - topic (string): The textual representation of topic.
//   - handler (func(T)): The provided handler data.
//
// Returns:
//   - unsubscribe (func()): The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (b *Bus[T]) Publish(ctx context.Context, topic string, msg T) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return b.client.Publish(ctx, topic, payload).Err()
}

// Subscribe subscribes to a Redis channel.
//
// Summary: Subscribe subscribes to a Redis channel.
//
// Parameters:
//   - ctx (context.Context): The cancellation and deadline context.
//   - topic (string): The textual representation of topic.
//   - handler (func(T)): The provided handler data.
//
// Returns:
//   - unsubscribe (func()): The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
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
// SubscribeOnce subscribes to a topic for a single message.
//
// Summary: SubscribeOnce subscribes to a topic for a single message.
//
// Parameters:
//   - ctx (context.Context): The cancellation and deadline context.
//   - topic (string): The textual representation of topic.
//   - handler (func(T)): The provided handler data.
//
// Returns:
//   - unsubscribe (func()): The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
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
// Summary: SubscribeOnce subscribes to a topic for a single message.
//
// Parameters:
//   - ctx (context.Context): The cancellation and deadline context.
//   - topic (string): The textual representation of topic.
//   - handler (func(T)): The provided handler data.
//
// Returns:
//   - unsubscribe (func()): The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
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

// Close closes the Redis client connection.
//
// Summary: Closes the Redis connection.
//
// Returns:
//   - error: An error if closing fails.
func (b *Bus[T]) Close() error {
	return b.client.Close()
}
