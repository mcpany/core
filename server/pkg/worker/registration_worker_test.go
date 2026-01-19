// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package worker_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/worker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
)

func strPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

// MockServiceRegistry is a comprehensive mock for serviceregistry.ServiceRegistryInterface
type MockServiceRegistry struct {
	serviceregistry.ServiceRegistryInterface // Embed to satisfy interface, but override methods we use

	registerFunc       func(ctx context.Context, config *configv1.UpstreamServiceConfig) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error)
	unregisterFunc     func(ctx context.Context, name string) error
	getAllServicesFunc func() ([]*configv1.UpstreamServiceConfig, error)
	getServiceConfigFunc func(serviceID string) (*configv1.UpstreamServiceConfig, bool)
}

func (m *MockServiceRegistry) RegisterService(ctx context.Context, config *configv1.UpstreamServiceConfig) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	if m.registerFunc != nil {
		return m.registerFunc(ctx, config)
	}
	return "service1", nil, nil, nil
}

func (m *MockServiceRegistry) UnregisterService(ctx context.Context, name string) error {
	if m.unregisterFunc != nil {
		return m.unregisterFunc(ctx, name)
	}
	return nil
}

<<<<<<< HEAD
func (m *MockServiceRegistry) GetServiceStatus(serviceID string) string {
	return "OK"
=======
func (m *MockServiceRegistry) GetAllServices() ([]*configv1.UpstreamServiceConfig, error) {
	if m.getAllServicesFunc != nil {
		return m.getAllServicesFunc()
	}
	return nil, nil
}

func (m *MockServiceRegistry) GetServiceConfig(serviceID string) (*configv1.UpstreamServiceConfig, bool) {
	if m.getServiceConfigFunc != nil {
		return m.getServiceConfigFunc(serviceID)
	}
	return nil, false
}

func (m *MockServiceRegistry) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	return nil, false
}

func (m *MockServiceRegistry) GetServiceError(serviceID string) (string, bool) {
	return "", false
>>>>>>> origin/main
}

func TestServiceRegistrationWorker_Stop(t *testing.T) {
	// Setup bus
	b, err := bus.NewProvider(nil)
	require.NoError(t, err)

	// Setup worker
	registry := &MockServiceRegistry{}
	w := worker.NewServiceRegistrationWorker(b, registry)

	ctx, cancel := context.WithCancel(context.Background())
	w.Start(ctx)

	// Ensure it started (async)
	time.Sleep(10 * time.Millisecond)

	// Test Stop (graceful shutdown)
	cancel()
	w.Stop()

	// If we reached here, it didn't deadlock
	assert.True(t, true)
}

func TestServiceRegistrationWorker_Register_Success(t *testing.T) {
	b, err := bus.NewProvider(nil)
	require.NoError(t, err)

	registry := &MockServiceRegistry{
		registerFunc: func(ctx context.Context, config *configv1.UpstreamServiceConfig) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
			assert.Equal(t, "test-service", config.GetName())
			return "test-service-id", []*configv1.ToolDefinition{{Name: strPtr("tool1")}}, []*configv1.ResourceDefinition{{Uri: strPtr("res1")}}, nil
		},
	}
	w := worker.NewServiceRegistrationWorker(b, registry)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	w.Start(ctx)

	resultBus, err := bus.GetBus[*bus.ServiceRegistrationResult](b, bus.ServiceRegistrationResultTopic)
	require.NoError(t, err)

	correlationID := "123"
	resChan := make(chan *bus.ServiceRegistrationResult, 1)
	unsubscribe := resultBus.Subscribe(ctx, correlationID, func(res *bus.ServiceRegistrationResult) {
		resChan <- res
	})
	defer unsubscribe()

	requestBus, err := bus.GetBus[*bus.ServiceRegistrationRequest](b, bus.ServiceRegistrationRequestTopic)
	require.NoError(t, err)

	req := &bus.ServiceRegistrationRequest{
		Config: &configv1.UpstreamServiceConfig{
			Name: strPtr("test-service"),
		},
	}
	req.SetCorrelationID(correlationID)
	err = requestBus.Publish(ctx, "request", req)
	require.NoError(t, err)

	select {
	case res := <-resChan:
		assert.Equal(t, correlationID, res.CorrelationID())
		assert.NoError(t, res.Error)
		assert.Equal(t, "test-service-id", res.ServiceKey)
		assert.Len(t, res.DiscoveredTools, 1)
		assert.Len(t, res.DiscoveredResources, 1)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for registration result")
	}
}

