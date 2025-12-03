/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
	pubsubs map[string]PubSubClient
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
		pubsubs: make(map[string]PubSubClient),
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

// GetPubsubs returns the pubsubs map for the bus. This is intended for testing purposes.
func (b *RedisBus[T]) GetPubsubs() map[string]PubSubClient {
	return b.pubsubs
}

// Close closes the Redis client and all pubsub connections.
func (b *RedisBus[T]) Close() error {
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
