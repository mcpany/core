// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package redis provides a Redis implementation of the bus.
package redis

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/proto/bus"
	"github.com/redis/go-redis/v9"
)

// Bus is a Redis-backed implementation of the Bus interface.
type Bus[T any] struct {
	client *redis.Client
	mu     sync.RWMutex
	// We need to keep track of pubsub clients to close them
	pubsubs map[string]*redis.PubSub
}

// New creates a new RedisBus.
func New[T any](redisConfig *bus.RedisBus) (*Bus[T], error) {
	var options redis.Options
	if redisConfig != nil {
		options.Addr = redisConfig.GetAddress()
		options.Password = redisConfig.GetPassword()
		options.DB = int(redisConfig.GetDb())
	}
	return NewWithClient[T](redis.NewClient(&options)), nil
}

// NewWithClient creates a new RedisBus with an existing Redis client.
func NewWithClient[T any](client *redis.Client) *Bus[T] {
	return &Bus[T]{
		client:  client,
		pubsubs: make(map[string]*redis.PubSub),
	}
}

// Publish publishes a message to a Redis channel.
func (b *Bus[T]) Publish(ctx context.Context, topic string, msg T) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return b.client.Publish(ctx, topic, payload).Err()
}

// Subscribe subscribes to a Redis channel.
func (b *Bus[T]) Subscribe(ctx context.Context, topic string, handler func(T)) (unsubscribe func()) {
	if handler == nil {
		logging.GetLogger().Error("redis bus: handler cannot be nil")
		return func() {}
	}
	b.mu.Lock()
	if ps, ok := b.pubsubs[topic]; ok {
		_ = ps.Close()
	}

	pubsub := b.client.Subscribe(ctx, topic)
	b.pubsubs[topic] = pubsub
	b.mu.Unlock()

	var unsubscribeOnce sync.Once
	unsubscribe = func() {
		unsubscribeOnce.Do(func() {
			b.mu.Lock()
			defer b.mu.Unlock()
			if ps, ok := b.pubsubs[topic]; ok && ps == pubsub {
				_ = ps.Close()
				delete(b.pubsubs, topic)
			}
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

// Close closes the Redis client and all pubsub connections.
func (b *Bus[T]) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	var lastErr error
	for topic, ps := range b.pubsubs {
		if err := ps.Close(); err != nil {
			lastErr = err
			// Optionally log the error
		}
		delete(b.pubsubs, topic)
	}

	if err := b.client.Close(); err != nil {
		lastErr = err
	}

	return lastErr
}
