// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package bus

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/mcpany/core/proto/bus"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func waitForSubscribers(t *testing.T, client *redis.Client, topic string, expected int) {
	require.Eventually(t, func() bool {
		subs := client.PubSubNumSub(context.Background(), topic).Val()
		t.Logf("Waiting for subscribers on %s: have %d, want >= %d", topic, subs[topic], expected)
		return subs[topic] >= int64(expected)
	}, 5*time.Second, 500*time.Millisecond, "timed out waiting for subscribers on topic %s", topic)
}

func TestBusProvider(t *testing.T) {
	t.Run("InMemory", func(t *testing.T) {
		messageBus := bus.MessageBus_builder{}.Build()
		messageBus.SetInMemory(bus.InMemoryBus_builder{}.Build())
		provider, err := NewProvider(messageBus)
		assert.NoError(t, err)

		bus1, _ := GetBus[string](provider, "strings")
		bus2, _ := GetBus[int](provider, "ints")
		bus3, _ := GetBus[string](provider, "strings")

		assert.NotNil(t, bus1)
		assert.NotNil(t, bus2)
		assert.Same(t, bus1, bus3)
	})


	t.Run("Nats", func(t *testing.T) {
		// NatsBus builder with empty URL triggers the embedded NATS server,
		// so this test does not require an external NATS server.
		messageBus := bus.MessageBus_builder{}.Build()
		natsBus := bus.NatsBus_builder{}.Build()
		messageBus.SetNats(natsBus)

		provider, err := NewProvider(messageBus)
		assert.NoError(t, err)

		bus1, _ := GetBus[string](provider, "strings")
		bus2, _ := GetBus[int](provider, "ints")

		assert.NotNil(t, bus1)
		assert.NotNil(t, bus2)

		// Simple publish/subscribe verification
		var wg sync.WaitGroup
		wg.Add(1)

		bus1.SubscribeOnce(context.Background(), "test-topic", func(msg string) {
			assert.Equal(t, "hello", msg)
			wg.Done()
		})

		// Give time for subscription to be active
		time.Sleep(100 * time.Millisecond)

		err = bus1.Publish(context.Background(), "test-topic", "hello")
		assert.NoError(t, err)

		wg.Wait()
	})

	t.Run("Kafka", func(t *testing.T) {
		// This tests that the provider can instantiate a Kafka bus with valid config.
		// It does not attempt to connect to the broker (lazy connection),
		// so it does not require an external Kafka server.
		messageBus := bus.MessageBus_builder{}.Build()
		kafkaBus := bus.KafkaBus_builder{}.Build()
		kafkaBus.SetBrokers([]string{"127.0.0.1:9092"})
		messageBus.SetKafka(kafkaBus)

		provider, err := NewProvider(messageBus)
		assert.NoError(t, err)

		bus1, _ := GetBus[string](provider, "strings")
		assert.NotNil(t, bus1)
	})
}

func TestBusProvider_Errors(t *testing.T) {
	t.Run("NewProvider_NilConfig", func(t *testing.T) {
		// Should default to InMemory
		provider, err := NewProvider(nil)
		assert.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, bus.MessageBus_InMemory_case, provider.config.WhichBusType())
	})

	t.Run("NewProvider_EmptyConfig", func(t *testing.T) {
		// Should default to InMemory
		messageBus := bus.MessageBus_builder{}.Build()
		provider, err := NewProvider(messageBus)
		assert.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, bus.MessageBus_InMemory_case, provider.config.WhichBusType())
	})

	t.Run("GetBus_HookError", func(t *testing.T) {
		messageBus := bus.MessageBus_builder{}.Build()
		messageBus.SetInMemory(bus.InMemoryBus_builder{}.Build())
		provider, err := NewProvider(messageBus)
		assert.NoError(t, err)

		// Set hook to simulate error
		originalHook := GetBusHook
		defer func() { GetBusHook = originalHook }()
		GetBusHook = func(p *Provider, topic string) (any, error) {
			return nil, errors.New("simulated error")
		}

		b, err := GetBus[string](provider, "test")
		assert.Error(t, err)
		assert.Nil(t, b)
		assert.Equal(t, "simulated error", err.Error())
	})

	t.Run("GetBus_HookSuccess", func(t *testing.T) {
		messageBus := bus.MessageBus_builder{}.Build()
		messageBus.SetInMemory(bus.InMemoryBus_builder{}.Build())
		provider, err := NewProvider(messageBus)
		assert.NoError(t, err)

		// Set hook to simulate success with a mocked bus (or just nil as Any)
		// But GetBus expects Bus[T]
		originalHook := GetBusHook
		defer func() { GetBusHook = originalHook }()

		// Use in-memory bus as return value
		mockBus := &MockBus[string]{}
		GetBusHook = func(p *Provider, topic string) (any, error) {
			return mockBus, nil
		}

		b, err := GetBus[string](provider, "test")
		assert.NoError(t, err)
		assert.Equal(t, mockBus, b)
	})
}

