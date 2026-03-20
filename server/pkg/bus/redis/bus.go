// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package redis provides a Redis implementation of the bus.
// Summary: Bus is a Redis-backed implementation of the Bus interface.
//
// Side Effects:
//   - None.
//
// Summary: New creates and initializes a new RedisBus.
//
// Parameters:
//   - redisConfig: *bus.RedisBus. The configuration settings for the Redis bus.
//
// Returns:
//   - *Bus[T]: A pointer to the initialized Redis bus.
//   - error: An error if initialization fails (currently always nil).
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
//
// Summary: NewWithClient creates a new RedisBus with an existing Redis client.
//
// Parameters:
//   - client: *redis.Client. The existing Redis client instance.
//
// Returns:
//   - *Bus[T]: A pointer to the initialized Redis bus.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
//
// Summary: Publish publishes a message to a Redis channel.
//
// The message is marshaled to JSON before being published.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - topic: string. The topic (channel) to publish to.
//   - msg: T. The message payload.
//
// Returns:
//   - error: An error if marshaling or publishing fails.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
//
// Summary: Subscribe subscribes to a Redis channel.
//
// It starts a goroutine that continuously receives messages from the channel
// and invokes the provided handler.
//
// Parameters:
//   - ctx: context.Context. The context for the subscription.
//   - topic: string. The topic (channel) to subscribe to.
//   - handler: func(T). The callback function invoked for each message.
//
// Returns:
//   - func(): A function that unsubscribes the handler when called.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
//
// Summary: SubscribeOnce subscribes to a topic for a single message.
//
// It ensures that the handler is called only once for the next message received.
//
// Parameters:
//   - ctx: context.Context. The context for the subscription.
//   - topic: string. The topic (channel) to subscribe to.
//   - handler: func(T). The callback function invoked for the single message.
//
// Returns:
//   - func(): A function that unsubscribes the handler if called before the message is received.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
package redis

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/mcpany/core/proto/bus"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/redis/go-redis/v9"
)

type Bus[T any] struct {
	client *redis.Client
}

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

func NewWithClient[T any](client *redis.Client) *Bus[T] {
	return &Bus[T]{
		client: client,
	}
}

func (b *Bus[T]) Publish(ctx context.Context, topic string, msg T) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return b.client.Publish(ctx, topic, payload).Err()
}

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
	// Close closes the Redis client connection.
	//
	// Summary: Closes the Redis connection.
	//
	// Returns:
	//   - error: An error if closing fails.
	//
	// Parameters:
	//   - None.
	//
	// Errors:
	//   - None.
	//
	// Side Effects:
	//   - None.
	close(ready)

	return proxyUnsub
}

func (b *Bus[T]) Close() error {
	return b.client.Close()
}