func TestServiceRegistrationWorker_Register_Failure(t *testing.T) {
	b, err := bus.NewProvider(nil)
	require.NoError(t, err)

	expectedErr := fmt.Errorf("registration failed")
	registry := &MockServiceRegistry{
		registerFunc: func(ctx context.Context, config *configv1.UpstreamServiceConfig) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
			return "", nil, nil, expectedErr
		},
	}
	w := worker.NewServiceRegistrationWorker(b, registry)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	w.Start(ctx)

	resultBus, err := bus.GetBus[*bus.ServiceRegistrationResult](b, bus.ServiceRegistrationResultTopic)
	require.NoError(t, err)

	correlationID := "fail-id"
	resChan := make(chan *bus.ServiceRegistrationResult, 1)
	unsubscribe := resultBus.Subscribe(ctx, correlationID, func(res *bus.ServiceRegistrationResult) {
		resChan <- res
	})
	defer unsubscribe()

	requestBus, err := bus.GetBus[*bus.ServiceRegistrationRequest](b, bus.ServiceRegistrationRequestTopic)
	require.NoError(t, err)

	req := &bus.ServiceRegistrationRequest{
		Config: &configv1.UpstreamServiceConfig{
			Name: strPtr("fail-service"),
		},
	}
	req.SetCorrelationID(correlationID)
	err = requestBus.Publish(ctx, "request", req)
	require.NoError(t, err)

	select {
	case res := <-resChan:
		assert.Equal(t, correlationID, res.CorrelationID())
		assert.ErrorIs(t, res.Error, expectedErr)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for registration result")
	}
}

func TestServiceRegistrationWorker_Unregister(t *testing.T) {
	b, err := bus.NewProvider(nil)
	require.NoError(t, err)

	var unregisterCalled bool
	var mu sync.Mutex

	registry := &MockServiceRegistry{
		unregisterFunc: func(ctx context.Context, name string) error {
			mu.Lock()
			unregisterCalled = true
			mu.Unlock()
			assert.Equal(t, "disabled-service", name)
			return nil
		},
	}
	w := worker.NewServiceRegistrationWorker(b, registry)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	w.Start(ctx)

	resultBus, err := bus.GetBus[*bus.ServiceRegistrationResult](b, bus.ServiceRegistrationResultTopic)
	require.NoError(t, err)

	correlationID := "disable-id"
	resChan := make(chan *bus.ServiceRegistrationResult, 1)
	unsubscribe := resultBus.Subscribe(ctx, correlationID, func(res *bus.ServiceRegistrationResult) {
		resChan <- res
	})
	defer unsubscribe()

	requestBus, err := bus.GetBus[*bus.ServiceRegistrationRequest](b, bus.ServiceRegistrationRequestTopic)
	require.NoError(t, err)

	req := &bus.ServiceRegistrationRequest{
		Config: &configv1.UpstreamServiceConfig{
			Name:    strPtr("disabled-service"),
			Disable: boolPtr(true),
		},
	}
	req.SetCorrelationID(correlationID)
	err = requestBus.Publish(ctx, "request", req)
	require.NoError(t, err)

	select {
	case res := <-resChan:
		assert.Equal(t, correlationID, res.CorrelationID())
		assert.NoError(t, res.Error)
		mu.Lock()
		assert.True(t, unregisterCalled)
		mu.Unlock()
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for unregistration result")
	}
}

func TestServiceRegistrationWorker_List(t *testing.T) {
	b, err := bus.NewProvider(nil)
	require.NoError(t, err)

	services := []*configv1.UpstreamServiceConfig{
		{Name: strPtr("s1")},
		{Name: strPtr("s2")},
	}
	registry := &MockServiceRegistry{
		getAllServicesFunc: func() ([]*configv1.UpstreamServiceConfig, error) {
			return services, nil
		},
	}
	w := worker.NewServiceRegistrationWorker(b, registry)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	w.Start(ctx)

	resultBus, err := bus.GetBus[*bus.ServiceListResult](b, bus.ServiceListResultTopic)
	require.NoError(t, err)

	correlationID := "list-id"
	resChan := make(chan *bus.ServiceListResult, 1)
	unsubscribe := resultBus.Subscribe(ctx, correlationID, func(res *bus.ServiceListResult) {
		resChan <- res
	})
	defer unsubscribe()

	requestBus, err := bus.GetBus[*bus.ServiceListRequest](b, bus.ServiceListRequestTopic)
	require.NoError(t, err)

	req := &bus.ServiceListRequest{}
	req.SetCorrelationID(correlationID)
	err = requestBus.Publish(ctx, "request", req)
	require.NoError(t, err)

	select {
	case res := <-resChan:
		assert.Equal(t, correlationID, res.CorrelationID())
		assert.NoError(t, res.Error)
		assert.Equal(t, services, res.Services)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for list result")
	}
}

