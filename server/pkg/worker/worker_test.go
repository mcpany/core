// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package worker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	bus_pb "github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mocks

type mockBus[T any] struct {
	bus.Bus[T]
	publishFunc   func(ctx context.Context, topic string, msg T) error
	subscribeFunc func(ctx context.Context, topic string, handler func(T)) func()
}

func (m *mockBus[T]) Publish(ctx context.Context, topic string, msg T) error {
	if m.publishFunc != nil {
		return m.publishFunc(ctx, topic, msg)
	}
	return nil
}

func (m *mockBus[T]) Subscribe(ctx context.Context, topic string, handler func(T)) func() {
	if m.subscribeFunc != nil {
		return m.subscribeFunc(ctx, topic, handler)
	}
	// Return a no-op unsubscribe function
	return func() {}
}

type mockServiceRegistry struct {
	serviceregistry.ServiceRegistryInterface
	registerFunc    func(ctx context.Context, serviceConfig *configv1.UpstreamServiceConfig) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error)
	registerResFunc func(ctx context.Context, resourceConfig *configv1.ResourceDefinition) error
	getAllServicesFunc func() ([]*configv1.UpstreamServiceConfig, error)
}

func (m *mockServiceRegistry) RegisterService(ctx context.Context, serviceConfig *configv1.UpstreamServiceConfig) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	if m.registerFunc != nil {
		return m.registerFunc(ctx, serviceConfig)
	}
	return "mock-service-key", nil, nil, nil
}

func (m *mockServiceRegistry) RegisterResource(ctx context.Context, resourceConfig *configv1.ResourceDefinition) error {
	if m.registerResFunc != nil {
		return m.registerResFunc(ctx, resourceConfig)
	}
	return nil
}

func (m *mockServiceRegistry) GetAllServices() ([]*configv1.UpstreamServiceConfig, error) {
	if m.getAllServicesFunc != nil {
		return m.getAllServicesFunc()
	}
	return nil, nil
}

func (m *mockServiceRegistry) GetServiceHealth(serviceID string) (string, bool) {
	return "healthy", true
}

type mockToolManager struct {
	tool.ManagerInterface
	executeFunc func(ctx context.Context, req *tool.ExecutionRequest) (any, error)
}

func (m *mockToolManager) ExecuteTool(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, req)
	}
	return "mock-result", nil
}

func (m *mockToolManager) AddTool(_ tool.Tool) error {
	return nil
}

func (m *mockToolManager) GetTool(_ string) (tool.Tool, bool) {
	return nil, false
}

func (m *mockToolManager) ListServices() []*tool.ServiceInfo {
	return nil
}

func (m *mockToolManager) ListTools() []tool.Tool {
	return nil
}

func (m *mockToolManager) ClearToolsForService(_ string) {
}

func (m *mockToolManager) SetProfiles(_ []string, _ []*configv1.ProfileDefinition) {
}

func (m *mockToolManager) SetMCPServer(_ tool.MCPServerProvider) {
}

func (m *mockToolManager) AddMiddleware(_ tool.ExecutionMiddleware) {
}

func (m *mockToolManager) AddServiceInfo(_ string, _ *tool.ServiceInfo) {
}

func (m *mockToolManager) GetServiceInfo(_ string) (*tool.ServiceInfo, bool) {
	return nil, false
}

