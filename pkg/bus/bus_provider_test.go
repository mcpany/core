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
	"sync"
	"testing"
	"time"

	"github.com/mcpany/core/proto/bus"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)


func TestBusProvider_GetBus_Redis(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		t.Skip("Redis is not available")
	}

	messageBus := &bus.MessageBus{}
	redisBus := &bus.RedisBus{}
	redisBus.SetAddress("localhost:6379")
	messageBus.SetRedis(redisBus)

	provider, err := NewBusProvider(messageBus)
	assert.NoError(t, err)

	bus1 := GetBus[string](provider, "test_topic")
	bus2 := GetBus[string](provider, "test_topic")

	assert.NotNil(t, bus1)
	assert.Same(t, bus1, bus2, "Expected the same bus instance for the same topic")
}

func TestBusProvider_GetBus_Nats(t *testing.T) {
	s, err := server.NewServer(&server.Options{Port: -1})
	assert.NoError(t, err)
	go s.Start()
	defer s.Shutdown()
	if !s.ReadyForConnections(4 * time.Second) {
		t.Fatalf("NATS server failed to start")
	}

	messageBus := &bus.MessageBus{}
	natsBus := &bus.NatsBus{}
	natsBus.SetServerUrl(s.ClientURL())
	messageBus.SetNats(natsBus)

	provider, err := NewBusProvider(messageBus)
	assert.NoError(t, err)

	bus1 := GetBus[string](provider, "test_topic")
	bus2 := GetBus[string](provider, "test_topic")

	assert.NotNil(t, bus1)
	assert.Same(t, bus1, bus2, "Expected the same bus instance for the same topic")
}

func TestBusProvider_GetBus_Concurrent(t *testing.T) {
	s, err := server.NewServer(&server.Options{Port: -1})
	assert.NoError(t, err)
	go s.Start()
	defer s.Shutdown()
	if !s.ReadyForConnections(4 * time.Second) {
		t.Fatalf("NATS server failed to start")
	}

	messageBus := &bus.MessageBus{}
	natsBus := &bus.NatsBus{}
	natsBus.SetServerUrl(s.ClientURL())
	messageBus.SetNats(natsBus)
	provider, err := NewBusProvider(messageBus)
	assert.NoError(t, err)

	var wg sync.WaitGroup
	numGoroutines := 100
	wg.Add(numGoroutines)

	buses := make(chan Bus[string], numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			bus := GetBus[string](provider, "concurrent_topic")
			buses <- bus
		}()
	}

	wg.Wait()
	close(buses)

	firstBus := <-buses
	for bus := range buses {
		assert.Same(t, firstBus, bus, "Expected all goroutines to get the same bus instance")
	}
}

func TestBusProvider_DefaultBus(t *testing.T) {
	provider, err := NewBusProvider(nil)
	assert.NoError(t, err)
	assert.NotNil(t, provider.config)
	assert.NotNil(t, provider.config.GetNats())
}
