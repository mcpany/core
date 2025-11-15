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

package memory

import (
	"context"
	"sync"
	"sync/atomic"
	"bytes"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/logging"
	"github.com/stretchr/testify/assert"
)

func TestDefaultBus(t *testing.T) {
	t.Run("Publish and Subscribe", func(t *testing.T) {
		bus := New[string]()
		var wg sync.WaitGroup
		wg.Add(1)

		bus.Subscribe(context.Background(), "test", func(msg string) {
			assert.Equal(t, "hello", msg)
			wg.Done()
		})

		bus.Publish(context.Background(), "test", "hello")
		wg.Wait()
	})

	t.Run("SubscribeOnce", func(t *testing.T) {
		bus := New[string]()
		var wg sync.WaitGroup
		var callCount int32
		wg.Add(1)

		bus.SubscribeOnce(context.Background(), "test", func(msg string) {
			atomic.AddInt32(&callCount, 1)
			assert.Equal(t, "hello", msg)
			wg.Done()
		})

		bus.Publish(context.Background(), "test", "hello")
		// This second publish should not be received by the handler
		bus.Publish(context.Background(), "test", "world")
		wg.Wait()

		// Allow some time for any extra messages to be processed
		time.Sleep(10 * time.Millisecond)
		assert.Equal(t, int32(1), atomic.LoadInt32(&callCount), "handler should only be called once")
	})

	t.Run("Unsubscribe", func(t *testing.T) {
		bus := New[string]()
		received := false

		unsub := bus.Subscribe(context.Background(), "test", func(msg string) {
			received = true
		})

		unsub()
		bus.Publish(context.Background(), "test", "hello")
		time.Sleep(10 * time.Millisecond) // Give it a moment to process
		assert.False(t, received)
	})
}

func TestDefaultBus_Concurrent(t *testing.T) {
	bus := New[int]()
	topic := "concurrent_topic"
	numSubscribers := 10
	numPublishers := 100
	var receivedCount int32

	var wg sync.WaitGroup
	expectedReceives := numSubscribers * numPublishers
	wg.Add(expectedReceives)

	for i := 0; i < numSubscribers; i++ {
		unsub := bus.Subscribe(context.Background(), topic, func(msg int) {
			atomic.AddInt32(&receivedCount, 1)
			wg.Done()
		})
		defer unsub()
	}

	for i := 0; i < numPublishers; i++ {
		go bus.Publish(context.Background(), topic, i)
	}

	// Wait for all messages to be received, with a timeout.
	timeout := time.After(5 * time.Second)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All goroutines completed.
	case <-timeout:
		t.Fatalf("Timed out waiting for messages. Got %d of %d.", atomic.LoadInt32(&receivedCount), expectedReceives)
	}

	assert.Equal(t, int32(expectedReceives), atomic.LoadInt32(&receivedCount))
}

func TestDefaultBus_PublishTimeout(t *testing.T) {
	// 1. Set up a logger to capture output
	var logBuffer bytes.Buffer
	logging.ForTestsOnlyResetLogger()
	logging.Init(slog.LevelWarn, &logBuffer)

	// 2. Create a bus with a very short timeout
	bus := New[string]()
	bus.publishTimeout = 1 * time.Millisecond

	// 3. Create a subscriber with a full channel that will not receive messages
	var wg sync.WaitGroup
	wg.Add(1)
	unsub := bus.Subscribe(context.Background(), "test_timeout", func(msg string) {
		// This handler will block, preventing further messages from being processed
		wg.Wait()
	})
	defer unsub()

	// 4. Publish messages to block the subscriber and fill the channel buffer.
	// The first message will be consumed by the handler, which will block.
	// The next 128 messages will fill the channel's buffer.
	for i := 0; i < 129; i++ {
		bus.Publish(context.Background(), "test_timeout", "fill")
	}

	// 5. Publish the message that should time out
	bus.Publish(context.Background(), "test_timeout", "should_drop")

	// 6. Check the log output for the dropped message warning
	assert.Eventually(t, func() bool {
		return strings.Contains(logBuffer.String(), "Message dropped on topic")
	}, 1*time.Second, 10*time.Millisecond)

	// 7. Clean up the blocking handler
	wg.Done()
}

func TestDefaultBus_SubscribeOnce_Unsubscribe(t *testing.T) {
	bus := New[string]()
	handlerCalled := false

	unsub := bus.SubscribeOnce(context.Background(), "test_unsub_once", func(msg string) {
		handlerCalled = true
	})

	// Unsubscribe immediately, before any message is published
	unsub()

	// Publish a message to the topic
	bus.Publish(context.Background(), "test_unsub_once", "hello")

	// Wait a moment to ensure the handler is not called
	time.Sleep(10 * time.Millisecond)

	assert.False(t, handlerCalled, "handler should not be called after unsubscribing")
}
