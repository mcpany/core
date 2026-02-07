// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package integration tests the Redis bus implementation.
package integration

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/bus/redis"
	bustypes "github.com/mcpany/core/proto/bus"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func waitForSubscribers(t *testing.T, client *goredis.Client, topic string, expected int) {
	require.Eventually(t, func() bool {
		subs := client.PubSubNumSub(context.Background(), topic).Val()
		// Check that the subscriber count is at least the expected number
		return subs[topic] >= int64(expected)
	}, 5*time.Second, 100*time.Millisecond, "timed out waiting for subscribers on topic %s", topic)
}

func TestRedisBus_Integration_Subscribe(t *testing.T) {
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skip("Skipping TestRedisBus_Integration_Subscribe in CI environment")
	}
	redisAddr, cleanup := StartRedisContainer(t)
	defer cleanup()

	client := goredis.NewClient(&goredis.Options{
		Addr: redisAddr,
	})

	bus := redis.NewWithClient[string](client)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	topic := "test-topic"
	msg := "hello"

	var wg sync.WaitGroup
	wg.Add(1)

	handler := func(m string) {
		assert.Equal(t, msg, m)
		wg.Done()
	}

	unsubscribe := bus.Subscribe(ctx, topic, handler)
	defer unsubscribe()

	// Wait for subscriber to connect
	waitForSubscribers(t, client, topic, 1)

	err := bus.Publish(ctx, topic, msg)
	assert.NoError(t, err)

	wg.Wait()
}

func TestRedisBus_Integration_SubscribeOnce(t *testing.T) {
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skip("Skipping TestRedisBus_Integration_SubscribeOnce in CI environment")
	}
	redisAddr, cleanup := StartRedisContainer(t)
	defer cleanup()

	client := goredis.NewClient(&goredis.Options{
		Addr: redisAddr,
	})

	bus := redis.NewWithClient[string](client)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	topic := "test-topic-once"
	msg := "hello once"

	var wg sync.WaitGroup
	wg.Add(1)

	handler := func(m string) {
		assert.Equal(t, msg, m)
		wg.Done()
	}

	unsubscribe := bus.SubscribeOnce(ctx, topic, handler)
	defer unsubscribe()

	// Wait for subscriber to connect
	waitForSubscribers(t, client, topic, 1)

	err := bus.Publish(ctx, topic, msg)
	assert.NoError(t, err)

	wg.Wait()
}

func TestBusProvider_Integration_Redis(t *testing.T) {
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skip("Skipping TestBusProvider_Integration_Redis in CI environment")
	}
	redisAddr, cleanup := StartRedisContainer(t)
	defer cleanup()

	messageBus := bustypes.MessageBus_builder{}.Build()
	redisBus := bustypes.RedisBus_builder{}.Build()
	redisBus.SetAddress(redisAddr)
	messageBus.SetRedis(redisBus)

	provider, err := bus.NewProvider(messageBus)
	assert.NoError(t, err)

	bus1, _ := bus.GetBus[string](provider, "strings")
	bus2, _ := bus.GetBus[int](provider, "ints")
	bus3, _ := bus.GetBus[string](provider, "strings")

	assert.NotNil(t, bus1)
	assert.NotNil(t, bus2)
	assert.Same(t, bus1, bus3)
}

func TestRedisBus_Integration_Unsubscribe(t *testing.T) {
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skip("Skipping TestRedisBus_Integration_Unsubscribe in CI environment")
	}
	redisAddr, cleanup := StartRedisContainer(t)
	defer cleanup()

	client := goredis.NewClient(&goredis.Options{
		Addr: redisAddr,
	})

	redisBus := redis.NewWithClient[string](client)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	topic := "test-topic"
	msg1 := "hello"
	msg2 := "world"

	var receivedMessages []string
	var mu sync.Mutex

	handler := func(m string) {
		mu.Lock()
		defer mu.Unlock()
		receivedMessages = append(receivedMessages, m)
	}

	unsubscribe := redisBus.Subscribe(ctx, topic, handler)

	// Wait for subscriber to connect
	waitForSubscribers(t, client, topic, 1)

	err := redisBus.Publish(ctx, topic, msg1)
	assert.NoError(t, err)

	// Wait for message 1
	time.Sleep(100 * time.Millisecond)

	unsubscribe()

	err = redisBus.Publish(ctx, topic, msg2)
	assert.NoError(t, err)

	// Wait for message 2 (should not be received)
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	messages := receivedMessages
	mu.Unlock()

	assert.Len(t, messages, 1)
	assert.Equal(t, msg1, messages[0])
}

func TestRedisBus_Integration_Concurrent(t *testing.T) {
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skip("Skipping TestRedisBus_Integration_Concurrent in CI environment")
	}
	redisAddr, cleanup := StartRedisContainer(t)
	defer cleanup()

	client := goredis.NewClient(&goredis.Options{
		Addr: redisAddr,
	})

	redisBus := redis.NewWithClient[string](client)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	topic := "test-topic-concurrent"
	numSubscribers := 10
	numMessages := 50

	var wg sync.WaitGroup
	wg.Add(numSubscribers * numMessages)

	var receivedMessages [][]string
	var mu sync.Mutex

    receivedMessages = make([][]string, numSubscribers)
    for i := 0; i < numSubscribers; i++ {
        receivedMessages[i] = make([]string, 0, numMessages)
    }

	for i := 0; i < numSubscribers; i++ {
		go func(subIdx int) {
			redisBus.Subscribe(ctx, topic, func(msg string) {
				mu.Lock()
				receivedMessages[subIdx] = append(receivedMessages[subIdx], msg)
				mu.Unlock()
				wg.Done()
			})
		}(i)
	}

	// Wait for all subscribers to connect
	waitForSubscribers(t, client, topic, numSubscribers)

	for i := 0; i < numMessages; i++ {
		err := redisBus.Publish(ctx, topic, "msg")
		assert.NoError(t, err)
	}

	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	for i := 0; i < numSubscribers; i++ {
		assert.Len(t, receivedMessages[i], numMessages)
	}
}
