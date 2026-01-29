package serviceregistry

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	mcp_routerv1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/upstream"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestServiceRegistry_ConcurrentAccess(t *testing.T) {
	f := &mockFactory{
		newUpstreamFunc: func() (upstream.Upstream, error) {
			return &mockUpstream{
				registerFunc: func(serviceName string) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
					serviceID, err := util.SanitizeServiceName(serviceName)
					if err != nil {
						return "", nil, nil, err
					}
					// Simulate some work
					time.Sleep(10 * time.Millisecond)
					return serviceID, nil, nil, nil
				},
			}, nil
		},
	}
	tm := newThreadSafeToolManager()
	registry := New(f, tm, prompt.NewManager(), resource.NewManager(), auth.NewManager())

	var wg sync.WaitGroup
	numRoutines := 50

	// Launch routines to register services
	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			serviceName := fmt.Sprintf("service-%d", i)
			config := configv1.UpstreamServiceConfig_builder{
				Name: proto.String(serviceName),
			}.Build()
			_, _, _, _ = registry.RegisterService(context.Background(), config)
		}(i)
	}

	// Launch routines to unregister services
	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// Wait a bit to let some registrations happen
			time.Sleep(5 * time.Millisecond)
			serviceName := fmt.Sprintf("service-%d", i)
			_ = registry.UnregisterService(context.Background(), serviceName)
		}(i)
	}

	// Launch routines to read services
	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = registry.GetAllServices()
		}()
	}

	wg.Wait()

	// Ensure final state is consistent (no panic is the main test here)
	_, err := registry.GetAllServices()
	require.NoError(t, err)
}

func TestServiceRegistry_RegisterService_Interrupted(t *testing.T) {
	// This test verifies that if UnregisterService is called while RegisterService is in progress (specifically during u.Register),
	// the service is correctly cleaned up.

	registrationBlock := make(chan struct{})
	registrationResume := make(chan struct{})

	f := &mockFactory{
		newUpstreamFunc: func() (upstream.Upstream, error) {
			return &mockUpstream{
				registerFunc: func(serviceName string) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
					serviceID, err := util.SanitizeServiceName(serviceName)
					if err != nil {
						return "", nil, nil, err
					}
					// Signal we are in Register
					close(registrationBlock)
					// Wait for resume signal
					<-registrationResume
					return serviceID, nil, nil, nil
				},
				shutdownFunc: func() error {
					return nil
				},
			}, nil
		},
	}
	tm := newThreadSafeToolManager()
	registry := New(f, tm, prompt.NewManager(), resource.NewManager(), auth.NewManager())

	serviceName := "interrupted-service"
	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String(serviceName),
	}.Build()

	errChan := make(chan error)
	go func() {
		_, _, _, err := registry.RegisterService(context.Background(), serviceConfig)
		errChan <- err
	}()

	// Wait until we are inside u.Register
	select {
	case <-registrationBlock:
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for registration to start")
	}

	// Now unregister the service
	// Note: RegisterService holds the lock initially, creates upstream, adds to upstreams map, RELEASES lock, calls u.Register.
	// So UnregisterService CAN proceed.
	// UnregisterService will see the service in serviceConfigs and upstreams (because it was added before releasing lock).
	// It will call u.Shutdown and remove it from maps.
	err := registry.UnregisterService(context.Background(), serviceName)
	require.NoError(t, err)

	// Allow Register to complete
	close(registrationResume)

	// Wait for RegisterService to return
	regErr := <-errChan

	// RegisterService checks if the service is still in serviceConfigs after u.Register returns.
	// Since we unregistered it, it should not be there.
	// It should verify that cleanup was performed and return an error or nil depending on implementation.
	// The current implementation returns an error: fmt.Errorf("service %q was unregistered during registration", serviceConfig.GetName())
	require.Error(t, regErr)
	assert.Contains(t, regErr.Error(), "unregistered during registration")

	// Verify cleanup
	_, ok := registry.serviceConfigs[serviceName] // accessing private field for verification, in test package it's allowed
	assert.False(t, ok)
	_, ok = registry.upstreams[serviceName]
	assert.False(t, ok)
}

