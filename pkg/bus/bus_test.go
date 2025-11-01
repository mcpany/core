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
	"testing"

	"github.com/mcpany/core/proto/bus"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestBusProvider(t *testing.T) {
	t.Run("InMemory", func(t *testing.T) {
		config := &bus.MessageBus{}
		config.SetInMemory(&bus.InMemoryBus{})
		provider, err := NewBusProvider(config)
		assert.NoError(t, err)

		bus1 := GetBus[string](provider, "strings")
		bus2 := GetBus[int](provider, "ints")
		bus3 := GetBus[string](provider, "strings")

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

		config := &bus.MessageBus{}
		redisBus := &bus.RedisBus{}
		redisBus.SetAddress("localhost:6379")
		config.SetRedis(redisBus)

		provider, err := NewBusProvider(config)
		assert.NoError(t, err)

		bus1 := GetBus[string](provider, "strings")
		bus2 := GetBus[int](provider, "ints")
		bus3 := GetBus[string](provider, "strings")

		assert.NotNil(t, bus1)
		assert.NotNil(t, bus2)
		assert.Same(t, bus1, bus3)
	})
}

func TestIntegration(t *testing.T) {
	config := &bus.MessageBus{}
	config.SetInMemory(&bus.InMemoryBus{})
	provider, err := NewBusProvider(config)
	assert.NoError(t, err)

	// Simulate a tool execution request/response
	reqBus := GetBus[*ToolExecutionRequest](provider, "tool_requests")
	resBus := GetBus[*ToolExecutionResult](provider, "tool_results")

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
		resBus.Publish(context.Background(), req.CorrelationID(), res)
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
	reqBus.Publish(context.Background(), "request", req)

	wg.Wait()
}

func TestBusProvider_Concurrent(t *testing.T) {
	config := &bus.MessageBus{}
	config.SetInMemory(&bus.InMemoryBus{})
	provider, err := NewBusProvider(config)
	assert.NoError(t, err)

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
