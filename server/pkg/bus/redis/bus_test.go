// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redismock/v9"
	bus_pb "github.com/mcpany/core/proto/bus"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"google.golang.org/protobuf/proto"
)

// ... other imports

func setupRedisIntegrationTest(t *testing.T) *redis.Client {
	t.Helper()
	// Use miniredis for a hermetic, dependency-free Redis implementation.
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	t.Cleanup(func() {
		_ = client.Close()
	})
	return client
}

// TestBus_Publish ...
// Summary: TestBus_Publish
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	client, mock := redismock.NewClientMock()
	bus := NewWithClient[string](client)

	msg, _ := json.Marshal("hello")
	mock.ExpectPublish("test", msg).SetVal(1)
	err := bus.Publish(context.Background(), "test", "hello")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestBus_Publish_MarshalError ...
// Summary: TestBus_Publish_MarshalError
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	client, _ := redismock.NewClientMock()
	bus := NewWithClient[chan int](client)

	err := bus.Publish(context.Background(), "test", make(chan int))
	assert.Error(t, err)
	assert.IsType(t, &json.UnsupportedTypeError{}, err)
}

// TestBus_Publish_RedisError ...
// Summary: TestBus_Publish_RedisError
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	client, mock := redismock.NewClientMock()
	bus := NewWithClient[string](client)

	msg, _ := json.Marshal("hello")
	mock.ExpectPublish("test", msg).SetErr(redis.ErrClosed)
	err := bus.Publish(context.Background(), "test", "hello")
	assert.Error(t, err)
	assert.Equal(t, redis.ErrClosed, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestBus_Subscribe ...
// Summary: TestBus_Subscribe
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
	}, 5*time.Second, 10*time.Millisecond, "subscriber did not appear")

	err := bus.Publish(context.Background(), topic, "hello")
	assert.NoError(t, err)

	wg.Wait()
}

// TestBus_SubscribeOnce_HandlerPanic ...
// Summary: TestBus_SubscribeOnce_HandlerPanic
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)
	topic := "handler-panic-topic"

	var wg sync.WaitGroup
	wg.Add(1)

	unsub := bus.SubscribeOnce(context.Background(), topic, func(_ string) {
		defer wg.Done()
		panic("test panic")
	})
	defer unsub()

	require.Eventually(t, func() bool {
		subs := client.PubSubNumSub(context.Background(), topic).Val()
		return len(subs) > 0 && subs[topic] == 1
	}, 5*time.Second, 10*time.Millisecond, "subscriber did not appear")

	err := bus.Publish(context.Background(), topic, "hello")
	assert.NoError(t, err)

	wg.Wait()
}

// TestBus_MultipleUnsubCalls ...
// Summary: TestBus_MultipleUnsubCalls
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	client, _ := redismock.NewClientMock()
	bus := NewWithClient[string](client)

	unsub := bus.Subscribe(context.Background(), "test-topic", func(_ string) {
		// No-op handler
	})

	// Call unsub multiple times and assert that it doesn't panic
	assert.NotPanics(t, unsub)
	assert.NotPanics(t, unsub)
}

// TestBus_SubscribeOnce_Correctness ...
// Summary: TestBus_SubscribeOnce_Correctness
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)
	topic := "correctness-topic"

	callCount := 0
	var mu sync.Mutex

	unsub := bus.SubscribeOnce(context.Background(), topic, func(_ string) {
		mu.Lock()
		callCount++
		mu.Unlock()
	})
	defer unsub()

	require.Eventually(t, func() bool {
		subs := client.PubSubNumSub(context.Background(), topic).Val()
		return len(subs) > 0 && subs[topic] == 1
	}, 5*time.Second, 10*time.Millisecond, "subscriber did not appear")

	for i := 0; i < 5; i++ {
		err := bus.Publish(context.Background(), topic, "hello")
		assert.NoError(t, err)
	}

	time.Sleep(200 * time.Millisecond) // Allow time for messages to be processed

	mu.Lock()
	assert.Equal(t, 1, callCount, "handler should be called exactly once")
	mu.Unlock()
}

