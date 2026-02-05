// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	bus_pb "github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceRegistrationWorker_Panic_Registration(t *testing.T) {
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

	registry := &mockServiceRegistry{
		registerFunc: func(_ context.Context, _ *configv1.UpstreamServiceConfig) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
			panic("simulated panic")
		},
	}

	worker := NewServiceRegistrationWorker(bp, registry)
	worker.Start(ctx)

	resultChan := make(chan *bus.ServiceRegistrationResult, 1)
	unsubscribe := resultBus.SubscribeOnce(ctx, "test-panic", func(result *bus.ServiceRegistrationResult) {
		resultChan <- result
	})
	defer unsubscribe()

	req := &bus.ServiceRegistrationRequest{Config: &configv1.UpstreamServiceConfig{}}
	req.SetCorrelationID("test-panic")
	err = requestBus.Publish(ctx, "request", req)
	require.NoError(t, err)

	select {
	case result := <-resultChan:
		require.Error(t, result.Error)
		assert.Contains(t, result.Error.Error(), "panic during registration: simulated panic")
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for registration result")
	}
}

func TestServiceRegistrationWorker_Panic_List(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	bp, err := bus.NewProvider(messageBus)
	require.NoError(t, err)
	requestBus, err := bus.GetBus[*bus.ServiceListRequest](bp, bus.ServiceListRequestTopic)
	require.NoError(t, err)
	resultBus, err := bus.GetBus[*bus.ServiceListResult](bp, bus.ServiceListResultTopic)
	require.NoError(t, err)

	registry := &mockServiceRegistry{
		getAllServicesFunc: func() ([]*configv1.UpstreamServiceConfig, error) {
			panic("simulated list panic")
		},
	}

	worker := NewServiceRegistrationWorker(bp, registry)
	worker.Start(ctx)

	resultChan := make(chan *bus.ServiceListResult, 1)
	unsubscribe := resultBus.SubscribeOnce(ctx, "test-panic-list", func(result *bus.ServiceListResult) {
		resultChan <- result
	})
	defer unsubscribe()

	req := &bus.ServiceListRequest{}
	req.SetCorrelationID("test-panic-list")
	err = requestBus.Publish(ctx, "request", req)
	require.NoError(t, err)

	select {
	case result := <-resultChan:
		require.Error(t, result.Error)
		assert.Contains(t, result.Error.Error(), "panic during service list: simulated list panic")
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for list result")
	}
}

type extendedMockServiceRegistry struct {
	*mockServiceRegistry
	getServiceConfigFunc func(name string) (*configv1.UpstreamServiceConfig, bool)
}

func (m *extendedMockServiceRegistry) GetServiceConfig(name string) (*configv1.UpstreamServiceConfig, bool) {
	if m.getServiceConfigFunc != nil {
		return m.getServiceConfigFunc(name)
	}
	return nil, false
}

