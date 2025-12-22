package bus

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/mcpany/core/proto/bus"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

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

	t.Run("Redis", func(t *testing.T) {
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
		kafkaBus.SetBrokers([]string{"localhost:9092"})
		messageBus.SetKafka(kafkaBus)

		provider, err := NewProvider(messageBus)
		assert.NoError(t, err)

		bus1, _ := GetBus[string](provider, "strings")
		assert.NotNil(t, bus1)
	})
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
		Addr: "localhost:6379",
	})
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		t.Skip("Redis is not available")
	}

	messageBus := bus.MessageBus_builder{}.Build()
	redisBus := bus.RedisBus_builder{}.Build()
	redisBus.SetAddress("localhost:6379")
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

	_ = bus.Publish(context.Background(), "test-message", "hello")
	_ = bus.Publish(context.Background(), "test-message", "world")

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

	provider, err := NewProvider(messageBus)
	assert.NoError(t, err)

	bus, _ := GetBus[string](provider, "test-topic")

	var receivedMessages []string
	unsubscribe := bus.Subscribe(context.Background(), "test-message", func(msg string) {
		receivedMessages = append(receivedMessages, msg)
	})

	_ = bus.Publish(context.Background(), "test-message", "hello")
	time.Sleep(100 * time.Millisecond) // Allow time for the message to be processed

	unsubscribe()

	_ = bus.Publish(context.Background(), "test-message", "world")
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

	provider, err := NewProvider(messageBus)
	assert.NoError(t, err)

	bus, _ := GetBus[string](provider, "test-topic")

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
		_ = bus.Publish(context.Background(), "test-message", "hello")
	}

	wg.Wait()

	for i := 0; i < numSubscribers; i++ {
		assert.Len(t, receivedMessages[i], numMessages)
	}
}