func TestServiceRegistrationWorker(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("successful registration", func(t *testing.T) {
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
				return "success-key", nil, []*configv1.ResourceDefinition{}, nil
			},
		}

		worker := NewServiceRegistrationWorker(bp, registry)
		worker.Start(ctx)

		resultChan := make(chan *bus.ServiceRegistrationResult, 1)
		unsubscribe := resultBus.SubscribeOnce(ctx, "test", func(result *bus.ServiceRegistrationResult) {
			resultChan <- result
		})
		defer unsubscribe()

		req := &bus.ServiceRegistrationRequest{Config: &configv1.UpstreamServiceConfig{}}
		req.SetCorrelationID("test")
		err = requestBus.Publish(ctx, "request", req)
		require.NoError(t, err)

		select {
		case result := <-resultChan:
			assert.NoError(t, result.Error)
			assert.Equal(t, "success-key", result.ServiceKey)
		case <-time.After(1 * time.Second):
			t.Fatal("timed out waiting for registration result")
		}
	})

	t.Run("registration failure", func(t *testing.T) {
		messageBus := bus_pb.MessageBus_builder{}.Build()
		messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
		bp, err := bus.NewProvider(messageBus)
		require.NoError(t, err)
		requestBus, err := bus.GetBus[*bus.ServiceRegistrationRequest](bp, bus.ServiceRegistrationRequestTopic)
		require.NoError(t, err)
		resultBus, err := bus.GetBus[*bus.ServiceRegistrationResult](bp, bus.ServiceRegistrationResultTopic)
		require.NoError(t, err)
		expectedErr := errors.New("registration failed")

		registry := &mockServiceRegistry{
			registerFunc: func(_ context.Context, _ *configv1.UpstreamServiceConfig) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
				return "", nil, nil, expectedErr
			},
		}

		worker := NewServiceRegistrationWorker(bp, registry)
		worker.Start(ctx)

		resultChan := make(chan *bus.ServiceRegistrationResult, 1)
		unsubscribe := resultBus.SubscribeOnce(ctx, "test-fail", func(result *bus.ServiceRegistrationResult) {
			resultChan <- result
		})
		defer unsubscribe()

		req := &bus.ServiceRegistrationRequest{Config: &configv1.UpstreamServiceConfig{}}
		req.SetCorrelationID("test-fail")
		err = requestBus.Publish(ctx, "request", req)
		require.NoError(t, err)

		select {
		case result := <-resultChan:
			require.Error(t, result.Error)
			assert.Contains(t, result.Error.Error(), expectedErr.Error())
		case <-time.After(1 * time.Second):
			t.Fatal("timed out waiting for registration result")
		}
	})

	t.Run("uses request context", func(t *testing.T) {
		messageBus := bus_pb.MessageBus_builder{}.Build()
		messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
		bp, err := bus.NewProvider(messageBus)
		require.NoError(t, err)
		requestBus, err := bus.GetBus[*bus.ServiceRegistrationRequest](bp, bus.ServiceRegistrationRequestTopic)
		require.NoError(t, err)
		resultBus, err := bus.GetBus[*bus.ServiceRegistrationResult](bp, bus.ServiceRegistrationResultTopic)
		require.NoError(t, err)

		// Use a distinct value key for context verification
		type key string
		const requestKey key = "is_request_context"

		registry := &mockServiceRegistry{
			registerFunc: func(ctx context.Context, _ *configv1.UpstreamServiceConfig) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
				// Verify the context has the value we put in the request context
				if val, ok := ctx.Value(requestKey).(bool); ok && val {
					return "success-key-request-context", nil, nil, nil
				}
				return "", nil, nil, errors.New("incorrect context used")
			},
		}

		// Start worker with background context
		worker := NewServiceRegistrationWorker(bp, registry)
		worker.Start(context.Background())

		resultChan := make(chan *bus.ServiceRegistrationResult, 1)
		unsubscribe := resultBus.SubscribeOnce(ctx, "test-req-ctx", func(result *bus.ServiceRegistrationResult) {
			resultChan <- result
		})
		defer unsubscribe()

		// Create a request context with the special value
		reqCtx := context.WithValue(context.Background(), requestKey, true)
		req := &bus.ServiceRegistrationRequest{
			Context: reqCtx,
			Config:  &configv1.UpstreamServiceConfig{},
		}
		req.SetCorrelationID("test-req-ctx")
		err = requestBus.Publish(ctx, "request", req)
		require.NoError(t, err)

		select {
		case result := <-resultChan:
			assert.NoError(t, result.Error)
			assert.Equal(t, "success-key-request-context", result.ServiceKey)
		case <-time.After(1 * time.Second):
			t.Fatal("timed out waiting for registration result")
		}
	})
}

