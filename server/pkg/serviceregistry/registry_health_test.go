// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package serviceregistry

import (
	"context"
	"errors"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/upstream"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockHealthCheckerUpstream struct {
	mockUpstream
	checkHealthFunc func(ctx context.Context) error
}

func (m *mockHealthCheckerUpstream) CheckHealth(ctx context.Context) error {
	if m.checkHealthFunc != nil {
		return m.checkHealthFunc(ctx)
	}
	return nil
}

func TestHealthCheck(t *testing.T) {
	healthErr := errors.New("health check failed")
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
					return healthErr
				},
			}, nil
		},
	}
	tm := &mockToolManager{}
	registry := New(f, tm, prompt.NewManager(), resource.NewManager(), auth.NewManager())

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("unhealthy-service")

	// Register service
	serviceID, _, _, err := registry.RegisterService(context.Background(), serviceConfig)
	require.NoError(t, err)

	// Since we added immediate health check in RegisterService, the error should be present immediately
	msg, ok := registry.GetServiceError(serviceID)
	assert.True(t, ok, "Service should have health error")
	assert.Equal(t, healthErr.Error(), msg)

	// Now simulate recovery
	// We need to update the checkHealthFunc of the EXISTING upstream instance.

	registry.mu.RLock()
	u := registry.upstreams[serviceID]
	registry.mu.RUnlock()

	mockU, ok := u.(*mockHealthCheckerUpstream)
	require.True(t, ok)

	// Update health check to succeed
	mockU.checkHealthFunc = func(ctx context.Context) error {
		return nil
	}

	// Manually trigger checkAllHealth (simulating ticker)
	registry.checkAllHealth(context.Background())

	// Verify error is gone
	msg, ok = registry.GetServiceError(serviceID)
	assert.False(t, ok, "Service should be healthy now")
}
