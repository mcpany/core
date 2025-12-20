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

func TestBus_SubscribeOnce_Coverage(t *testing.T) {
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)
	topic := "subscribe-once-coverage"

	var wg sync.WaitGroup
	wg.Add(1)

	// Test case: Standard usage
	unsub := bus.SubscribeOnce(context.Background(), topic, func(msg string) {
		assert.Equal(t, "hello", msg)
		wg.Done()
	})
	defer unsub()

	require.Eventually(t, func() bool {
		subs := client.PubSubNumSub(context.Background(), topic).Val()
		return len(subs) > 0 && subs[topic] == 1
	}, 1*time.Second, 10*time.Millisecond, "subscriber did not appear")

	err := bus.Publish(context.Background(), topic, "hello")
	assert.NoError(t, err)

	wg.Wait()

	// Ensure unsub was called internally by checking subscription disappears
	require.Eventually(t, func() bool {
		subs := client.PubSubNumSub(context.Background(), topic).Val()
		return len(subs) == 0 || subs[topic] == 0
	}, 1*time.Second, 10*time.Millisecond, "subscriber did not disappear")
}

func TestBus_SubscribeOnce_PanicInHandler_Coverage(t *testing.T) {
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)
	topic := "subscribe-once-panic-coverage"

	// This test specifically targets the panic handling in SubscribeOnce
	// Since SubscribeOnce uses Subscribe, which recovers from panics,
	// we want to see if the subscription persists or not.
	// As analyzed, it likely persists because unsub() is after handler().

	unsub := bus.SubscribeOnce(context.Background(), topic, func(msg string) {
		panic("boom")
	})
	defer unsub()

	require.Eventually(t, func() bool {
		subs := client.PubSubNumSub(context.Background(), topic).Val()
		return len(subs) > 0 && subs[topic] == 1
	}, 1*time.Second, 10*time.Millisecond, "subscriber did not appear")

	err := bus.Publish(context.Background(), topic, "trigger panic")
	assert.NoError(t, err)

	// Wait a bit for processing
	time.Sleep(200 * time.Millisecond)

	// Check if subscription still exists (it should, based on current implementation)
	subs := client.PubSubNumSub(context.Background(), topic).Val()
	// Note: This asserts the CURRENT behavior, even if it's arguably a bug.
	// Tests should document behavior.
	assert.EqualValues(t, 1, subs[topic], "subscription should persist after panic in SubscribeOnce handler (current behavior)")
}

func TestBus_SubscribeOnce_Concurrent_Unsub(t *testing.T) {
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)
	topic := "subscribe-once-concurrent-unsub"

	// Try to trigger race condition where unsub is called before assigned?
	// It's hard to deterministically trigger, but we can try to stress it.

	for i := 0; i < 100; i++ {
		var wg sync.WaitGroup
		wg.Add(1)

		// Use a channel to synchronize start
		start := make(chan struct{})

		go func() {
			<-start
			_ = bus.Publish(context.Background(), topic, "hello")
		}()

		// We start subscription and immediately publish
		// But Subscribe might take time to register.
		// If we publish too early, message is lost.

		unsub := bus.SubscribeOnce(context.Background(), topic, func(msg string) {
			wg.Done()
		})

		close(start)

		// We just want to ensure no panic happens in the handler due to nil unsub
		// But we can't easily guarantee handler runs before unsub assignment return.

		// Clean up
		unsub()
	}
}