type MockBus[T any] struct {
}

func (m *MockBus[T]) Publish(ctx context.Context, topic string, msg T) error {
	return nil
}

func (m *MockBus[T]) Subscribe(ctx context.Context, topic string, handler func(T)) (unsubscribe func()) {
	return func() {}
}

func (m *MockBus[T]) SubscribeOnce(ctx context.Context, topic string, handler func(T)) (unsubscribe func()) {
	return func() {}
}


func TestIntegration(t *testing.T) {
	messageBus := bus.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus.InMemoryBus_builder{}.Build())
	provider, err := NewProvider(messageBus)
	assert.NoError(t, err)

	// Simulate a tool execution request/response
	reqBus, _ := GetBus[*ToolExecutionRequest](provider, "tool_requests")
	resBus, _ := GetBus[*ToolExecutionResult](provider, "tool_results")

	var wg sync.WaitGroup
	wg.Add(1)

	// Worker subscribing to requests
	reqBus.Subscribe(context.Background(), "request", func(req *ToolExecutionRequest) {
		assert.Equal(t, "test-tool", req.ToolName)
		resultData, err := json.Marshal(map[string]any{"status": "ok"})
		assert.NoError(t, err)
		res := &ToolExecutionResult{
			BaseMessage: BaseMessage{CID: req.CorrelationID()},
			Result:      resultData,
		}
		_ = resBus.Publish(context.Background(), req.CorrelationID(), res)
	})

	// Client subscribing to results
	resBus.SubscribeOnce(context.Background(), "test-correlation-id", func(res *ToolExecutionResult) {
		assert.Equal(t, "test-correlation-id", res.CorrelationID())
		expectedResultData, err := json.Marshal(map[string]any{"status": "ok"})
		assert.NoError(t, err)
		assert.JSONEq(t, string(expectedResultData), string(res.Result))
		wg.Done()
	})

	// Client publishing a request
	inputData, err := json.Marshal(map[string]any{"input": "data"})
	assert.NoError(t, err)
	req := &ToolExecutionRequest{
		BaseMessage: BaseMessage{CID: "test-correlation-id"},
		Context:     context.Background(),
		ToolName:    "test-tool",
		ToolInputs:  inputData,
	}
	_ = reqBus.Publish(context.Background(), "request", req)

	wg.Wait()
}

func TestBusProvider_Concurrent(t *testing.T) {
	messageBus := bus.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus.InMemoryBus_builder{}.Build())
	provider, err := NewProvider(messageBus)
	assert.NoError(t, err)

	numGoroutines := 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	stringBus, _ := GetBus[string](provider, "string_topic")

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			// Concurrently get the same bus
			bus, _ := GetBus[string](provider, "string_topic")
			assert.Same(t, stringBus, bus, "Expected the same bus instance to be returned")
		}()
	}

	wg.Wait()
}

