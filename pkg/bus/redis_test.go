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
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func setupRedisBus[T any](t *testing.T) Bus[T] {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		t.Skip("Redis server not available. Skipping test.")
	}

	provider := NewBusProvider(newMessageBusConfig(t, `{"redis":{"address":"localhost:6379"}}`))
	return GetBus[T](provider, "test-topic")
}

func TestRedisBus_PublishSubscribe(t *testing.T) {
	bus := setupRedisBus[string](t)
	topic := "test-topic"
	message := "hello-world"
	received := make(chan string, 1)

	unsub := bus.Subscribe(topic, func(msg string) {
		received <- msg
	})
	defer unsub()

	bus.Publish(topic, message)

	select {
	case msg := <-received:
		assert.Equal(t, message, msg)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for message")
	}
}

func TestRedisBus_SubscribeOnce(t *testing.T) {
	bus := setupRedisBus[string](t)
	topic := "test-topic-once"
	message := "hello-world-once"
	received := make(chan string, 1)

	unsub := bus.SubscribeOnce(topic, func(msg string) {
		received <- msg
	})
	defer unsub()

	bus.Publish(topic, message)
	bus.Publish(topic, "should-not-be-received")

	select {
	case msg := <-received:
		assert.Equal(t, message, msg)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for message")
	}

	// Make sure the second message wasn't received
	select {
	case <-received:
		t.Fatal("received more than one message")
	case <-time.After(100 * time.Millisecond):
		// This is expected
	}
}