// TestBus_SubscribeOnce_ConcurrentPublish ...
// Summary: TestBus_SubscribeOnce_ConcurrentPublish
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
	}, 5*time.Second, 10*time.Millisecond, "subscriber did not appear")

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = bus.Publish(context.Background(), topic, "message")
		}()
	}
	wg.Wait()

	// Wait for the message to be processed and the subscription to be closed.
	// We check for the subscription to be gone to ensure SubscribeOnce's unsub ran.
	require.Eventually(t, func() bool {
		subs := client.PubSubNumSub(context.Background(), topic).Val()
		return len(subs) == 0 || subs[topic] == 0
	}, 5*time.Second, 10*time.Millisecond, "subscriber did not disappear after message")

	assert.Len(t, handlerCalled, 1, "handler should have been called exactly once")
	// close(handlerCalled) - unsafe to close as handler might race if PubSubNumSub returned 0 prematurely
}

// TestBus_Subscribe_UnmarshalError ...
// Summary: TestBus_Subscribe_UnmarshalError
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)

	// Capture log output
	var logBuffer ThreadSafeBuffer
	logging.ForTestsOnlyResetLogger()
	logging.Init(slog.LevelDebug, &logBuffer, "")
	defer logging.ForTestsOnlyResetLogger()

	handlerCalled := make(chan bool, 1)

	unsub := bus.Subscribe(context.Background(), "test-unmarshal-error", func(_ string) {
		handlerCalled <- true
	})
	defer unsub()

	require.Eventually(t, func() bool {
		subs := client.PubSubNumSub(context.Background(), "test-unmarshal-error").Val()
		return len(subs) > 0 && subs["test-unmarshal-error"] == 1
	}, 5*time.Second, 10*time.Millisecond, "subscriber did not appear")

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

// TestBus_Subscribe_NullPayload ...
// Summary: TestBus_Subscribe_NullPayload
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[*string](client)
	topic := "test-null-payload"

	var wg sync.WaitGroup
	wg.Add(1)

	unsub := bus.Subscribe(context.Background(), topic, func(msg *string) {
		assert.Nil(t, msg)
		wg.Done()
	})
	defer unsub()

	require.Eventually(t, func() bool {
		subs := client.PubSubNumSub(context.Background(), topic).Val()
		return len(subs) > 0 && subs[topic] == 1
	}, 5*time.Second, 10*time.Millisecond, "subscriber did not appear")

	// Publish a "null" JSON payload
	err := client.Publish(context.Background(), topic, "null").Err()
	assert.NoError(t, err)

	wg.Wait()
}

// TestBus_SubscribeOnce tests that a handler for a topic is only called once.
// Note: Go's coverage tool may report 0% coverage for this function. This is a
// known issue with the tool's ability to track coverage in goroutines,
// especially in short-lived test scenarios. The test is valid and does
// exercise the code path.
// TestBus_SubscribeOnce tests that a handler for a topic is only called once.
// Summary: TestBus_SubscribeOnce
// Note: Go's coverage tool may report 0% coverage for this function. This is a
// known issue with the tool's ability to track coverage in goroutines,
// especially in short-lived test scenarios. The test is valid and does
// exercise the code path.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
	}, 5*time.Second, 10*time.Millisecond, "subscriber did not appear")

	err := bus.Publish(context.Background(), topic, "first message")
	assert.NoError(t, err)

	wg.Wait()

	// The subscription should be automatically removed after one message
	require.Eventually(t, func() bool {
		subs := client.PubSubNumSub(context.Background(), topic).Val()
		return len(subs) == 0 || subs[topic] == 0
	}, 5*time.Second, 10*time.Millisecond, "subscriber did not disappear")

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

// TestBus_SubscribeOnce_NilHandler ...
// Summary: TestBus_SubscribeOnce_NilHandler
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)
	topic := "test-nil-handler"

	assert.NotPanics(t, func() {
		bus.SubscribeOnce(context.Background(), topic, nil)
	})
}

// TestBus_Unsubscribe ...
// Summary: TestBus_Unsubscribe
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)

	handlerCalled := make(chan bool, 1)

	unsub := bus.Subscribe(context.Background(), "test-unsubscribe", func(_ string) {
		handlerCalled <- true
	})

	require.Eventually(t, func() bool {
		subs := client.PubSubNumSub(context.Background(), "test-unsubscribe").Val()
		return len(subs) > 0 && subs["test-unsubscribe"] == 1
	}, 5*time.Second, 10*time.Millisecond, "subscriber did not appear")

	unsub()

	require.Eventually(t, func() bool {
		subs := client.PubSubNumSub(context.Background(), "test-unsubscribe").Val()
		return len(subs) == 0 || subs["test-unsubscribe"] == 0
	}, 5*time.Second, 10*time.Millisecond, "subscriber did not disappear")

	err := bus.Publish(context.Background(), "test-unsubscribe", "hello")
	assert.NoError(t, err)

	select {
	case <-handlerCalled:
		t.Fatal("handler should not have been called after unsubscribe")
	case <-time.After(200 * time.Millisecond):
		// Test passed
	}
}

