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

func TestBusProvider_GetBus_InMemory(t *testing.T) {
	messageBus := &bus.MessageBus{}
	messageBus.SetInMemory(&bus.InMemoryBus{})
	provider, err := NewProvider(messageBus)
	assert.NoError(t, err)

	bus1, _ := GetBus[string](provider, "test_topic")
	bus2, _ := GetBus[string](provider, "test_topic")
	bus3, _ := GetBus[int](provider, "another_topic")

	assert.NotNil(t, bus1)
	assert.Same(t, bus1, bus2, "Expected the same bus instance for the same topic")
	assert.NotSame(t, bus1, bus3, "Expected different bus instances for different topics")
}

func TestBusProvider_GetBus_Redis(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		// t.Skip("Redis is not available")
	}

	messageBus := &bus.MessageBus{}
	redisBus := &bus.RedisBus{}
	redisBus.SetAddress("127.0.0.1:6379")
	messageBus.SetRedis(redisBus)

	provider, err := NewProvider(messageBus)
	assert.NoError(t, err)

	bus1, _ := GetBus[string](provider, "test_topic")
	bus2, _ := GetBus[string](provider, "test_topic")

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

	provider, err := NewProvider(messageBus)
	assert.NoError(t, err)

	bus1, _ := GetBus[string](provider, "test_topic")
	bus2, _ := GetBus[string](provider, "test_topic")

	assert.NotNil(t, bus1)
	assert.Same(t, bus1, bus2, "Expected the same bus instance for the same topic")
}

func TestBusProvider_GetBus_Concurrent(t *testing.T) {
	messageBus := &bus.MessageBus{}
	messageBus.SetInMemory(&bus.InMemoryBus{})
	provider, err := NewProvider(messageBus)
	assert.NoError(t, err)

	var wg sync.WaitGroup
	numGoroutines := 100
	wg.Add(numGoroutines)

	buses := make(chan Bus[string], numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			bus, _ := GetBus[string](provider, "concurrent_topic")
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
	provider, err := NewProvider(nil)
	assert.NoError(t, err)
	assert.NotNil(t, provider.config)
	assert.NotNil(t, provider.config.GetInMemory())
}
