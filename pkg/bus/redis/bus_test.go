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
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"testing"
	"time"

	bus_pb "github.com/mcpany/core/proto/bus"
	"github.com/go-redis/redismock/v9"
	"github.com/mcpany/core/pkg/logging"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

import (
	"os"
	// ... other imports
)

func setupRedisIntegrationTest(t *testing.T) *redis.Client {
	t.Helper()
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		t.Skip("Redis is not available")
	}
	return client
}

func TestRedisBus_Publish(t *testing.T) {
	client, mock := redismock.NewClientMock()
	bus := NewWithClient[string](client)

	msg, _ := json.Marshal("hello")
	mock.ExpectPublish("test", msg).SetVal(1)
	err := bus.Publish(context.Background(), "test", "hello")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisBus_Publish_MarshalError(t *testing.T) {
	client, _ := redismock.NewClientMock()
	bus := NewWithClient[chan int](client)

	err := bus.Publish(context.Background(), "test", make(chan int))
	assert.Error(t, err)
	assert.IsType(t, &json.UnsupportedTypeError{}, err)
}

func TestRedisBus_Publish_RedisError(t *testing.T) {
	client, mock := redismock.NewClientMock()
	bus := NewWithClient[string](client)

	msg, _ := json.Marshal("hello")
	mock.ExpectPublish("test", msg).SetErr(redis.ErrClosed)
	err := bus.Publish(context.Background(), "test", "hello")
	assert.Error(t, err)
	assert.Equal(t, redis.ErrClosed, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisBus_Subscribe(t *testing.T) {
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)
	topic := "test-subscribe"

	var wg sync.WaitGroup
	wg.Add(1)

	unsub := bus.Subscribe(context.Background(), topic, func(msg string) {
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
}



func TestRedisBus_SubscribeOnce_HandlerPanic(t *testing.T) {
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)
	topic := "handler-panic-topic"

	var wg sync.WaitGroup
	wg.Add(1)

	unsub := bus.SubscribeOnce(context.Background(), topic, func(msg string) {
		defer wg.Done()
		panic("test panic")
	})
	defer unsub()

	require.Eventually(t, func() bool {
		subs := client.PubSubNumSub(context.Background(), topic).Val()
		return len(subs) > 0 && subs[topic] == 1
	}, 1*time.Second, 10*time.Millisecond, "subscriber did not appear")

	err := bus.Publish(context.Background(), topic, "hello")
	assert.NoError(t, err)

	wg.Wait()
}

func TestRedisBus_MultipleUnsubCalls(t *testing.T) {
	client, _ := redismock.NewClientMock()
	bus := NewWithClient[string](client)

	unsub := bus.Subscribe(context.Background(), "test-topic", func(msg string) {
		// No-op handler
	})

	// Call unsub multiple times and assert that it doesn't panic
	assert.NotPanics(t, unsub)
	assert.NotPanics(t, unsub)
}


func TestRedisBus_SubscribeOnce_ConcurrentPublish(t *testing.T) {
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)
	topic := "once-concurrent-publish"

	handlerCalled := make(chan string, 10) // Buffer to avoid blocking publishers

	unsub := bus.SubscribeOnce(context.Background(), topic, func(msg string) {
		handlerCalled <- msg
	})
	defer unsub()

	require.Eventually(t, func() bool {
		subs := client.PubSubNumSub(context.Background(), topic).Val()
		return len(subs) > 0 && subs[topic] == 1
	}, 1*time.Second, 10*time.Millisecond, "subscriber did not appear")

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bus.Publish(context.Background(), topic, "message")
		}()
	}
	wg.Wait()

	// Wait for the message to be processed and the subscription to be closed.
	// We check for the subscription to be gone to ensure SubscribeOnce's unsub ran.
	require.Eventually(t, func() bool {
		subs := client.PubSubNumSub(context.Background(), topic).Val()
		return len(subs) == 0 || subs[topic] == 0
	}, 1*time.Second, 10*time.Millisecond, "subscriber did not disappear after message")

	assert.Len(t, handlerCalled, 1, "handler should have been called exactly once")
	close(handlerCalled)
}

func TestRedisBus_Subscribe_UnmarshalError(t *testing.T) {
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)

	// Capture log output
	var logBuffer bytes.Buffer
	logging.ForTestsOnlyResetLogger()
	logging.Init(slog.LevelDebug, &logBuffer)
	defer logging.ForTestsOnlyResetLogger()

	handlerCalled := make(chan bool, 1)

	unsub := bus.Subscribe(context.Background(), "test-unmarshal-error", func(msg string) {
		handlerCalled <- true
	})
	defer unsub()

	require.Eventually(t, func() bool {
		subs := client.PubSubNumSub(context.Background(), "test-unmarshal-error").Val()
		return len(subs) > 0 && subs["test-unmarshal-error"] == 1
	}, 1*time.Second, 10*time.Millisecond, "subscriber did not appear")

	err := client.Publish(context.Background(), "test-unmarshal-error", "invalid json").Err()
	assert.NoError(t, err)

	select {
	case <-handlerCalled:
		t.Fatal("handler should not have been called")
	case <-time.After(200 * time.Millisecond):
		// Test passed, handler was not called
	}

	assert.Contains(t, logBuffer.String(), "Failed to unmarshal message")
}