// TestBus_New ...
// Summary: TestBus_New
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	redisBus := bus_pb.RedisBus_builder{
		Address:  proto.String("127.0.0.1:6379"),
		Password: proto.String("password"),
		Db:       proto.Int32(1),
	}.Build()

	bus, _ := New[string](redisBus)
	assert.NotNil(t, bus)
	assert.NotNil(t, bus.client)
	options := bus.client.Options()
	assert.Equal(t, "127.0.0.1:6379", options.Addr)
	assert.Equal(t, "password", options.Password)
	assert.Equal(t, 1, options.DB)
}

// TestBus_New_WithValidConfig ...
// Summary: TestBus_New_WithValidConfig
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	redisBus := bus_pb.RedisBus_builder{
		Address:  proto.String("127.0.0.1:6380"),
		Password: proto.String("testpassword"),
		Db:       proto.Int32(2),
	}.Build()

	bus, _ := New[string](redisBus)
	assert.NotNil(t, bus)
	assert.NotNil(t, bus.client)
	options := bus.client.Options()
	assert.Equal(t, "127.0.0.1:6380", options.Addr)
	assert.Equal(t, "testpassword", options.Password)
	assert.Equal(t, 2, options.DB)
}

// TestBus_New_NilConfig ...
// Summary: TestBus_New_NilConfig
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	var bus *Bus[string]
	assert.NotPanics(t, func() {
		bus, _ = New[string](nil)
	})
	assert.NotNil(t, bus)
	assert.NotNil(t, bus.client)
	options := bus.client.Options()
	assert.Equal(t, "127.0.0.1:6379", options.Addr)
	assert.Equal(t, "", options.Password)
	assert.Equal(t, 0, options.DB)
}

// TestBus_New_PartialConfig ...
// Summary: TestBus_New_PartialConfig
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	redisBus := bus_pb.RedisBus_builder{
		Address: proto.String("127.0.0.1:6381"),
	}.Build()

	bus, _ := New[string](redisBus)
	assert.NotNil(t, bus)
	assert.NotNil(t, bus.client)
	options := bus.client.Options()
	assert.Equal(t, "127.0.0.1:6381", options.Addr)
	assert.Equal(t, "", options.Password)
	assert.Equal(t, 0, options.DB)
}

// TestBus_ConcurrentSubscribeAndUnsubscribe ...
// Summary: TestBus_ConcurrentSubscribeAndUnsubscribe
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)
	topic := "concurrent-topic"

	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			unsub := bus.Subscribe(context.Background(), topic, func(_ string) {
				// No-op handler
			})
			time.Sleep(10 * time.Millisecond) // Give a chance for other goroutines to run
			unsub()
		}()
	}

	wg.Wait()
}

// TestBus_Subscribe_NilHandler ...
// Summary: TestBus_Subscribe_NilHandler
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)
	topic := "test-nil-handler"

	assert.NotPanics(t, func() {
		bus.Subscribe(context.Background(), topic, nil)
	})
}

// TestBus_Subscribe_CloseSubscription ...
// Summary: TestBus_Subscribe_CloseSubscription
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)
	topic := "test-close-subscription"

	// Use goleak to verify that the goroutine exits.
	// We need to ignore the goroutines that are part of the test framework.
	// Ensure client is closed BEFORE goleak check runs (LIFO defer order: close first then verify).
	defer goleak.VerifyNone(t,
		goleak.IgnoreTopFunction("testing.runTests"),
		goleak.IgnoreTopFunction("testing.(*T).Run"),
		goleak.IgnoreTopFunction("sync.runtime_notifyListWait"),
		goleak.IgnoreTopFunction("internal/poll.runtime_pollWait"),
		goleak.IgnoreTopFunction("github.com/go-redis/redis/v9/internal/pool.(*ConnPool).reaper"),
		goleak.IgnoreTopFunction("github.com/redis/go-redis/v9/maintnotifications.(*CircuitBreakerManager).cleanupLoop"),
	)
	defer client.Close()

	unsub := bus.Subscribe(context.Background(), topic, func(_ string) {
		// This handler should not be called.
		t.Error("handler called unexpectedly")
	})

	require.Eventually(t, func() bool {
		subs, err := client.PubSubNumSub(context.Background(), topic).Result()
		if err != nil {
			return false
		}
		if val, ok := subs[topic]; ok {
			return val == 1
		}
		return false
	}, 5*time.Second, 10*time.Millisecond, "subscriber did not appear")

	unsub()

	// Allow some time for the goroutine to exit.
	time.Sleep(100 * time.Millisecond)
}