func TestServiceRegistry_HealthCheckLoop(t *testing.T) {
	healthErr := errors.New("simulated health error")
	checkChan := make(chan struct{}, 1) // buffered to avoid blocking

	f := &mockFactory{
		newUpstreamFunc: func() (upstream.Upstream, error) {
			return &mockHealthCheckerUpstream{
				mockUpstream: mockUpstream{
					registerFunc: func(serviceName string) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
						serviceID, err := util.SanitizeServiceName(serviceName)
						require.NoError(t, err)
						return serviceID, nil, nil, nil
					},
				},
				checkHealthFunc: func(ctx context.Context) error {
					select {
					case checkChan <- struct{}{}:
					default:
					}
					return healthErr
				},
			}, nil
		},
	}
	tm := &mockToolManager{}
	registry := New(f, tm, prompt.NewManager(), resource.NewManager(), auth.NewManager())

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("health-service")

	// Register service
	serviceID, _, _, err := registry.RegisterService(context.Background(), serviceConfig)
	require.NoError(t, err)

	// Start health checks with short interval
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	registry.StartHealthChecks(ctx, 10*time.Millisecond)

	// Wait for health check to run
	select {
	case <-checkChan:
	case <-time.After(1 * time.Second):
		t.Fatal("health check did not run")
	}

	// Verify error is set
	// Note: It might take a moment for the goroutine to update the map after checkHealthFunc returns
	assert.Eventually(t, func() bool {
		msg, ok := registry.GetServiceError(serviceID)
		return ok && msg == healthErr.Error()
	}, 1*time.Second, 10*time.Millisecond)
}

func TestServiceRegistry_RuntimeInfo(t *testing.T) {
	f := &mockFactory{
		newUpstreamFunc: func() (upstream.Upstream, error) {
			return &mockUpstream{
				registerFunc: func(serviceName string) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
					serviceID, err := util.SanitizeServiceName(serviceName)
					require.NoError(t, err)
					return serviceID, nil, nil, nil
				},
			}, nil
		},
	}
	tm := newThreadSafeToolManager()
	registry := New(f, tm, prompt.NewManager(), resource.NewManager(), auth.NewManager())

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("info-service"),
	}.Build()
	serviceID, _, _, err := registry.RegisterService(context.Background(), serviceConfig)
	require.NoError(t, err)

	// 1. Verify Tool Count
	// Add 3 tools for this service
	tm.AddTool(&mockTool{tool: mcp_routerv1.Tool_builder{Name: proto.String("tool1"), ServiceId: proto.String(serviceID)}.Build()})
	tm.AddTool(&mockTool{tool: mcp_routerv1.Tool_builder{Name: proto.String("tool2"), ServiceId: proto.String(serviceID)}.Build()})
	tm.AddTool(&mockTool{tool: mcp_routerv1.Tool_builder{Name: proto.String("tool3"), ServiceId: proto.String(serviceID)}.Build()})

	// Add tool for another service
	tm.AddTool(&mockTool{tool: mcp_routerv1.Tool_builder{Name: proto.String("other_tool"), ServiceId: proto.String("other_service")}.Build()})

	config, ok := registry.GetServiceConfig(serviceID)
	require.True(t, ok)
	assert.Equal(t, int32(3), config.GetToolCount())

	// 2. Verify LastError Priority
	// Set both registration error and health error manually
	registry.mu.Lock()
	registry.serviceErrors[serviceID] = "registration error"
	registry.healthErrors[serviceID] = "health error"
	registry.mu.Unlock()

	config, ok = registry.GetServiceConfig(serviceID)
	require.True(t, ok)
	assert.Equal(t, "registration error", config.GetLastError(), "Should prioritize registration error")

	// Remove registration error
	registry.mu.Lock()
	delete(registry.serviceErrors, serviceID)
	registry.mu.Unlock()

	config, ok = registry.GetServiceConfig(serviceID)
	require.True(t, ok)
	assert.Equal(t, "health error", config.GetLastError(), "Should show health error if no registration error")
}