// TestRedisBus_SubscribeOnce tests that a handler for a topic is only called once.
// Note: Go's coverage tool may report 0% coverage for this function. This is a
// known issue with the tool's ability to track coverage in goroutines,
// especially in short-lived test scenarios. The test is valid and does
// exercise the code path.
func TestRedisBus_SubscribeOnce(t *testing.T) {
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)
	topic := "test-subscribe-once"

	var wg sync.WaitGroup
	wg.Add(1)

	handlerCalled := make(chan string, 1)

	unsub := bus.SubscribeOnce(context.Background(), topic, func(msg string) {
		handlerCalled <- msg
		wg.Done()
	})
	defer unsub()

	require.Eventually(t, func() bool {
		subs := client.PubSubNumSub(context.Background(), topic).Val()
		return len(subs) > 0 && subs[topic] == 1
	}, 1*time.Second, 10*time.Millisecond, "subscriber did not appear")

	err := bus.Publish(context.Background(), topic, "first message")
	assert.NoError(t, err)

	wg.Wait()

	// The subscription should be automatically removed after one message
	require.Eventually(t, func() bool {
		subs := client.PubSubNumSub(context.Background(), topic).Val()
		return len(subs) == 0 || subs[topic] == 0
	}, 1*time.Second, 10*time.Millisecond, "subscriber did not disappear")

	// Publish another message to ensure the handler is not called again
	err = bus.Publish(context.Background(), topic, "second message")
	assert.NoError(t, err)

	select {
	case msg := <-handlerCalled:
		assert.Equal(t, "first message", msg)
	default:
		t.Fatal("handler should have been called once")
	}

	// Ensure no more messages are received
	select {
	case msg := <-handlerCalled:
		t.Fatalf("handler should not have been called again, but received: %s", msg)
	case <-time.After(100 * time.Millisecond):
		// Test passed, no second message received
	}
}

func TestRedisBus_Unsubscribe(t *testing.T) {
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)

	handlerCalled := make(chan bool, 1)

	unsub := bus.Subscribe(context.Background(), "test-unsubscribe", func(msg string) {
		handlerCalled <- true
	})

	require.Eventually(t, func() bool {
		subs := client.PubSubNumSub(context.Background(), "test-unsubscribe").Val()
		return len(subs) > 0 && subs["test-unsubscribe"] == 1
	}, 1*time.Second, 10*time.Millisecond, "subscriber did not appear")

	unsub()

	require.Eventually(t, func() bool {
		subs := client.PubSubNumSub(context.Background(), "test-unsubscribe").Val()
		return len(subs) == 0 || subs["test-unsubscribe"] == 0
	}, 1*time.Second, 10*time.Millisecond, "subscriber did not disappear")

	err := bus.Publish(context.Background(), "test-unsubscribe", "hello")
	assert.NoError(t, err)

	select {
	case <-handlerCalled:
		t.Fatal("handler should not have been called after unsubscribe")
	case <-time.After(200 * time.Millisecond):
		// Test passed
	}
}

func TestRedisBus_New(t *testing.T) {
	redisBus := bus_pb.RedisBus_builder{
		Address:  proto.String("localhost:6379"),
		Password: proto.String("password"),
		Db:       proto.Int32(1),
	}.Build()

	bus := New[string](redisBus)
	assert.NotNil(t, bus)
	assert.NotNil(t, bus.client)
	options := bus.client.Options()
	assert.Equal(t, "localhost:6379", options.Addr)
	assert.Equal(t, "password", options.Password)
	assert.Equal(t, 1, options.DB)
}

func TestRedisBus_New_NilConfig(t *testing.T) {
	bus := New[string](nil)
	assert.NotNil(t, bus)
	assert.NotNil(t, bus.client)
	options := bus.client.Options()
	assert.Equal(t, "localhost:6379", options.Addr)
	assert.Equal(t, "", options.Password)
	assert.Equal(t, 0, options.DB)
}

func TestRedisBus_ConcurrentSubscribeAndUnsubscribe(t *testing.T) {
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)
	topic := "concurrent-topic"

	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			unsub := bus.Subscribe(context.Background(), topic, func(msg string) {
				// No-op handler
			})
			time.Sleep(10 * time.Millisecond) // Give a chance for other goroutines to run
			unsub()
		}()
	}

	wg.Wait()
}