// TestBus_SubscribeOnce_UnsubscribeBeforeMessage ...
// Summary: TestBus_SubscribeOnce_UnsubscribeBeforeMessage
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)
	topic := "once-unsubscribe-before-message"

	handlerCalled := make(chan bool, 1)

	unsub := bus.SubscribeOnce(context.Background(), topic, func(_ string) {
		handlerCalled <- true
	})

	require.Eventually(t, func() bool {
		subs := client.PubSubNumSub(context.Background(), topic).Val()
		return len(subs) > 0 && subs[topic] == 1
	}, 5*time.Second, 10*time.Millisecond, "subscriber did not appear")

	unsub()

	require.Eventually(t, func() bool {
		subs := client.PubSubNumSub(context.Background(), topic).Val()
		return len(subs) == 0 || subs[topic] == 0
	}, 5*time.Second, 10*time.Millisecond, "subscriber did not disappear")

	err := bus.Publish(context.Background(), topic, "hello")
	assert.NoError(t, err)

	select {
	case <-handlerCalled:
		t.Fatal("handler should not have been called after unsubscribe")
	case <-time.After(200 * time.Millisecond):
		// Test passed
	}
}

// TestBus_Subscribe_HandlerPanic ...
// Summary: TestBus_Subscribe_HandlerPanic
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
		if err != nil {
			return false
		}
		if val, ok := subs[topic]; ok {
			return val == 1
		}
		return false
	}, 5*time.Second, 10*time.Millisecond, "subscriber did not appear")

	// Even if the handler panics, the subscription should remain active
	_ = bus.Publish(context.Background(), topic, "first message")
	_ = bus.Publish(context.Background(), topic, "second message")

	// Wait for both messages to be processed
	<-handlerCalled
	<-handlerCalled

	assert.Len(t, handlerCalled, 0, "handler should have been called twice")
}

// TestBus_Subscribe_ContextCancellation ...
// Summary: TestBus_Subscribe_ContextCancellation
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)
	topic := "test-context-cancellation"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)

	unsub := bus.Subscribe(ctx, topic, func(_ string) {
		// This handler should not be called after the context is canceled.
		t.Error("handler called after context cancellation")
	})
	defer unsub()

	require.Eventually(t, func() bool {
		subs, err := client.PubSubNumSub(context.Background(), topic).Result()
		if err != nil {
			return false
		}
		if val, ok := subs[topic]; ok {
			return val == 1
		}
		return false
	}, 5*time.Second, 10*time.Millisecond, "subscriber did not appear")

	cancel()

	// Allow some time for the cancellation to propagate
	time.Sleep(100 * time.Millisecond)

	err := bus.Publish(context.Background(), topic, "hello")
	assert.NoError(t, err)

	// We expect the handler not to be called, so we don't wait for a WaitGroup.
	// Instead, we just wait a bit to see if the handler is called.
	time.Sleep(100 * time.Millisecond)
}

// TestBus_Subscribe_AlreadyCancelledContext ...
// Summary: TestBus_Subscribe_AlreadyCancelledContext
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)
	topic := "already-cancelled-context"

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	unsub := bus.Subscribe(ctx, topic, func(_ string) {
		t.Error("handler should not be called")
	})
	defer unsub()

	// The subscription should not be created in the first place, but we check regardless
	require.Eventually(t, func() bool {
		subs, err := client.PubSubNumSub(context.Background(), topic).Result()
		if err != nil {
			return false
		}
		if val, ok := subs[topic]; ok {
			return val == 0
		}
		return true // No subscribers is the desired state
	}, 5*time.Second, 10*time.Millisecond, "subscriber should not be created with a cancelled context")

	err := bus.Publish(context.Background(), topic, "hello")
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
}