func TestRedisBus_SubscribeOnce(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		t.Skip("Redis is not available")
	}

	messageBus := bus.MessageBus_builder{}.Build()
	redisBus := bus.RedisBus_builder{}.Build()
	redisBus.SetAddress("127.0.0.1:6379")
	messageBus.SetRedis(redisBus)

	provider, err := NewProvider(messageBus)
	assert.NoError(t, err)

	bus, _ := GetBus[string](provider, "test-topic")

	var wg sync.WaitGroup
	wg.Add(1)

	var receivedMessages []string
	bus.SubscribeOnce(context.Background(), "test-message", func(msg string) {
		receivedMessages = append(receivedMessages, msg)
		wg.Done()
	})

	// Wait for subscription to be active
	waitForSubscribers(t, client, "test-message", 1)

	_ = bus.Publish(context.Background(), "test-message", "hello")
	_ = bus.Publish(context.Background(), "test-message", "world")

	wg.Wait()

	assert.Len(t, receivedMessages, 1)
	assert.Equal(t, "hello", receivedMessages[0])
}

func TestRedisBus_Unsubscribe(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		t.Skip("Redis is not available")
	}

	messageBus := bus.MessageBus_builder{}.Build()
	redisBus := bus.RedisBus_builder{}.Build()
	redisBus.SetAddress("127.0.0.1:6379")
	messageBus.SetRedis(redisBus)

	provider, err := NewProvider(messageBus)
	assert.NoError(t, err)

	bus, _ := GetBus[string](provider, "test-topic")

	var mu sync.Mutex
	var receivedMessages []string
	var wg sync.WaitGroup
	wg.Add(1)

	unsubscribe := bus.Subscribe(context.Background(), "test-message", func(msg string) {
		mu.Lock()
		defer mu.Unlock()
		receivedMessages = append(receivedMessages, msg)
		if msg == "hello" {
			wg.Done()
		}
	})

	// Wait for subscription to be active
	waitForSubscribers(t, client, "test-message", 1)

	_ = bus.Publish(context.Background(), "test-message", "hello")

	// Wait for the first message to be processed
	wg.Wait()

	unsubscribe()

	_ = bus.Publish(context.Background(), "test-message", "world")
	time.Sleep(100 * time.Millisecond) // Allow time for the potential message to be processed (should not be)

	mu.Lock()
	defer mu.Unlock()
	assert.Len(t, receivedMessages, 1)
	assert.Equal(t, "hello", receivedMessages[0])
}

func TestRedisBus_Concurrent(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		t.Skip("Redis is not available")
	}

	messageBus := bus.MessageBus_builder{}.Build()
	redisBus := bus.RedisBus_builder{}.Build()
	redisBus.SetAddress("127.0.0.1:6379")
	messageBus.SetRedis(redisBus)

	provider, err := NewProvider(messageBus)
	assert.NoError(t, err)

	bus, _ := GetBus[string](provider, "test-topic")

	numSubscribers := 2
	numMessages := 10
	var wg sync.WaitGroup
	wg.Add(numSubscribers * numMessages)

	var receivedMessages [][]string
	for i := 0; i < numSubscribers; i++ {
		receivedMessages = append(receivedMessages, []string{})
		go func(i int) {
			// Create a new provider for each subscriber to simulate distributed nodes
			// This ensures we have distinct Bus instances and distinct Redis subscriptions
			localProvider, _ := NewProvider(messageBus)
			localBus, _ := GetBus[string](localProvider, "test-message")

			localBus.Subscribe(context.Background(), "test-message", func(msg string) {
				receivedMessages[i] = append(receivedMessages[i], msg)
				wg.Done()
			})
		}(i)
	}

	// Wait for subscriptions to be active
	waitForSubscribers(t, client, "test-message", numSubscribers)

	for i := 0; i < numMessages; i++ {
		_ = bus.Publish(context.Background(), "test-message", "hello")
	}

	wg.Wait()

	for i := 0; i < numSubscribers; i++ {
		assert.Len(t, receivedMessages[i], numMessages)
	}
}
