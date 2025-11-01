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

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

// TestRedisBus_PublishSubscribe tests the basic publish-subscribe functionality of the RedisBus.
func TestRedisBus_PublishSubscribe(t *testing.T) {
	// Skip this test if Redis is not available.
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx := context.Background()
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		t.Skip("Redis not available, skipping test.")
	}

	bus := NewRedisBus[string](redisClient)
	var wg sync.WaitGroup
	wg.Add(1)

	var receivedMsg string
	unsub := bus.Subscribe("test-topic", func(msg string) {
		receivedMsg = msg
		wg.Done()
	})
	defer unsub()

	// Allow some time for the subscription to be registered in Redis.
	time.Sleep(100 * time.Millisecond)

	bus.Publish("test-topic", "hello")

	wg.Wait()
	assert.Equal(t, "hello", receivedMsg)
}

// TestRedisBus_SubscribeOnce tests that a handler is invoked only once for a given topic.
func TestRedisBus_SubscribeOnce(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx := context.Background()
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		t.Skip("Redis not available, skipping test.")
	}

	bus := NewRedisBus[string](redisClient)
	var count int
	var wg sync.WaitGroup
	wg.Add(1)

	unsub := bus.SubscribeOnce("test-topic-once", func(msg string) {
		count++
		wg.Done()
	})
	defer unsub()

	time.Sleep(100 * time.Millisecond)

	bus.Publish("test-topic-once", "hello once")
	wg.Wait()

	// Publish again to ensure the handler is not called a second time.
	bus.Publish("test-topic-once", "hello again")
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 1, count)
}
