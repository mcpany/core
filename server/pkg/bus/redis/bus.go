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

// subscriber holds a handler function and a unique ID.
type subscriber[T any] struct {
	id      int
	handler func(T)
}

// Bus is a Redis-backed implementation of the Bus interface.
type Bus[T any] struct {
	client     *redis.Client
	mu         sync.RWMutex
	nextID     int
	handlers   map[string][]*subscriber[T]
	pubsubs    map[string]*redis.PubSub
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
		client:   client,
		handlers: make(map[string][]*subscriber[T]),
		pubsubs:  make(map[string]*redis.PubSub),
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

	// Add new handler
	b.nextID++
	sub := &subscriber[T]{id: b.nextID, handler: handler}
	b.handlers[topic] = append(b.handlers[topic], sub)

	// If no active subscription for this topic, create one.
	if _, ok := b.pubsubs[topic]; !ok {
		pubsub := b.client.Subscribe(ctx, topic)
		b.pubsubs[topic] = pubsub

		// Start a single goroutine per topic to distribute messages
		go func() {
			log := logging.GetLogger()
			ch := pubsub.Channel()
			defer func() {
				b.mu.Lock()
				_ = pubsub.Close()
				delete(b.pubsubs, topic)
				b.mu.Unlock()
			}()
			for {
				select {
				case <-ctx.Done():
					return
				case msg, ok := <-ch:
					if !ok {
						return
					}
					var message T
					if err := json.Unmarshal([]byte(msg.Payload), &message); err != nil {
						log.Error("Failed to unmarshal message", "error", err)
						continue
					}

					// Call all handlers for this topic
					b.mu.RLock()
					// Create a copy of the slice to avoid holding the lock while calling handlers
					currentHandlers := make([]*subscriber[T], len(b.handlers[topic]))
					copy(currentHandlers, b.handlers[topic])
					b.mu.RUnlock()

					for _, s := range currentHandlers {
						func(s *subscriber[T]) {
							defer func() {
								if r := recover(); r != nil {
									log.Error("panic in handler", "error", r)
								}
							}()
							s.handler(message)
						}(s)
					}
				}
			}
		}()
	}

	b.mu.Unlock()

	var unsubscribeOnce sync.Once
	unsubscribe = func() {
		unsubscribeOnce.Do(func() {
			b.mu.Lock()
			defer b.mu.Unlock()

			// Remove the specific handler
			subs, ok := b.handlers[topic]
			if !ok {
				return
			}
			for i, s := range subs {
				if s.id == sub.id {
					b.handlers[topic] = append(subs[:i], subs[i+1:]...)
					break
				}
			}

			// If no handlers left, close the pubsub connection
			if len(b.handlers[topic]) == 0 {
				if ps, ok := b.pubsubs[topic]; ok {
					_ = ps.Close()
					delete(b.pubsubs, topic)
				}
				delete(b.handlers, topic)
			}
		})
	}
	return unsubscribe
}

// SubscribeOnce subscribes to a topic for a single message.
func (b *Bus[T]) SubscribeOnce(ctx context.Context, topic string, handler func(T)) (unsubscribe func()) {
	if handler == nil {
		logging.GetLogger().Error("redis bus: handler cannot be nil")
		return func() {}
	}
	var once sync.Once
	var unsub func()

	unsub = b.Subscribe(ctx, topic, func(msg T) {
		once.Do(func() {
			defer unsub()
			handler(msg)
		})
	})
	return unsub
}

// Close closes the Redis client and all pubsub connections.
func (b *Bus[T]) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	var lastErr error
	for topic, ps := range b.pubsubs {
		if err := ps.Close(); err != nil {
			lastErr = err
		}
		delete(b.pubsubs, topic)
	}

	if b.client != nil {
		if err := b.client.Close(); err != nil {
			lastErr = err
		}
	}

	return lastErr
}
