// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package serviceregistry

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	mcp_routerv1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

type mockHealthCheckedUpstream struct {
	mockUpstream
	mu              sync.Mutex
	checkHealthFunc func(context.Context) error
}

func (m *mockHealthCheckedUpstream) CheckHealth(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.checkHealthFunc != nil {
		return m.checkHealthFunc(ctx)
	}
	return nil
}

func (m *mockHealthCheckedUpstream) setCheckHealthFunc(f func(context.Context) error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.checkHealthFunc = f
}

func TestStartHealthChecks(t *testing.T) {
	healthErr := errors.New("health check failed")
	mu := &mockHealthCheckedUpstream{
		mockUpstream: mockUpstream{
			registerFunc: func(serviceName string) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
				s, err := util.SanitizeServiceName(serviceName)
				return s, nil, nil, err
			},
		},
		checkHealthFunc: func(ctx context.Context) error {
			return healthErr
		},
	}

	f := &mockFactory{
		newUpstreamFunc: func() (upstream.Upstream, error) {
			return mu, nil
		},
	}
	tm := newThreadSafeToolManager()
	registry := New(f, tm, prompt.NewManager(), resource.NewManager(), auth.NewManager())

	// Register service
	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("health-service"),
	}.Build()
	serviceID, _, _, err := registry.RegisterService(context.Background(), serviceConfig)
	require.NoError(t, err)

	// Verify initial health check (RegisterService performs one)
	errStr, ok := registry.GetServiceError(serviceID)
	assert.True(t, ok)
	assert.Equal(t, healthErr.Error(), errStr)

	// Change mock behavior to succeed
	mu.setCheckHealthFunc(func(ctx context.Context) error {
		return nil
	})

	// Start health checks
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	registry.StartHealthChecks(ctx, 100*time.Millisecond)

	// Wait for health check to pass
	assert.Eventually(t, func() bool {
		_, hasError := registry.GetServiceError(serviceID)
		return !hasError
	}, 2*time.Second, 50*time.Millisecond)

	// Make it fail again
	mu.setCheckHealthFunc(func(ctx context.Context) error {
		return errors.New("failed again")
	})

	// Wait for health check to fail
	assert.Eventually(t, func() bool {
		msg, hasError := registry.GetServiceError(serviceID)
		return hasError && msg == "failed again"
	}, 2*time.Second, 50*time.Millisecond)
}

func TestGetServiceInfo_RuntimeInjection(t *testing.T) {
	f := &mockFactory{
		newUpstreamFunc: func() (upstream.Upstream, error) {
			return &mockUpstream{
				registerFunc: func(serviceName string) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
					s, err := util.SanitizeServiceName(serviceName)
					return s, nil, nil, err
				},
			}, nil
		},
	}
	tm := newThreadSafeToolManager()
	registry := New(f, tm, prompt.NewManager(), resource.NewManager(), auth.NewManager())

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("runtime-service"),
	}.Build()
	serviceID, _, _, err := registry.RegisterService(context.Background(), serviceConfig)
	require.NoError(t, err)

	// Initially no error and 0 tools
	config, ok := registry.GetServiceConfig(serviceID)
	require.True(t, ok)
	assert.Equal(t, int32(0), config.GetToolCount())
	assert.Empty(t, config.GetLastError())

	// Add a tool
	tm.AddTool(&mockTool{tool: mcp_routerv1.Tool_builder{
		Name:      proto.String("my-tool"),
		ServiceId: proto.String(serviceID),
	}.Build()})

	// Manually add service info to registry because mockUpstream doesn't do it
	registry.AddServiceInfo(serviceID, &tool.ServiceInfo{
		Name:   "runtime-service",
		Config: serviceConfig,
	})

	// Inject an error (simulate failure)
	registry.serviceErrors[serviceID] = "something went wrong"

	// Check updated info
	config, ok = registry.GetServiceConfig(serviceID)
	require.True(t, ok)
	assert.Equal(t, int32(1), config.GetToolCount())
	assert.Equal(t, "something went wrong", config.GetLastError())

	// Check GetServiceInfo as well
	info, ok := registry.GetServiceInfo(serviceID)
	require.True(t, ok)
	assert.Equal(t, int32(1), info.Config.GetToolCount())
	assert.Equal(t, "something went wrong", info.Config.GetLastError())
}

func TestRegisterService_ConcurrentUnregister(t *testing.T) {
	// Channel to control when Register proceeds
	proceedReg := make(chan struct{})
	// Channel to signal that Register is blocked
	regBlocked := make(chan struct{})

	blockingUpstream := &mockUpstream{
		registerFunc: func(serviceName string) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
			close(regBlocked) // Signal we are in Register
			<-proceedReg      // Wait for signal to proceed
			s, err := util.SanitizeServiceName(serviceName)
			return s, nil, nil, err
		},
	}

	f := &mockFactory{
		newUpstreamFunc: func() (upstream.Upstream, error) {
			return blockingUpstream, nil
		},
	}
	tm := newThreadSafeToolManager()
	registry := New(f, tm, prompt.NewManager(), resource.NewManager(), auth.NewManager())

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("concurrent-service"),
	}.Build()

	// Start RegisterService in a goroutine
	errChan := make(chan error)
	go func() {
		_, _, _, err := registry.RegisterService(context.Background(), serviceConfig)
		errChan <- err
	}()

	// Wait until RegisterService reaches the blocking point
	<-regBlocked

	// At this point, the service should be in serviceConfigs and upstreams (but Register not finished)
	serviceID, _ := util.SanitizeServiceName("concurrent-service")
	_, exists := registry.GetServiceConfig(serviceID)
	assert.True(t, exists, "Service config should exist during registration")

	// Call UnregisterService
	err := registry.UnregisterService(context.Background(), "concurrent-service")
	require.NoError(t, err)

	// Now verify service is gone
	_, exists = registry.GetServiceConfig(serviceID)
	assert.False(t, exists, "Service config should be gone after unregister")

	// Allow RegisterService to proceed
	close(proceedReg)

	// Check result of RegisterService
	regErr := <-errChan
	require.Error(t, regErr)
	assert.Contains(t, regErr.Error(), "was unregistered during registration")
}