func TestUpstreamWorker(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("successful execution", func(t *testing.T) {
		messageBus := bus_pb.MessageBus_builder{}.Build()
		messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
		bp, err := bus.NewProvider(messageBus)
		require.NoError(t, err)
		requestBus, err := bus.GetBus[*bus.ToolExecutionRequest](bp, bus.ToolExecutionRequestTopic)
		require.NoError(t, err)
		resultBus, err := bus.GetBus[*bus.ToolExecutionResult](bp, bus.ToolExecutionResultTopic)
		require.NoError(t, err)
		var wg sync.WaitGroup
		wg.Add(1)

		tm := &mockToolManager{
			executeFunc: func(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
				return "success", nil
			},
		}

		worker := NewUpstreamWorker(bp, tm)
		worker.Start(ctx)

		resultChan := make(chan *bus.ToolExecutionResult, 1)
		unsubscribe := resultBus.SubscribeOnce(ctx, "exec-test", func(result *bus.ToolExecutionResult) {
			resultChan <- result
			wg.Done()
		})
		defer unsubscribe()

		req := &bus.ToolExecutionRequest{}
		req.SetCorrelationID("exec-test")
		err = requestBus.Publish(ctx, "request", req)
		require.NoError(t, err)

		wg.Wait()
		select {
		case result := <-resultChan:
			assert.NoError(t, result.Error)
			assert.JSONEq(t, `"success"`, string(result.Result))
		case <-time.After(1 * time.Second):
			t.Fatal("timed out waiting for execution result")
		}
	})

	t.Run("execution failure", func(t *testing.T) {
		messageBus := bus_pb.MessageBus_builder{}.Build()
		messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
		bp, err := bus.NewProvider(messageBus)
		require.NoError(t, err)
		requestBus, err := bus.GetBus[*bus.ToolExecutionRequest](bp, bus.ToolExecutionRequestTopic)
		require.NoError(t, err)
		resultBus, err := bus.GetBus[*bus.ToolExecutionResult](bp, bus.ToolExecutionResultTopic)
		require.NoError(t, err)
		var wg sync.WaitGroup
		wg.Add(1)
		expectedErr := errors.New("execution failed")

		tm := &mockToolManager{
			executeFunc: func(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
				return nil, expectedErr
			},
		}

		worker := NewUpstreamWorker(bp, tm)
		worker.Start(ctx)

		resultChan := make(chan *bus.ToolExecutionResult, 1)
		unsubscribe := resultBus.SubscribeOnce(ctx, "exec-fail", func(result *bus.ToolExecutionResult) {
			resultChan <- result
			wg.Done()
		})
		defer unsubscribe()

		req := &bus.ToolExecutionRequest{}
		req.SetCorrelationID("exec-fail")
		err = requestBus.Publish(ctx, "request", req)
		require.NoError(t, err)

		wg.Wait()
		select {
		case result := <-resultChan:
			assert.Error(t, result.Error)
			assert.Equal(t, expectedErr, result.Error)
		case <-time.After(1 * time.Second):
			t.Fatal("timed out waiting for execution result")
		}
	})

	t.Run("result marshaling failure", func(t *testing.T) {
		messageBus := bus_pb.MessageBus_builder{}.Build()
		messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
		bp, err := bus.NewProvider(messageBus)
		require.NoError(t, err)
		requestBus, err := bus.GetBus[*bus.ToolExecutionRequest](bp, bus.ToolExecutionRequestTopic)
		require.NoError(t, err)
		resultBus, err := bus.GetBus[*bus.ToolExecutionResult](bp, bus.ToolExecutionResultTopic)
		require.NoError(t, err)
		var wg sync.WaitGroup
		wg.Add(1)

		tm := &mockToolManager{
			executeFunc: func(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
				// Functions are not serializable to JSON
				return func() {}, nil
			},
		}

		worker := NewUpstreamWorker(bp, tm)
		worker.Start(ctx)

		resultChan := make(chan *bus.ToolExecutionResult, 1)
		unsubscribe := resultBus.SubscribeOnce(ctx, "marshal-fail-test", func(result *bus.ToolExecutionResult) {
			resultChan <- result
			wg.Done()
		})
		defer unsubscribe()

		req := &bus.ToolExecutionRequest{}
		req.SetCorrelationID("marshal-fail-test")
		err = requestBus.Publish(ctx, "request", req)
		require.NoError(t, err)

		wg.Wait()
		select {
		case result := <-resultChan:
			assert.Error(t, result.Error)
			assert.Contains(t, result.Error.Error(), "json: unsupported type: func()")
			assert.Nil(t, result.Result, "Result should be nil due to marshaling failure")
		case <-time.After(1 * time.Second):
			t.Fatal("timed out waiting for execution result")
		}
	})

	t.Run("execution with partial result and error", func(t *testing.T) {
		messageBus := bus_pb.MessageBus_builder{}.Build()
		messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
		bp, err := bus.NewProvider(messageBus)
		require.NoError(t, err)
		requestBus, err := bus.GetBus[*bus.ToolExecutionRequest](bp, bus.ToolExecutionRequestTopic)
		require.NoError(t, err)
		resultBus, err := bus.GetBus[*bus.ToolExecutionResult](bp, bus.ToolExecutionResultTopic)
		require.NoError(t, err)
		var wg sync.WaitGroup
		wg.Add(1)
		expectedErr := errors.New("execution partially failed")

		tm := &mockToolManager{
			executeFunc: func(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
				return "partial result", expectedErr
			},
		}

		worker := NewUpstreamWorker(bp, tm)
		worker.Start(ctx)

		resultChan := make(chan *bus.ToolExecutionResult, 1)
		unsubscribe := resultBus.SubscribeOnce(ctx, "exec-partial-fail", func(result *bus.ToolExecutionResult) {
			resultChan <- result
			wg.Done()
		})
		defer unsubscribe()

		req := &bus.ToolExecutionRequest{}
		req.SetCorrelationID("exec-partial-fail")
		err = requestBus.Publish(ctx, "request", req)
		require.NoError(t, err)

		wg.Wait()
		select {
		case result := <-resultChan:
			assert.Error(t, result.Error)
			assert.Equal(t, expectedErr, result.Error)
			assert.JSONEq(t, `"partial result"`, string(result.Result))
		case <-time.After(1 * time.Second):
			t.Fatal("timed out waiting for execution result")
		}
	})
}

