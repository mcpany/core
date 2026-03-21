// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package serviceregistry

import (
	"context"
	"errors"
	"testing"

	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/upstream"
	"github.com/mcpany/core/server/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceRegistry_UnregisterService_ShutdownFailure(t *testing.T) {
	mockUpstream := &mockUpstream{
		registerFunc: func(serviceName string) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
			serviceID, err := util.SanitizeServiceName(serviceName)
			require.NoError(t, err)
			return serviceID, nil, nil, nil
		},
	}
	f := &mockFactory{
		newUpstreamFunc: func() (upstream.Upstream, error) {
			return mockUpstream, nil
		},
	}
	tm := &mockToolManager{}
	prm := prompt.NewManager()
	rm := resource.NewManager()
	am := auth.NewManager()
	registry := New(f, tm, prm, rm, am)

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("broken-shutdown-service")

	// 1. Register the service
	serviceID, _, _, err := registry.RegisterService(context.Background(), serviceConfig)
	require.NoError(t, err)

	// 2. Configure mock to fail shutdown
	mockUpstream.shutdownFunc = func() error {
		return errors.New("cannot stop me")
	}

	// 3. Attempt to Unregister
	err = registry.UnregisterService(context.Background(), serviceID)
	// We expect an error, but the service SHOULD be removed
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot stop me")

	// 4. Verify that the service is REMOVED (Fix verification)
	_, ok := registry.GetServiceConfig(serviceID)
	assert.False(t, ok, "Service should be removed even if shutdown failed")

	// 5. Attempt to Register again (simulating reload/update)
	// This should now SUCCEED
	_, _, _, err = registry.RegisterService(context.Background(), serviceConfig)
	require.NoError(t, err)
}
