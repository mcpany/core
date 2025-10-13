/*
 * Copyright 2025 Author(s) of MCP-XY
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

package worker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/mcpxy/core/pkg/bus"
	"github.com/mcpxy/core/pkg/serviceregistry"
	"github.com/mcpxy/core/pkg/tool"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mocks

type mockServiceRegistry struct {
	serviceregistry.ServiceRegistryInterface
	registerFunc func(ctx context.Context, serviceConfig *configv1.UpstreamServiceConfig) (string, []*configv1.ToolDefinition, error)
}

func (m *mockServiceRegistry) RegisterService(ctx context.Context, serviceConfig *configv1.UpstreamServiceConfig) (string, []*configv1.ToolDefinition, error) {
	if m.registerFunc != nil {
		return m.registerFunc(ctx, serviceConfig)
	}
	return "mock-service-key", nil, nil
}

type mockToolManager struct {
	tool.ToolManagerInterface
	executeFunc func(ctx context.Context, req *tool.ExecutionRequest) (any, error)
}

func (m *mockToolManager) ExecuteTool(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, req)
	}
	return "mock-result", nil
}

func TestServiceRegistrationWorker(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("successful registration", func(t *testing.T) {
		bp := bus.NewBusProvider()
		requestBus := bus.GetBus[*bus.ServiceRegistrationRequest](bp, bus.ServiceRegistrationRequestTopic)
		resultBus := bus.GetBus[*bus.ServiceRegistrationResult](bp, bus.ServiceRegistrationResultTopic)
		var wg sync.WaitGroup
		wg.Add(1)

		registry := &mockServiceRegistry{
			registerFunc: func(ctx context.Context, serviceConfig *configv1.UpstreamServiceConfig) (string, []*configv1.ToolDefinition, error) {
				return "success-key", nil, nil
			},
		}

		worker := NewServiceRegistrationWorker(bp, registry)
		worker.Start(ctx)

		resultChan := make(chan *bus.ServiceRegistrationResult, 1)
		unsubscribe := resultBus.SubscribeOnce("test", func(result *bus.ServiceRegistrationResult) {
			resultChan <- result
			wg.Done()
		})
		defer unsubscribe()

		req := &bus.ServiceRegistrationRequest{Config: &configv1.UpstreamServiceConfig{}}
		req.SetCorrelationID("test")
		requestBus.Publish("request", req)

		wg.Wait()
		select {
		case result := <-resultChan:
			assert.NoError(t, result.Error)
			assert.Equal(t, "success-key", result.ServiceKey)
		case <-time.After(1 * time.Second):
			t.Fatal("timed out waiting for registration result")
		}
	})

	t.Run("registration failure", func(t *testing.T) {
		bp := bus.NewBusProvider()
		requestBus := bus.GetBus[*bus.ServiceRegistrationRequest](bp, bus.ServiceRegistrationRequestTopic)
		resultBus := bus.GetBus[*bus.ServiceRegistrationResult](bp, bus.ServiceRegistrationResultTopic)
		var wg sync.WaitGroup
		wg.Add(1)
		expectedErr := errors.New("registration failed")

		registry := &mockServiceRegistry{
			registerFunc: func(ctx context.Context, serviceConfig *configv1.UpstreamServiceConfig) (string, []*configv1.ToolDefinition, error) {
				return "", nil, expectedErr
			},
		}

		worker := NewServiceRegistrationWorker(bp, registry)
		worker.Start(ctx)

		resultChan := make(chan *bus.ServiceRegistrationResult, 1)
		unsubscribe := resultBus.SubscribeOnce("test-fail", func(result *bus.ServiceRegistrationResult) {
			resultChan <- result
			wg.Done()
		})
		defer unsubscribe()

		req := &bus.ServiceRegistrationRequest{Config: &configv1.UpstreamServiceConfig{}}
		req.SetCorrelationID("test-fail")
		requestBus.Publish("request", req)

		wg.Wait()
		select {
		case result := <-resultChan:
			require.Error(t, result.Error)
			assert.Contains(t, result.Error.Error(), expectedErr.Error())
		case <-time.After(1 * time.Second):
			t.Fatal("timed out waiting for registration result")
		}
	})
}

func TestUpstreamWorker(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("successful execution", func(t *testing.T) {
		bp := bus.NewBusProvider()
		requestBus := bus.GetBus[*bus.ToolExecutionRequest](bp, bus.ToolExecutionRequestTopic)
		resultBus := bus.GetBus[*bus.ToolExecutionResult](bp, bus.ToolExecutionResultTopic)
		var wg sync.WaitGroup
		wg.Add(1)

		tm := &mockToolManager{
			executeFunc: func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
				return "success", nil
			},
		}

		worker := NewUpstreamWorker(bp, tm)
		worker.Start(ctx)

		resultChan := make(chan *bus.ToolExecutionResult, 1)
		unsubscribe := resultBus.SubscribeOnce("exec-test", func(result *bus.ToolExecutionResult) {
			resultChan <- result
			wg.Done()
		})
		defer unsubscribe()

		req := &bus.ToolExecutionRequest{}
		req.SetCorrelationID("exec-test")
		requestBus.Publish("request", req)

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
		bp := bus.NewBusProvider()
		requestBus := bus.GetBus[*bus.ToolExecutionRequest](bp, bus.ToolExecutionRequestTopic)
		resultBus := bus.GetBus[*bus.ToolExecutionResult](bp, bus.ToolExecutionResultTopic)
		var wg sync.WaitGroup
		wg.Add(1)
		expectedErr := errors.New("execution failed")

		tm := &mockToolManager{
			executeFunc: func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
				return nil, expectedErr
			},
		}

		worker := NewUpstreamWorker(bp, tm)
		worker.Start(ctx)

		resultChan := make(chan *bus.ToolExecutionResult, 1)
		unsubscribe := resultBus.SubscribeOnce("exec-fail", func(result *bus.ToolExecutionResult) {
			resultChan <- result
			wg.Done()
		})
		defer unsubscribe()

		req := &bus.ToolExecutionRequest{}
		req.SetCorrelationID("exec-fail")
		requestBus.Publish("request", req)

		wg.Wait()
		select {
		case result := <-resultChan:
			assert.Error(t, result.Error)
			assert.Equal(t, expectedErr, result.Error)
		case <-time.After(1 * time.Second):
			t.Fatal("timed out waiting for execution result")
		}
	})
}

func TestServiceRegistrationWorker_Concurrent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bp := bus.NewBusProvider()
	requestBus := bus.GetBus[*bus.ServiceRegistrationRequest](bp, bus.ServiceRegistrationRequestTopic)
	resultBus := bus.GetBus[*bus.ServiceRegistrationResult](bp, bus.ServiceRegistrationResultTopic)

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
			unsubscribe := resultBus.SubscribeOnce(correlationID, func(result *bus.ServiceRegistrationResult) {
				resultChan <- result
			})
			defer unsubscribe()

			requestBus.Publish("request", req)

			select {
			case result := <-resultChan:
				assert.NoError(t, result.Error)
				assert.Equal(t, "mock-service-key", result.ServiceKey)
			case <-time.After(2 * time.Second):
				t.Errorf("timed out waiting for registration result for request %d", i)
			}
		}(i)
	}

	wg.Wait()
}

func TestUpstreamWorker_Concurrent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bp := bus.NewBusProvider()
	requestBus := bus.GetBus[*bus.ToolExecutionRequest](bp, bus.ToolExecutionRequestTopic)
	resultBus := bus.GetBus[*bus.ToolExecutionResult](bp, bus.ToolExecutionResultTopic)

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
			unsubscribe := resultBus.SubscribeOnce(correlationID, func(result *bus.ToolExecutionResult) {
				resultChan <- result
			})
			defer unsubscribe()

			requestBus.Publish("request", req)

			select {
			case result := <-resultChan:
				assert.NoError(t, result.Error)
				assert.JSONEq(t, `"mock-result"`, string(result.Result))
			case <-time.After(2 * time.Second):
				t.Errorf("timed out waiting for execution result for request %d", i)
			}
		}(i)
	}

	wg.Wait()
}
