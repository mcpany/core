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

package bus

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/mcpany/core/pkg/logging"
)

// RedisBus is a Redis-based implementation of the Bus interface. It uses Redis
// Pub/Sub to facilitate communication between different parts of the application.
type RedisBus[T any] struct {
	mu             sync.RWMutex
	redisClient    *redis.Client
	subscribers    map[string]map[uintptr]chan T
	nextID         uintptr
	publishTimeout time.Duration
}

// NewRedisBus creates a new instance of RedisBus. It requires a Redis client to
// connect to the Redis server.
func NewRedisBus[T any](redisClient *redis.Client) *RedisBus[T] {
	return &RedisBus[T]{
		redisClient:    redisClient,
		subscribers:    make(map[string]map[uintptr]chan T),
		publishTimeout: defaultPublishTimeout,
	}
}

// Publish sends a message to all subscribers of a given topic using Redis Pub/Sub.
// The message is marshaled to JSON before being published.
func (b *RedisBus[T]) Publish(topic string, msg T) {
	log := logging.GetLogger()
	payload, err := json.Marshal(msg)
	if err != nil {
		log.Error("Failed to marshal message", "error", err)
		return
	}
	ctx := context.Background()
	if err := b.redisClient.Publish(ctx, topic, payload).Err(); err != nil {
		log.Error("Failed to publish message to Redis", "error", err)
	}
}

// Subscribe registers a handler function for a given topic. It subscribes to the
// topic on Redis and starts a goroutine to listen for messages.
func (b *RedisBus[T]) Subscribe(topic string, handler func(T)) (unsubscribe func()) {
	// This is a simplified implementation. A more robust implementation would
	// handle connection errors and reconnections.
	ctx := context.Background()
	pubsub := b.redisClient.Subscribe(ctx, topic)

	// Wait for confirmation that subscription is created before publishing anything.
	_, err := pubsub.Receive(ctx)
	if err != nil {
		logging.GetLogger().Error("Failed to subscribe to topic", "topic", topic, "error", err)
		return func() {}
	}

	// Go channel which receives messages.
	ch := pubsub.Channel()

	go func() {
		for msg := range ch {
			var message T
			if err := json.Unmarshal([]byte(msg.Payload), &message); err != nil {
				logging.GetLogger().Error("Failed to unmarshal message", "error", err)
				continue
			}
			handler(message)
		}
	}()

	return func() {
		if err := pubsub.Unsubscribe(ctx, topic); err != nil {
			logging.GetLogger().Error("Failed to unsubscribe from topic", "topic", topic, "error", err)
		}
		if err := pubsub.Close(); err != nil {
			logging.GetLogger().Error("Failed to close pubsub", "error", err)
		}
	}
}

// SubscribeOnce registers a handler function that will be invoked only once.
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
