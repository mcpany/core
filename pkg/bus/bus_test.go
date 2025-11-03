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
	"sync"
	"testing"
	"time"

	"github.com/mcpany/core/proto/bus"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)


func TestIntegration(t *testing.T) {
	provider, err := NewBusProvider(nil)
	assert.NoError(t, err)
	defer provider.Shutdown()

	bus := GetBus[string](provider, "test-topic")

	var wg sync.WaitGroup
	wg.Add(1)

	bus.SubscribeOnce(context.Background(), "test-message", func(msg string) {
		assert.Equal(t, "hello", msg)
		wg.Done()
	})

	bus.Publish(context.Background(), "test-message", "hello")

	wg.Wait()
}


func TestRedisBus_SubscribeOnce(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		t.Skip("Redis is not available")
	}

	messageBus := bus.MessageBus_builder{}.Build()
	redisBus := bus.RedisBus_builder{}.Build()
	redisBus.SetAddress("localhost:6379")
	messageBus.SetRedis(redisBus)

	provider, err := NewBusProvider(messageBus)
	assert.NoError(t, err)

	bus := GetBus[string](provider, "test-topic")

	var wg sync.WaitGroup
	wg.Add(1)

	var receivedMessages []string
	bus.SubscribeOnce(context.Background(), "test-message", func(msg string) {
		receivedMessages = append(receivedMessages, msg)
		wg.Done()
	})

	bus.Publish(context.Background(), "test-message", "hello")
	bus.Publish(context.Background(), "test-message", "world")

	wg.Wait()

	assert.Len(t, receivedMessages, 1)
	assert.Equal(t, "hello", receivedMessages[0])
}

func TestRedisBus_Unsubscribe(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		t.Skip("Redis is not available")
	}

	messageBus := bus.MessageBus_builder{}.Build()
	redisBus := bus.RedisBus_builder{}.Build()
	redisBus.SetAddress("localhost:6379")
	messageBus.SetRedis(redisBus)

	provider, err := NewBusProvider(messageBus)
	assert.NoError(t, err)

	bus := GetBus[string](provider, "test-topic")

	var receivedMessages []string
	unsubscribe := bus.Subscribe(context.Background(), "test-message", func(msg string) {
		receivedMessages = append(receivedMessages, msg)
	})

	bus.Publish(context.Background(), "test-message", "hello")
	time.Sleep(100 * time.Millisecond) // Allow time for the message to be processed

	unsubscribe()

	bus.Publish(context.Background(), "test-message", "world")
	time.Sleep(100 * time.Millisecond) // Allow time for the message to be processed

	assert.Len(t, receivedMessages, 1)
	assert.Equal(t, "hello", receivedMessages[0])
}

func TestRedisBus_Concurrent(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		t.Skip("Redis is not available")
	}

	messageBus := bus.MessageBus_builder{}.Build()
	redisBus := bus.RedisBus_builder{}.Build()
	redisBus.SetAddress("localhost:6379")
	messageBus.SetRedis(redisBus)

	provider, err := NewBusProvider(messageBus)
	assert.NoError(t, err)

	bus := GetBus[string](provider, "test-topic")

	numSubscribers := 10
	numMessages := 100
	var wg sync.WaitGroup
	wg.Add(numSubscribers * numMessages)

	var receivedMessages [][]string
	for i := 0; i < numSubscribers; i++ {
		receivedMessages = append(receivedMessages, []string{})
		go func(i int) {
			bus.Subscribe(context.Background(), "test-message", func(msg string) {
				receivedMessages[i] = append(receivedMessages[i], msg)
				wg.Done()
			})
		}(i)
	}

	for i := 0; i < numMessages; i++ {
		bus.Publish(context.Background(), "test-message", "hello")
	}

	wg.Wait()

	for i := 0; i < numSubscribers; i++ {
		assert.Len(t, receivedMessages[i], numMessages)
	}
}
