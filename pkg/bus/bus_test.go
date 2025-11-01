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
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestDefaultBus(t *testing.T) {
	t.Run("Publish and Subscribe", func(t *testing.T) {
		bus := New[string]()
		var wg sync.WaitGroup
		wg.Add(1)

		bus.Subscribe("test", func(msg string) {
			assert.Equal(t, "hello", msg)
			wg.Done()
		})

		bus.Publish("test", "hello")
		wg.Wait()
	})

	t.Run("SubscribeOnce", func(t *testing.T) {
		bus := New[string]()
		var wg sync.WaitGroup
		wg.Add(1)

		bus.SubscribeOnce("test", func(msg string) {
			assert.Equal(t, "hello", msg)
			wg.Done()
		})

		bus.Publish("test", "hello")
		wg.Wait()

		// This second publish should not be received
		bus.Publish("test", "world")
	})

	t.Run("Unsubscribe", func(t *testing.T) {
		bus := New[string]()
		received := false

		unsub := bus.Subscribe("test", func(msg string) {
			received = true
		})

		unsub()
		bus.Publish("test", "hello")
		time.Sleep(10 * time.Millisecond) // Give it a moment to process
		assert.False(t, received)
	})
}

func TestBusProvider(t *testing.T) {
	t.Run("InMemory", func(t *testing.T) {
		busConfig := &configv1.MessageBus{}
		busConfig.SetInMemory(&configv1.InMemoryBus{})
		provider := NewBusProvider(busConfig)

		bus1 := GetBus[string](provider, "strings")
		bus2 := GetBus[int](provider, "ints")
		bus3 := GetBus[string](provider, "strings")

		assert.NotNil(t, bus1)
		assert.NotNil(t, bus2)
		assert.Same(t, bus1, bus3)
		assert.IsType(t, &DefaultBus[string]{}, bus1)
	})

	t.Run("Redis", func(t *testing.T) {
		busConfig := &configv1.MessageBus{}
		redisBusConfig := &configv1.RedisBus{}
		redisBusConfig.SetAddress("localhost:6379")
		busConfig.SetRedis(redisBusConfig)
		provider := NewBusProvider(busConfig)

		bus1 := GetBus[string](provider, "strings")
		assert.NotNil(t, bus1)
		assert.IsType(t, &RedisBus[string]{}, bus1)
	})
}

func TestIntegration(t *testing.T) {
	busConfig := &configv1.MessageBus{}
	busConfig.SetInMemory(&configv1.InMemoryBus{})
	provider := NewBusProvider(busConfig)

	// Simulate a tool execution request/response
	reqBus := GetBus[*ToolExecutionRequest](provider, "tool_requests")
	resBus := GetBus[*ToolExecutionResult](provider, "tool_results")

	var wg sync.WaitGroup
	wg.Add(1)

	// Worker subscribing to requests
	reqBus.Subscribe("request", func(req *ToolExecutionRequest) {
		assert.Equal(t, "test-tool", req.ToolName)
		resultData, err := json.Marshal(map[string]any{"status": "ok"})
		assert.NoError(t, err)
		res := &ToolExecutionResult{
			BaseMessage: BaseMessage{CID: req.CorrelationID()},
			Result:      resultData,
		}
		resBus.Publish(req.CorrelationID(), res)
	})

	// Client subscribing to results
	resBus.SubscribeOnce("test-correlation-id", func(res *ToolExecutionResult) {
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
	reqBus.Publish("request", req)

	wg.Wait()
}

func TestBusProvider_Concurrent(t *testing.T) {
	busConfig := &configv1.MessageBus{}
	busConfig.SetInMemory(&configv1.InMemoryBus{})
	provider := NewBusProvider(busConfig)
	numGoroutines := 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	stringBus := GetBus[string](provider, "string_topic")

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			// Concurrently get the same bus
			bus := GetBus[string](provider, "string_topic")
			assert.Same(t, stringBus, bus, "Expected the same bus instance to be returned")
		}()
	}

	wg.Wait()
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
		unsub := bus.Subscribe(topic, func(msg int) {
			atomic.AddInt32(&receivedCount, 1)
			wg.Done()
		})
		defer unsub()
	}

	for i := 0; i < numPublishers; i++ {
		go bus.Publish(topic, i)
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