// TestBus_PublishAndSubscribe ...
// Summary: TestBus_PublishAndSubscribe
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)
	topic := "test-publish-subscribe"

	var wg sync.WaitGroup
	wg.Add(1)

	handlerCalled := make(chan string, 1)

	unsub := bus.Subscribe(context.Background(), topic, func(msg string) {
		handlerCalled <- msg
		wg.Done()
	})
	defer unsub()

	require.Eventually(t, func() bool {
		subs := client.PubSubNumSub(context.Background(), topic).Val()
		return len(subs) > 0 && subs[topic] == 1
	}, 5*time.Second, 10*time.Millisecond, "subscriber did not appear")

	err := bus.Publish(context.Background(), topic, "hello")
	assert.NoError(t, err)

	wg.Wait()

	select {
	case msg := <-handlerCalled:
		assert.Equal(t, "hello", msg)
	default:
		t.Fatal("handler should have been called")
	}

	unsub()

	require.Eventually(t, func() bool {
		subs := client.PubSubNumSub(context.Background(), topic).Val()
		return len(subs) == 0 || subs[topic] == 0
	}, 5*time.Second, 10*time.Millisecond, "subscriber did not disappear")

	err = bus.Publish(context.Background(), topic, "world")
	assert.NoError(t, err)

	select {
	case <-handlerCalled:
		t.Fatal("handler should not have been called after unsubscribe")
	case <-time.After(200 * time.Millisecond):
		// Test passed
	}
}

// TestBus_UnsubscribeFromHandler ...
// Summary: TestBus_UnsubscribeFromHandler
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)
	topic := "unsubscribe-from-handler"

	var wg sync.WaitGroup
	wg.Add(1)

	var unsub func()
	unsub = bus.Subscribe(context.Background(), topic, func(_ string) {
		unsub()
		wg.Done()
	})

	require.Eventually(t, func() bool {
		subs, err := client.PubSubNumSub(context.Background(), topic).Result()
		if err != nil {
			return false
		}
		if val, ok := subs[topic]; ok {
			return val == 1
		}
		return false
	}, 5*time.Second, 10*time.Millisecond, "subscriber did not appear")

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
	}, 5*time.Second, 10*time.Millisecond, "subscriber did not disappear after unsubscribing from handler")
}

// TestBus_SubscribeOnce_CancelledContext ...
// Summary: TestBus_SubscribeOnce_CancelledContext
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)
	topic := "once-cancelled-context"

	ctx, cancel := context.WithCancel(context.Background())

	unsub := bus.SubscribeOnce(ctx, topic, func(_ string) {
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
	}, 5*time.Second, 10*time.Millisecond, "subscriber did not appear")

	cancel()

	require.Eventually(t, func() bool {
		subs, err := client.PubSubNumSub(context.Background(), topic).Result()
		require.NoError(t, err)
		if val, ok := subs[topic]; ok {
			return val == 0
		}
		return true
	}, 5*time.Second, 10*time.Millisecond, "subscriber did not disappear after context cancellation")

	err := bus.Publish(context.Background(), topic, "hello")
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
}

// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// TestBus_Subscribe_ConcurrentSubscribers ...
// Summary: TestBus_Subscribe_ConcurrentSubscribers
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// t.Skip("Skipping flaky concurrent subscribers test")
	client := setupRedisIntegrationTest(t)
	// Explicitly close client to avoid leaks if this test becomes sensitive or other tests run
	defer client.Close()

	bus := NewWithClient[string](client)
	topic := "concurrent-subscribers"
	numSubscribers := 2
	var wg sync.WaitGroup
	wg.Add(numSubscribers)
	var setupWg sync.WaitGroup
	setupWg.Add(numSubscribers)

	unsubs := make(chan func(), numSubscribers)
	for i := 0; i < numSubscribers; i++ {
		go func() {
			defer setupWg.Done()
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
	setupWg.Wait()

	close(unsubs)
	for unsub := range unsubs {
		unsub()
	}
}

// TestBus_SubscribeOnce_UnsubscribeFromHandler ...
// Summary: TestBus_SubscribeOnce_UnsubscribeFromHandler
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
	}, 5*time.Second, 10*time.Millisecond, "subscriber did not appear")

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
	}, 5*time.Second, 10*time.Millisecond, "subscriber did not disappear after unsubscribing from handler")
}

// TestBus_Subscribe_CloseClient ...
// Summary: TestBus_Subscribe_CloseClient
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
	}, 5*time.Second, 10*time.Millisecond, "subscriber did not appear")

	// Close the client, which should terminate the subscription loop
	err := client.Close()
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

}

// TestBus_Subscribe_CloseClient_Race ...
// Summary: TestBus_Subscribe_CloseClient_Race
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestBus_Unsubscribe_Race ...
// Summary: TestBus_Unsubscribe_Race
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestBus_Subscribe_And_Unsubscribe_Race ...
// Summary: TestBus_Subscribe_And_Unsubscribe_Race
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
