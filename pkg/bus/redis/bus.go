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
	"github.com/redis/go-redis/v9"
)

// RedisBus is a Redis-backed implementation of the Bus interface.
type RedisBus[T any] struct {
	client *redis.Client
	ctx    context.Context
	mu     sync.RWMutex
	// We need to keep track of pubsub clients to close them
	pubsubs map[string]*redis.PubSub
}

// New creates a new RedisBus.
func New[T any](client *redis.Client) *RedisBus[T] {
	return &RedisBus[T]{
		client:  client,
		ctx:     context.Background(),
		pubsubs: make(map[string]*redis.PubSub),
	}
}

// Publish publishes a message to a Redis channel.
func (b *RedisBus[T]) Publish(topic string, msg T) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return b.client.Publish(b.ctx, topic, payload).Err()
}

// Subscribe subscribes to a Redis channel.
func (b *RedisBus[T]) Subscribe(topic string, handler func(T)) (unsubscribe func()) {
	b.mu.Lock()
	defer b.mu.Unlock()

	pubsub := b.client.Subscribe(b.ctx, topic)
	b.pubsubs[topic] = pubsub

	go func() {
		ch := pubsub.Channel()
		for msg := range ch {
			var message T
			err := json.Unmarshal([]byte(msg.Payload), &message)
			if err != nil {
				log := logging.GetLogger()
				log.Error("Failed to unmarshal message", "error", err)
				continue
			}
			handler(message)
		}
	}()

	return func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		if ps, ok := b.pubsubs[topic]; ok {
			ps.Close()
			delete(b.pubsubs, topic)
		}
	}
}

// SubscribeOnce subscribes to a topic for a single message.
func (b *RedisBus[T]) SubscribeOnce(topic string, handler func(T)) (unsubscribe func()) {
	var once sync.Once
	var unsub func()

	unsub = b.Subscribe(topic, func(msg T) {
		once.Do(func() {
			handler(msg)
			unsub()
		})
	})
	return unsub
}

