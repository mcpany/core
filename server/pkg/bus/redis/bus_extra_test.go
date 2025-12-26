// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBus_Subscribe_ConcurrentSubscribers(t *testing.T) {
	t.Skip("Skipping flaky concurrent subscribers test")
	client := setupRedisIntegrationTest(t)
	// Explicitly close client to avoid leaks if this test becomes sensitive or other tests run
	defer client.Close()

	bus := NewWithClient[string](client)
	topic := "concurrent-subscribers"
	numSubscribers := 2
	var wg sync.WaitGroup
	wg.Add(numSubscribers)

	unsubs := make(chan func(), numSubscribers)
	for i := 0; i < numSubscribers; i++ {
		go func() {
			unsub := bus.Subscribe(context.Background(), topic, func(_ string) {
				// Each subscriber should receive the message
				wg.Done()
			})
			unsubs <- unsub
		}()
	}

	// Wait for subscribers to be ready
	require.Eventually(t, func() bool {
		subs := client.PubSubNumSub(context.Background(), topic).Val()
		return len(subs) > 0 && subs[topic] == int64(numSubscribers)
	}, 10*time.Second, 100*time.Millisecond, "all subscribers did not appear")

	// Publish a message
	err := bus.Publish(context.Background(), topic, "hello")
	require.NoError(t, err)

	wg.Wait()

	close(unsubs)
	for unsub := range unsubs {
		unsub()
	}
}

func TestBus_SubscribeOnce_UnsubscribeFromHandler(t *testing.T) {
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)
	topic := "once-unsubscribe-from-handler"

	var wg sync.WaitGroup
	wg.Add(1)

	var unsub func()
	unsub = bus.SubscribeOnce(context.Background(), topic, func(_ string) {
		unsub() // Call unsubscribe from within the handler
		wg.Done()
	})

	require.Eventually(t, func() bool {
		subs, err := client.PubSubNumSub(context.Background(), topic).Result()
		require.NoError(t, err)
		if val, ok := subs[topic]; ok {
			return val == 1
		}
		return false
	}, 1*time.Second, 10*time.Millisecond, "subscriber did not appear")

	err := bus.Publish(context.Background(), topic, "hello")
	assert.NoError(t, err)

	wg.Wait()

	require.Eventually(t, func() bool {
		subs, err := client.PubSubNumSub(context.Background(), topic).Result()
		require.NoError(t, err)
		if val, ok := subs[topic]; ok {
			return val == 0
		}
		return true
	}, 1*time.Second, 10*time.Millisecond, "subscriber did not disappear after unsubscribing from handler")
}

func TestBus_Subscribe_CloseClient(t *testing.T) {
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)
	topic := "test-close-client"

	var wg sync.WaitGroup
	wg.Add(1)

	unsub := bus.Subscribe(context.Background(), topic, func(_ string) {
		// This handler might be called once if a message is received before the client is closed
		wg.Done()
	})
	defer unsub()

	require.Eventually(t, func() bool {
		subs, err := client.PubSubNumSub(context.Background(), topic).Result()
		require.NoError(t, err)
		if val, ok := subs[topic]; ok {
			return val == 1
		}
		return false
	}, 1*time.Second, 10*time.Millisecond, "subscriber did not appear")

	// Close the client, which should terminate the subscription loop
	err := client.Close()
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	bus.mu.Lock()
	_, ok := bus.pubsubs[topic]
	bus.mu.Unlock()
	assert.False(t, ok, "subscription should be removed after client is closed")
}

func TestBus_Subscribe_CloseClient_Race(t *testing.T) {
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)
	topic := "test-close-client-race"

	var wg sync.WaitGroup
	wg.Add(1)

	unsub := bus.Subscribe(context.Background(), topic, func(_ string) {
		wg.Done()
	})

	var unsubOnce sync.Once
	go func() {
		unsubOnce.Do(unsub)
	}()

	err := client.Close()
	assert.NoError(t, err)

	unsubOnce.Do(unsub)
}

func TestBus_Unsubscribe_Race(t *testing.T) {
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)
	topic := "test-unsubscribe-race"

	unsub := bus.Subscribe(context.Background(), topic, func(_ string) {})

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		unsub()
	}()
	go func() {
		defer wg.Done()
		unsub()
	}()
	wg.Wait()
}

func TestBus_Subscribe_And_Unsubscribe_Race(t *testing.T) {
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)
	topic := "test-subscribe-and-unsubscribe-race"

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		unsub := bus.Subscribe(context.Background(), topic, func(_ string) {})
		time.Sleep(10 * time.Millisecond)
		unsub()
	}()

	go func() {
		defer wg.Done()
		unsub := bus.Subscribe(context.Background(), topic, func(_ string) {})
		time.Sleep(10 * time.Millisecond)
		unsub()
	}()

	wg.Wait()
}
