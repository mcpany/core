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

	"github.com/mcpany/core/pkg/logging"
	"github.com/redis/go-redis/v9"
)

// RedisBus is a message bus implementation that uses Redis as the backend.
// It supports publishing messages to topics and subscribing to topics to
// receive messages.
type RedisBus[T any] struct {
	client         *redis.Client
	publishTimeout time.Duration
}

// NewRedisBus creates a new RedisBus with the given Redis client.
func NewRedisBus[T any](client *redis.Client) *RedisBus[T] {
	return &RedisBus[T]{
		client:         client,
		publishTimeout: defaultPublishTimeout,
	}
}

// Publish sends a message to the specified topic using Redis Pub/Sub. The
// message is marshaled to JSON before being sent.
func (b *RedisBus[T]) Publish(topic string, msg T) {
	log := logging.GetLogger()
	payload, err := json.Marshal(msg)
	if err != nil {
		log.Error("Failed to marshal message", "error", err)
		return
	}

	if err := b.client.Publish(context.Background(), topic, payload).Err(); err != nil {
		log.Error("Failed to publish message to Redis", "error", err)
	}
}

// Subscribe registers a handler for a topic. It listens for messages on the
// given topic using Redis Pub/Sub and invokes the handler when a message is
// received.
func (b *RedisBus[T]) Subscribe(topic string, handler func(T)) (unsubscribe func()) {
	pubsub := b.client.Subscribe(context.Background(), topic)
	_, err := pubsub.Receive(context.Background())
	if err != nil {
		logging.GetLogger().Error("Failed to subscribe to topic", "error", err)
		return func() {}
	}

	ch := pubsub.Channel()
	closeCh := make(chan struct{})

	go func() {
		for {
			select {
			case msg, ok := <-ch:
				if !ok {
					return
				}
				var message T
				if err := json.Unmarshal([]byte(msg.Payload), &message); err != nil {
					logging.GetLogger().Error("Failed to unmarshal message", "error", err)
					continue
				}
				handler(message)
			case <-closeCh:
				return
			}
		}
	}()

	return func() {
		close(closeCh)
		pubsub.Close()
	}
}

// SubscribeOnce registers a handler for a topic that will be executed only
// once. After the handler is invoked, the subscription is automatically
// removed.
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
