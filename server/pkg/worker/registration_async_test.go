package worker

import (
	"context"
	"testing"
	"time"

	bus_pb "github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceRegistrationWorker_Async(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	bp, err := bus.NewProvider(messageBus)
	require.NoError(t, err)
	requestBus, err := bus.GetBus[*bus.ServiceRegistrationRequest](bp, bus.ServiceRegistrationRequestTopic)
	require.NoError(t, err)
	resultBus, err := bus.GetBus[*bus.ServiceRegistrationResult](bp, bus.ServiceRegistrationResultTopic)
	require.NoError(t, err)

	slowServiceKey := "slow-service"
	fastServiceKey := "fast-service"

	registry := &mockServiceRegistry{
		registerFunc: func(_ context.Context, cfg *configv1.UpstreamServiceConfig) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
			t.Logf("Mock register received: '%s'", cfg.GetName())
			if cfg.GetName() == slowServiceKey {
				t.Log("Sleeping for slow service...")
				time.Sleep(200 * time.Millisecond)
				return "slow-key", nil, nil, nil
			}
			return "fast-key", nil, nil, nil
		},
	}

	worker := NewServiceRegistrationWorker(bp, registry)
	worker.Start(ctx)

	// Subscribe for results
	results := make(map[string]*bus.ServiceRegistrationResult)
	resultChan := make(chan *bus.ServiceRegistrationResult, 2)

	// Subscribe to specific correlation IDs
	resultBus.Subscribe(ctx, "slow", func(result *bus.ServiceRegistrationResult) {
		resultChan <- result
	})
	resultBus.Subscribe(ctx, "fast", func(result *bus.ServiceRegistrationResult) {
		resultChan <- result
	})

	start := time.Now()

	// Publish slow request first
	cfgSlow := &configv1.UpstreamServiceConfig{}
	cfgSlow.SetName(slowServiceKey)
	reqSlow := &bus.ServiceRegistrationRequest{Config: cfgSlow}
	reqSlow.SetCorrelationID("slow")
	err = requestBus.Publish(ctx, "request", reqSlow)
	require.NoError(t, err)

	// Publish fast request immediately after
	cfgFast := &configv1.UpstreamServiceConfig{}
	cfgFast.SetName(fastServiceKey)
	reqFast := &bus.ServiceRegistrationRequest{Config: cfgFast}
	reqFast.SetCorrelationID("fast")
	err = requestBus.Publish(ctx, "request", reqFast)
	require.NoError(t, err)

	timeout := time.After(1 * time.Second)

	// Wait for first result
	select {
	case res := <-resultChan:
		t.Logf("Received first result: %s at %v", res.CorrelationID(), time.Since(start))
		assert.Equal(t, "fast", res.CorrelationID(), "Expected fast result first")
		results[res.CorrelationID()] = res
	case <-timeout:
		t.Fatal("Timeout waiting for first result")
	}

	// Wait for second result
	select {
	case res := <-resultChan:
		t.Logf("Received second result: %s at %v", res.CorrelationID(), time.Since(start))
		results[res.CorrelationID()] = res
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for second result")
	}

	assert.Equal(t, "fast-key", results["fast"].ServiceKey)
	assert.Equal(t, "slow-key", results["slow"].ServiceKey)
}
