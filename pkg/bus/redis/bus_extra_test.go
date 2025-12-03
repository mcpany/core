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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisBus_Subscribe_ConcurrentSubscribers(t *testing.T) {
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)
	topic := "concurrent-subscribers-topic"

	handler1Called := make(chan bool, 1)
	handler2Called := make(chan bool, 1)

	unsub1 := bus.Subscribe(context.Background(), topic, func(msg string) {
		handler1Called <- true
	})
	defer unsub1()

	unsub2 := bus.Subscribe(context.Background(), topic, func(msg string) {
		handler2Called <- true
	})
	defer unsub2()

	require.Eventually(t, func() bool {
		subs := client.PubSubNumSub(context.Background(), topic).Val()
		return len(subs) > 0 && subs[topic] == 2
	}, 1*time.Second, 10*time.Millisecond, "subscribers did not appear")

	err := bus.Publish(context.Background(), topic, "hello")
	assert.NoError(t, err)

	<-handler1Called
	<-handler2Called
}

func TestRedisBus_Subscribe_CloseClientDuringSubscription(t *testing.T) {
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)
	topic := "close-client-during-subscription"

	handlerCalled := make(chan bool, 1)

	unsub := bus.Subscribe(context.Background(), topic, func(msg string) {
		handlerCalled <- true
	})
	defer unsub()

	require.Eventually(t, func() bool {
		subs := client.PubSubNumSub(context.Background(), topic).Val()
		return len(subs) > 0 && subs[topic] == 1
	}, 1*time.Second, 10*time.Millisecond, "subscriber did not appear")

	err := client.Close()
	assert.NoError(t, err)

	// The subscription goroutine should exit gracefully.
	// We verify this by checking that it doesn't panic and the test completes.
	time.Sleep(200 * time.Millisecond)

	// A new publish should fail
	err = bus.Publish(context.Background(), topic, "hello")
	assert.Error(t, err)
}

func TestRedisBus_Close(t *testing.T) {
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)

	// Subscribe to a topic to create a pubsub connection
	unsubscribe := bus.Subscribe(context.Background(), "test-topic-close", func(msg string) {})

	err := bus.Close()
	assert.NoError(t, err)

	// After closing, publish should fail
	err = bus.Publish(context.Background(), "test-topic-close", "hello")
	assert.Error(t, err)

	// calling unsubscribe after close should not panic
	assert.NotPanics(t, func() {
		unsubscribe()
	})
}

func TestRedisBus_Close_Error(t *testing.T) {
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)

	// Subscribe to a topic to create a pubsub connection
	bus.Subscribe(context.Background(), "test-topic-close-error", func(msg string) {})

	// Close the underlying client to trigger an error when bus.Close() is called.
	err := client.Close()
	assert.NoError(t, err)

	// This should now return an error because the client is already closed.
	err = bus.Close()
	assert.Error(t, err)
}

func TestRedisBus_Close_ClientCloseError(t *testing.T) {
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)

	// Subscribe to a topic to create a pubsub connection
	bus.Subscribe(context.Background(), "test-topic-client-close-error", func(msg string) {})

	// Close the underlying client to trigger an error when bus.Close() is called.
	err := client.Close()
	assert.NoError(t, err)

	// This should now return an error because the client is already closed.
	err = bus.Close()
	assert.Error(t, err)
}

func TestRedisBus_Close_PubSubCloseError(t *testing.T) {
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)

	// Subscribe to a topic to create a pubsub connection
	bus.Subscribe(context.Background(), "test-topic-pubsub-close-error", func(msg string) {})

	// Manually close the pubsub to trigger an error on the second close
	pubsub := bus.pubsubs["test-topic-pubsub-close-error"]
	err := pubsub.Close()
	assert.NoError(t, err)

	err = bus.Close()
	assert.Error(t, err)
}
