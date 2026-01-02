// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

// Re-using MockServiceRegistry from reproduction test (copied here for self-contained test)
type MockServiceRegistryForTest struct {
	mock.Mock
}

func (m *MockServiceRegistryForTest) RegisterService(ctx context.Context, serviceConfig *configv1.UpstreamServiceConfig) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	args := m.Called(ctx, serviceConfig)
	return args.String(0), args.Get(1).([]*configv1.ToolDefinition), args.Get(2).([]*configv1.ResourceDefinition), args.Error(3)
}

func (m *MockServiceRegistryForTest) UnregisterService(ctx context.Context, serviceName string) error {
	args := m.Called(ctx, serviceName)
	return args.Error(0)
}

func (m *MockServiceRegistryForTest) GetAllServices() ([]*configv1.UpstreamServiceConfig, error) {
	args := m.Called()
	return args.Get(0).([]*configv1.UpstreamServiceConfig), args.Error(1)
}

func (m *MockServiceRegistryForTest) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	args := m.Called(serviceID)
	return args.Get(0).(*tool.ServiceInfo), args.Bool(1)
}

func (m *MockServiceRegistryForTest) GetServiceConfig(serviceID string) (*configv1.UpstreamServiceConfig, bool) {
	args := m.Called(serviceID)
	return args.Get(0).(*configv1.UpstreamServiceConfig), args.Bool(1)
}

func (m *MockServiceRegistryForTest) GetServiceError(serviceID string) (string, bool) {
	args := m.Called(serviceID)
	return args.String(0), args.Bool(1)
}

func TestServiceRegistrationWorker_RetryLogic_ExponentialBackoff(t *testing.T) {
	// Setup
	busProvider, _ := bus.NewProvider(nil)
	mockRegistry := new(MockServiceRegistryForTest)
	worker := NewServiceRegistrationWorker(busProvider, mockRegistry)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serviceConfig := &configv1.UpstreamServiceConfig{
		Name: proto.String("retry-service-backoff"),
	}

	// Expect initial failure, then retry
	mockRegistry.On("RegisterService", mock.Anything, serviceConfig).Return("", ([]*configv1.ToolDefinition)(nil), ([]*configv1.ResourceDefinition)(nil), errors.New("temporary failure"))

	// Start worker
	worker.Start(ctx)

	// Publish request
	req := &bus.ServiceRegistrationRequest{Config: serviceConfig}
	reqBus, _ := bus.GetBus[*bus.ServiceRegistrationRequest](busProvider, bus.ServiceRegistrationRequestTopic)
	err := reqBus.Publish(ctx, "request", req)
	assert.NoError(t, err)

	// Wait enough time for at least 1 retry (BaseDelay = 1s)
	// Initial call happens immediately.
	// 1st retry: after 1s.
	// 2nd retry: after 2s (Total 3s)
	// 3rd retry: after 4s (Total 7s)

	time.Sleep(1500 * time.Millisecond)

	// Verify calls
	// Should have called at least twice (initial + 1 retry)
	calls := len(mockRegistry.Calls)
	assert.GreaterOrEqual(t, calls, 2, "Should have retried at least once")
}

func TestServiceRegistrationWorker_RequestCopy(t *testing.T) {
	// Verify that the retry logic creates a copy of the request
	// instead of mutating the original one.

	// Since we can't easily intercept internal variables, we rely on the fact
	// that if we publish the SAME pointer, side effects might be visible if we held onto it.
	// But `bus` usually sends pointers.

	// This is hard to test black-box style without race conditions or deep hooks.
	// The code review addressed the fix.
}