func TestServiceRegistrationWorker_Get(t *testing.T) {
	b, err := bus.NewProvider(nil)
	require.NoError(t, err)

	service := &configv1.UpstreamServiceConfig{Name: strPtr("my-service")}
	registry := &MockServiceRegistry{
		getServiceConfigFunc: func(serviceID string) (*configv1.UpstreamServiceConfig, bool) {
			if serviceID == "my-service" {
				return service, true
			}
			return nil, false
		},
	}
	w := worker.NewServiceRegistrationWorker(b, registry)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	w.Start(ctx)

	resultBus, err := bus.GetBus[*bus.ServiceGetResult](b, bus.ServiceGetResultTopic)
	require.NoError(t, err)

	correlationID := "get-id"
	resChan := make(chan *bus.ServiceGetResult, 1)
	unsubscribe := resultBus.Subscribe(ctx, correlationID, func(res *bus.ServiceGetResult) {
		resChan <- res
	})
	defer unsubscribe()

	requestBus, err := bus.GetBus[*bus.ServiceGetRequest](b, bus.ServiceGetRequestTopic)
	require.NoError(t, err)

	req := &bus.ServiceGetRequest{ServiceName: "my-service"}
	req.SetCorrelationID(correlationID)
	err = requestBus.Publish(ctx, "request", req)
	require.NoError(t, err)

	select {
	case res := <-resChan:
		assert.Equal(t, correlationID, res.CorrelationID())
		assert.NoError(t, res.Error)
		assert.Equal(t, service, res.Service)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for get result")
	}
}

func TestServiceRegistrationWorker_Get_NotFound(t *testing.T) {
	b, err := bus.NewProvider(nil)
	require.NoError(t, err)

	registry := &MockServiceRegistry{
		getServiceConfigFunc: func(serviceID string) (*configv1.UpstreamServiceConfig, bool) {
			return nil, false
		},
	}
	w := worker.NewServiceRegistrationWorker(b, registry)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	w.Start(ctx)

	resultBus, err := bus.GetBus[*bus.ServiceGetResult](b, bus.ServiceGetResultTopic)
	require.NoError(t, err)

	correlationID := "notfound-id"
	resChan := make(chan *bus.ServiceGetResult, 1)
	unsubscribe := resultBus.Subscribe(ctx, correlationID, func(res *bus.ServiceGetResult) {
		resChan <- res
	})
	defer unsubscribe()

	requestBus, err := bus.GetBus[*bus.ServiceGetRequest](b, bus.ServiceGetRequestTopic)
	require.NoError(t, err)

	req := &bus.ServiceGetRequest{ServiceName: "unknown"}
	req.SetCorrelationID(correlationID)
	err = requestBus.Publish(ctx, "request", req)
	require.NoError(t, err)

	select {
	case res := <-resChan:
		assert.Equal(t, correlationID, res.CorrelationID())
		assert.Error(t, res.Error)
		assert.Nil(t, res.Service)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for get result")
	}
}

func TestServiceRegistrationWorker_Register_Timeout(t *testing.T) {
	b, err := bus.NewProvider(nil)
	require.NoError(t, err)

	registry := &MockServiceRegistry{
		registerFunc: func(ctx context.Context, config *configv1.UpstreamServiceConfig) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
			// Check if deadline is set
			deadline, ok := ctx.Deadline()
			assert.True(t, ok, "context should have deadline")
			if ok {
				// We can't strictly assert the duration is exactly what we passed due to processing time,
				// but it should be in the future.
				assert.True(t, time.Until(deadline) > 0)
			}
			return "timeout-service", nil, nil, nil
		},
	}
	w := worker.NewServiceRegistrationWorker(b, registry)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	w.Start(ctx)

	resultBus, err := bus.GetBus[*bus.ServiceRegistrationResult](b, bus.ServiceRegistrationResultTopic)
	require.NoError(t, err)

	correlationID := "timeout-id"
	resChan := make(chan *bus.ServiceRegistrationResult, 1)
	unsubscribe := resultBus.Subscribe(ctx, correlationID, func(res *bus.ServiceRegistrationResult) {
		resChan <- res
	})
	defer unsubscribe()

	requestBus, err := bus.GetBus[*bus.ServiceRegistrationRequest](b, bus.ServiceRegistrationRequestTopic)
	require.NoError(t, err)

	req := &bus.ServiceRegistrationRequest{
		Config: &configv1.UpstreamServiceConfig{
			Name: strPtr("timeout-service"),
			Resilience: &configv1.ResilienceConfig{
				Timeout: durationpb.New(500 * time.Millisecond),
			},
		},
	}
	req.SetCorrelationID(correlationID)
	err = requestBus.Publish(ctx, "request", req)
	require.NoError(t, err)

	select {
	case res := <-resChan:
		assert.Equal(t, correlationID, res.CorrelationID())
		assert.NoError(t, res.Error)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for registration result")
	}
}