func TestServiceRegistrationWorker_Concurrent(t *testing.T) {
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

	registry := &mockServiceRegistry{}
	worker := NewServiceRegistrationWorker(bp, registry)
	worker.Start(ctx)

	numRequests := 100
	var wg sync.WaitGroup
	wg.Add(numRequests)

	for i := 0; i < numRequests; i++ {
		go func(i int) {
			defer wg.Done()
			req := &bus.ServiceRegistrationRequest{Config: &configv1.UpstreamServiceConfig{}}
			correlationID := fmt.Sprintf("test-%d", i)
			req.SetCorrelationID(correlationID)

			resultChan := make(chan *bus.ServiceRegistrationResult, 1)
			unsubscribe := resultBus.SubscribeOnce(ctx, correlationID, func(result *bus.ServiceRegistrationResult) {
				resultChan <- result
			})
			defer unsubscribe()

			err := requestBus.Publish(ctx, "request", req)
			require.NoError(t, err)

			select {
			case result := <-resultChan:
				assert.NoError(t, result.Error)
				assert.Equal(t, "mock-service-key", result.ServiceKey)
			case <-time.After(5 * time.Second):
				t.Errorf("timed out waiting for registration result for request %d", i)
			}
		}(i)
	}

	wg.Wait()
}

func TestWorker_ContextPropagation(t *testing.T) {
	t.Log("Running TestWorker_ContextPropagation")
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	bp, err := bus.NewProvider(messageBus)
	require.NoError(t, err)

	reqBusMock := &mockBus[*bus.ToolExecutionRequest]{}
	resBusMock := &mockBus[*bus.ToolExecutionResult]{}

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
		bus.GetBusHook = nil
	})

	var wg sync.WaitGroup
	wg.Add(1)

	readyToPublish := make(chan struct{})
	var capturedHandler func(*bus.ToolExecutionRequest)

	reqBusMock.subscribeFunc = func(_ context.Context, _ string, handler func(*bus.ToolExecutionRequest)) func() {
		capturedHandler = handler
		close(readyToPublish)
		return func() {}
	}

	resBusMock.publishFunc = func(ctx context.Context, _ string, _ *bus.ToolExecutionResult) error {
		defer wg.Done()
		// Block until context is canceled. This proves the correct context was passed.
		// If context.Background() was passed, this will block forever and the test will time out.
		<-ctx.Done()
		require.Error(t, ctx.Err(), "Context should be canceled")
		return nil
	}

	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()

	worker := New(bp, &Config{MaxWorkers: 1, MaxQueueSize: 1})
	worker.Start(workerCtx)

	<-readyToPublish // Wait for subscription

	req := &bus.ToolExecutionRequest{}
	req.SetCorrelationID("test")
	go capturedHandler(req) // Simulate message arrival

	// Give the worker goroutine time to run and block inside publishFunc
	time.Sleep(100 * time.Millisecond)

	// Now, cancel the context
	workerCancel()

	// Wait for publishFunc to complete its checks
	wg.Wait()
}

