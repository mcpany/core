// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/proto/bus"
	"github.com/redis/go-redis/v9"
)

// RedisBus is a Redis-backed implementation of the Bus interface.
type RedisBus[T any] struct {
	client *redis.Client
	mu     sync.RWMutex
	// We need to keep track of pubsub clients to close them
	pubsubs map[string]*redis.PubSub
}

// New creates a new RedisBus.
func New[T any](redisConfig *bus.RedisBus) *RedisBus[T] {
	var options redis.Options
	if redisConfig != nil {
		options.Addr = redisConfig.GetAddress()
		options.Password = redisConfig.GetPassword()
		options.DB = int(redisConfig.GetDb())
	}
	return NewWithClient[T](redis.NewClient(&options))
}

// NewWithClient creates a new RedisBus with an existing Redis client.
func NewWithClient[T any](client *redis.Client) *RedisBus[T] {
	return &RedisBus[T]{
		client:  client,
		pubsubs: make(map[string]*redis.PubSub),
	}
}

// Publish publishes a message to a Redis channel.
func (b *RedisBus[T]) Publish(ctx context.Context, topic string, msg T) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return b.client.Publish(ctx, topic, payload).Err()
}

// Subscribe subscribes to a Redis channel.
func (b *RedisBus[T]) Subscribe(ctx context.Context, topic string, handler func(T)) (unsubscribe func()) {
	if handler == nil {
		panic("redis bus: handler cannot be nil")
	}
	b.mu.Lock()
	if ps, ok := b.pubsubs[topic]; ok {
		ps.Close()
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
				ps.Close()
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
func (b *RedisBus[T]) SubscribeOnce(ctx context.Context, topic string, handler func(T)) (unsubscribe func()) {
	if handler == nil {
		panic("redis bus: handler cannot be nil")
	}
	var once sync.Once
	var unsub func()

	unsub = b.Subscribe(ctx, topic, func(msg T) {
		once.Do(func() {
			handler(msg)
			unsub()
		})
	})
	return unsub
}
