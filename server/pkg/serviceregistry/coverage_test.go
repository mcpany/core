// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package serviceregistry

import (
	"context"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartHealthChecks(t *testing.T) {
	healthCheckCalled := make(chan struct{}, 1)
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
					case healthCheckCalled <- struct{}{}:
					default:
					}
					return nil
				},
			}, nil
		},
	}
	tm := &mockToolManager{}
	registry := New(f, tm, prompt.NewManager(), resource.NewManager(), auth.NewManager())

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("health-service")

	_, _, _, err := registry.RegisterService(context.Background(), serviceConfig)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start health checks with a very short interval
	registry.StartHealthChecks(ctx, 10*time.Millisecond)

	select {
	case <-healthCheckCalled:
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("Health check was not called within timeout")
	}
}

func TestGetServiceInfo_WithConfig(t *testing.T) {
	registry := New(nil, &mockToolManager{}, nil, nil, nil)

	serviceID, _ := util.SanitizeServiceName("test-service")
	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("test-service")

	serviceInfo := &tool.ServiceInfo{
		Name:   "test-service",
		Config: serviceConfig,
	}
	registry.AddServiceInfo(serviceID, serviceInfo)

	// Inject an error to verify injectRuntimeInfo is called
	registry.mu.Lock()
	registry.serviceErrors[serviceID] = "some error"
	registry.mu.Unlock()

	retrievedInfo, ok := registry.GetServiceInfo(serviceID)
	require.True(t, ok)
	assert.Equal(t, serviceInfo.Name, retrievedInfo.Name)
	assert.NotNil(t, retrievedInfo.Config)
	// Verify that the config was cloned (different pointer)
	assert.True(t, retrievedInfo.Config != serviceConfig)
	// Verify runtime info injection
	assert.Equal(t, "some error", retrievedInfo.Config.GetLastError())
}

func TestInjectRuntimeInfo_Fallback(t *testing.T) {
	registry := New(nil, &mockToolManager{}, nil, nil, nil)

	// Case 1: Config is nil
	registry.injectRuntimeInfo(nil)
	// No panic implies success

	// Case 2: Config has no sanitized name (should trigger fallback)
	serviceName := "test service" // Needs sanitization
	config := &configv1.UpstreamServiceConfig{}
	config.SetName(serviceName)

	// Do NOT set SanitizedName on config (it's usually set by RegisterService/util)

	sanitizedName, _ := util.SanitizeServiceName(serviceName)
	registry.mu.Lock()
	registry.serviceErrors[sanitizedName] = "fallback error"
	registry.mu.Unlock()

	registry.mu.Lock()
	registry.injectRuntimeInfo(config)
	registry.mu.Unlock()

	assert.Equal(t, "fallback error", config.GetLastError())
}