func TestServiceRegistrationWorker_GetServiceConfig(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	bp, err := bus.NewProvider(messageBus)
	require.NoError(t, err)
	requestBus, err := bus.GetBus[*bus.ServiceGetRequest](bp, bus.ServiceGetRequestTopic)
	require.NoError(t, err)
	resultBus, err := bus.GetBus[*bus.ServiceGetResult](bp, bus.ServiceGetResultTopic)
	require.NoError(t, err)

	// Name that needs sanitization: "My Service" -> "my_service" (assuming sanitization does something like this)
	serviceName := "My Service"

	registry := &extendedMockServiceRegistry{
		mockServiceRegistry: &mockServiceRegistry{},
		getServiceConfigFunc: func(name string) (*configv1.UpstreamServiceConfig, bool) {
			// First call with "My Service" should fail (return false)
			// Second call with sanitized name.
			// SanitizeID logic: "My Service" -> "MyService" (space removed)
			// Since length changed, it appends hash: "MyService_<hash>"
			if len(name) > 10 && name[:10] == "MyService_" {
				cfg := &configv1.UpstreamServiceConfig{}
				cfg.SetName(name)
				return cfg, true
			}
			return nil, false
		},
	}

	worker := NewServiceRegistrationWorker(bp, registry)
	worker.Start(ctx)

	resultChan := make(chan *bus.ServiceGetResult, 1)
	unsubscribe := resultBus.SubscribeOnce(ctx, "test-get", func(result *bus.ServiceGetResult) {
		resultChan <- result
	})
	defer unsubscribe()

	req := &bus.ServiceGetRequest{ServiceName: serviceName}
	req.SetCorrelationID("test-get")
	err = requestBus.Publish(ctx, "request", req)
	require.NoError(t, err)

	select {
	case result := <-resultChan:
		// If result.Error is nil, it means it found it.
		assert.NoError(t, result.Error)
		if result.Service != nil {
			assert.Contains(t, result.Service.GetName(), "MyService_")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for get result")
	}
}

func TestServiceRegistrationWorker_Panic_Get(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	bp, err := bus.NewProvider(messageBus)
	require.NoError(t, err)
	requestBus, err := bus.GetBus[*bus.ServiceGetRequest](bp, bus.ServiceGetRequestTopic)
	require.NoError(t, err)
	resultBus, err := bus.GetBus[*bus.ServiceGetResult](bp, bus.ServiceGetResultTopic)
	require.NoError(t, err)

	registry := &extendedMockServiceRegistry{
		mockServiceRegistry: &mockServiceRegistry{},
		getServiceConfigFunc: func(_ string) (*configv1.UpstreamServiceConfig, bool) {
			panic("simulated get panic")
		},
	}

	worker := NewServiceRegistrationWorker(bp, registry)
	worker.Start(ctx)

	resultChan := make(chan *bus.ServiceGetResult, 1)
	unsubscribe := resultBus.SubscribeOnce(ctx, "test-panic-get", func(result *bus.ServiceGetResult) {
		resultChan <- result
	})
	defer unsubscribe()

	req := &bus.ServiceGetRequest{ServiceName: "any"}
	req.SetCorrelationID("test-panic-get")
	err = requestBus.Publish(ctx, "request", req)
	require.NoError(t, err)

	select {
	case result := <-resultChan:
		require.Error(t, result.Error)
		assert.Contains(t, result.Error.Error(), "panic during service get: simulated get panic")
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for get result")
	}
}

func TestWorker_PublishFailure(t *testing.T) {
	t.Log("Running TestWorker_PublishFailure")

	// Setup the mock bus provider hook
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	bp, err := bus.NewProvider(messageBus)
	require.NoError(t, err)

	reqBusMock := &mockBus[*bus.ToolExecutionRequest]{}
	resBusMock := &mockBus[*bus.ToolExecutionResult]{}

	// We need to save previous hook just in case
	prevHook := bus.GetBusHook
	bus.GetBusHook = func(_ *bus.Provider, topic string) (any, error) {
		if topic == bus.ToolExecutionRequestTopic {
			return reqBusMock, nil
		}
		if topic == bus.ToolExecutionResultTopic {
			return resBusMock, nil
		}
		return nil, nil
	}
	t.Cleanup(func() {
		bus.GetBusHook = prevHook
	})

	capturedHandler := make(chan func(*bus.ToolExecutionRequest), 1)

	reqBusMock.subscribeFunc = func(_ context.Context, _ string, handler func(*bus.ToolExecutionRequest)) func() {
		capturedHandler <- handler
		return func() {}
	}

	// Mock Publish to fail
	resBusMock.publishFunc = func(ctx context.Context, _ string, _ *bus.ToolExecutionResult) error {
		return errors.New("publish failed")
	}

	worker := New(bp, &Config{MaxWorkers: 1, MaxQueueSize: 1})
	// Start calls Subscribe
	worker.Start(context.Background())
	defer worker.Stop()

	handler := <-capturedHandler

	req := &bus.ToolExecutionRequest{ToolInputs: []byte("test")}
	req.SetCorrelationID("test")

	// Execute handler
	handler(req)

	// Wait a bit for execution
	time.Sleep(100 * time.Millisecond)

	// If we are here and no panic, it passed.
}

func TestUpstreamWorker_PublishFailure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	bp, err := bus.NewProvider(messageBus)
	require.NoError(t, err)

	// Setup mocks
	reqBusMock := &mockBus[*bus.ToolExecutionRequest]{}
	resultBusMock := &mockBus[*bus.ToolExecutionResult]{}

	prevHook := bus.GetBusHook
	bus.GetBusHook = func(_ *bus.Provider, topic string) (any, error) {
		if topic == bus.ToolExecutionRequestTopic {
			return reqBusMock, nil
		}
		if topic == bus.ToolExecutionResultTopic {
			return resultBusMock, nil
		}
		return nil, nil
	}
	t.Cleanup(func() {
		bus.GetBusHook = prevHook
	})

	capturedHandler := make(chan func(*bus.ToolExecutionRequest), 1)
	reqBusMock.subscribeFunc = func(_ context.Context, _ string, handler func(*bus.ToolExecutionRequest)) func() {
		capturedHandler <- handler
		return func() {}
	}

	publishCalled := make(chan struct{})
	resultBusMock.publishFunc = func(ctx context.Context, topic string, msg *bus.ToolExecutionResult) error {
		close(publishCalled)
		return errors.New("publish error")
	}

	tm := &mockToolManager{}
	worker := NewUpstreamWorker(bp, tm)
	worker.Start(ctx)

	handler := <-capturedHandler

	req := &bus.ToolExecutionRequest{ToolName: "test"}
	req.SetCorrelationID("test")

	// Execute
	handler(req)

	select {
	case <-publishCalled:
		// success
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for publish")
	}
}

func TestWorker_Start_BusErrors(t *testing.T) {
	tests := []struct {
		name      string
		failTopic string
	}{
		{
			name:      "fail_request_bus",
			failTopic: bus.ToolExecutionRequestTopic,
		},
		{
			name:      "fail_result_bus",
			failTopic: bus.ToolExecutionResultTopic,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messageBus := bus_pb.MessageBus_builder{}.Build()
			messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
			bp, err := bus.NewProvider(messageBus)
			require.NoError(t, err)

			prevHook := bus.GetBusHook
			bus.GetBusHook = func(p *bus.Provider, topic string) (any, error) {
				if topic == tt.failTopic {
					return nil, errors.New("simulated bus error")
				}
				return nil, nil
			}
			t.Cleanup(func() {
				bus.GetBusHook = prevHook
			})

			worker := New(bp, &Config{MaxWorkers: 1, MaxQueueSize: 1})
			worker.Start(context.Background())
			worker.Stop()
		})
	}
}