func TestRedisBus_Subscribe_Resubscribe_ShouldReplacePreviousSubscription(t *testing.T) {
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)
	topic := "test-resubscribe-replace"

	unsub1 := bus.Subscribe(context.Background(), topic, func(msg string) {})

	require.Eventually(t, func() bool {
		subs, err := client.PubSubNumSub(context.Background(), topic).Result()
		require.NoError(t, err)
		if val, ok := subs[topic]; ok {
			return val == 1
		}
		return false
	}, 1*time.Second, 10*time.Millisecond, "first subscriber did not appear")

	unsub2 := bus.Subscribe(context.Background(), topic, func(msg string) {})
	defer unsub2()

	var subCount int64
	require.Eventually(t, func() bool {
		subs, err := client.PubSubNumSub(context.Background(), topic).Result()
		require.NoError(t, err)
		if val, ok := subs[topic]; ok {
			subCount = val
			return subCount == 1
		}
		return false
	}, 1*time.Second, 10*time.Millisecond, "subscriber count should be 1 after resubscribe, but was %d", subCount)

	unsub1()
}

func TestRedisBus_Subscribe_HandlerPanic(t *testing.T) {
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)
	topic := "test-subscribe-panic"

	handlerCalled := make(chan bool, 2)

	unsub := bus.Subscribe(context.Background(), topic, func(msg string) {
		handlerCalled <- true
		if len(handlerCalled) == 1 {
			panic("handler panic")
		}
		assert.Equal(t, "second message", msg)
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

	// Even if the handler panics, the subscription should remain active
	bus.Publish(context.Background(), topic, "first message")
	bus.Publish(context.Background(), topic, "second message")

	// Wait for both messages to be processed
	<-handlerCalled
	<-handlerCalled

	assert.Len(t, handlerCalled, 0, "handler should have been called twice")
}

func TestRedisBus_Subscribe_ContextCancellation(t *testing.T) {
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)
	topic := "test-context-cancellation"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)

	unsub := bus.Subscribe(ctx, topic, func(msg string) {
		// This handler should not be called after the context is canceled.
		t.Error("handler called after context cancellation")
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

	cancel()

	// Allow some time for the cancellation to propagate
	time.Sleep(100 * time.Millisecond)

	err := bus.Publish(context.Background(), topic, "hello")
	assert.NoError(t, err)

	// We expect the handler not to be called, so we don't wait for a WaitGroup.
	// Instead, we just wait a bit to see if the handler is called.
	time.Sleep(100 * time.Millisecond)
}

func TestRedisBus_Subscribe_AlreadyCancelledContext(t *testing.T) {
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)
	topic := "already-cancelled-context"

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	unsub := bus.Subscribe(ctx, topic, func(msg string) {
		t.Error("handler should not be called")
	})
	defer unsub()

	// The subscription should not be created in the first place, but we check regardless
	require.Eventually(t, func() bool {
		subs, err := client.PubSubNumSub(context.Background(), topic).Result()
		require.NoError(t, err)
		if val, ok := subs[topic]; ok {
			return val == 0
		}
		return true // No subscribers is the desired state
	}, 1*time.Second, 10*time.Millisecond, "subscriber should not be created with a cancelled context")

	err := bus.Publish(context.Background(), topic, "hello")
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
}

func TestRedisBus_Close(t *testing.T) {
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)
	topic := "test-close"

	// Subscribe to create a pubsub connection
	unsub := bus.Subscribe(context.Background(), topic, func(msg string) {})

	// Check that the subscription is active
	require.Eventually(t, func() bool {
		subs := client.PubSubNumSub(context.Background(), topic).Val()
		return len(subs) > 0 && subs[topic] == 1
	}, 1*time.Second, 10*time.Millisecond, "subscriber did not appear")

	// Close the bus
	err := bus.Close()
	assert.NoError(t, err)

	// Verify that the client is closed by trying to use it
	err = client.Ping(context.Background()).Err()
	assert.Error(t, err)
	assert.Equal(t, redis.ErrClosed, err)

	// After closing the bus, the pubsubs map should be empty.
	bus.mu.RLock()
	assert.Empty(t, bus.pubsubs)
	bus.mu.RUnlock()

	// Unsubscribing after closing should not panic
	assert.NotPanics(t, unsub)
}

func TestRedisBus_UnsubscribeFromHandler(t *testing.T) {
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)
	topic := "unsubscribe-from-handler"

	var wg sync.WaitGroup
	wg.Add(1)

	var unsub func()
	unsub = bus.Subscribe(context.Background(), topic, func(msg string) {
		unsub()
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

func TestRedisBus_SubscribeOnce_CancelledContext(t *testing.T) {
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)
	topic := "once-cancelled-context"

	ctx, cancel := context.WithCancel(context.Background())

	unsub := bus.SubscribeOnce(ctx, topic, func(msg string) {
		t.Error("handler should not be called")
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

	cancel()

	require.Eventually(t, func() bool {
		subs, err := client.PubSubNumSub(context.Background(), topic).Result()
		require.NoError(t, err)
		if val, ok := subs[topic]; ok {
			return val == 0
		}
		return true
	}, 1*time.Second, 10*time.Millisecond, "subscriber did not disappear after context cancellation")

	err := bus.Publish(context.Background(), topic, "hello")
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
}