func TestUpstreamWorker_Concurrent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	bp, err := bus.NewProvider(messageBus)
	require.NoError(t, err)
	requestBus, err := bus.GetBus[*bus.ToolExecutionRequest](bp, bus.ToolExecutionRequestTopic)
	require.NoError(t, err)
	resultBus, err := bus.GetBus[*bus.ToolExecutionResult](bp, bus.ToolExecutionResultTopic)
	require.NoError(t, err)

	tm := &mockToolManager{}
	worker := NewUpstreamWorker(bp, tm)
	worker.Start(ctx)

	numRequests := 100
	var wg sync.WaitGroup
	wg.Add(numRequests)

	for i := 0; i < numRequests; i++ {
		go func(i int) {
			defer wg.Done()
			req := &bus.ToolExecutionRequest{}
			correlationID := fmt.Sprintf("exec-test-%d", i)
			req.SetCorrelationID(correlationID)

			resultChan := make(chan *bus.ToolExecutionResult, 1)
			unsubscribe := resultBus.SubscribeOnce(ctx, correlationID, func(result *bus.ToolExecutionResult) {
				resultChan <- result
			})
			defer unsubscribe()

			err := requestBus.Publish(ctx, "request", req)
			require.NoError(t, err)

			select {
			case result := <-resultChan:
				assert.NoError(t, result.Error)
				assert.JSONEq(t, `"mock-result"`, string(result.Result))
			case <-time.After(5 * time.Second):
				t.Errorf("timed out waiting for execution result for request %d", i)
			}
		}(i)
	}

	wg.Wait()
}

func ptr[T any](v T) *T {
	return &v
}

func TestServiceRegistrationWorker_ListRequest(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("list services", func(t *testing.T) {
		messageBus := bus_pb.MessageBus_builder{}.Build()
		messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
		bp, err := bus.NewProvider(messageBus)
		require.NoError(t, err)
		requestBus, err := bus.GetBus[*bus.ServiceListRequest](bp, bus.ServiceListRequestTopic)
		require.NoError(t, err)
		resultBus, err := bus.GetBus[*bus.ServiceListResult](bp, bus.ServiceListResultTopic)
		require.NoError(t, err)

		s1 := &configv1.UpstreamServiceConfig{}
		s1.SetName("service1")
		expectedServices := []*configv1.UpstreamServiceConfig{s1}

		registry := &mockServiceRegistry{
			getAllServicesFunc: func() ([]*configv1.UpstreamServiceConfig, error) {
				return expectedServices, nil
			},
		}

		worker := NewServiceRegistrationWorker(bp, registry)
		worker.Start(ctx)

		resultChan := make(chan *bus.ServiceListResult, 1)
		unsubscribe := resultBus.SubscribeOnce(ctx, "list-test", func(result *bus.ServiceListResult) {
			resultChan <- result
		})
		defer unsubscribe()

		req := &bus.ServiceListRequest{}
		req.SetCorrelationID("list-test")
		err = requestBus.Publish(ctx, "request", req)
		require.NoError(t, err)

		select {
		case result := <-resultChan:
			assert.NoError(t, result.Error)
			assert.Equal(t, expectedServices, result.Services)
		case <-time.After(1 * time.Second):
			t.Fatal("timed out waiting for list result")
		}
	})
}
