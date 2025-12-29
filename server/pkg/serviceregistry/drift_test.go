// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package serviceregistry

import (
	"context"
	"errors"
	"testing"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/upstream"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceRegistry_RegisterService_PersistConfigOnFailure(t *testing.T) {
	upstreamErr := errors.New("upstream connection error")
	f := &mockFactory{
		newUpstreamFunc: func() (upstream.Upstream, error) {
			return &mockUpstream{
				registerFunc: func(_ string) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
					return "", nil, nil, upstreamErr
				},
			}, nil
		},
	}
	tm := &mockToolManager{}
	registry := New(f, tm, prompt.NewManager(), resource.NewManager(), auth.NewManager())

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("broken-service")

	// Try to register a broken service
	_, _, _, err := registry.RegisterService(context.Background(), serviceConfig)
	require.Error(t, err)
	assert.Contains(t, err.Error(), upstreamErr.Error())

	// Verify that the service config IS still in the registry
	serviceID, _ := util.SanitizeServiceName("broken-service")
	_, ok := registry.GetServiceConfig(serviceID)
	assert.True(t, ok, "Service config should be persisted even after registration failure")

	// Verify that the error is stored
	storedErr, ok := registry.GetServiceError(serviceID)
	assert.True(t, ok, "Service error should be stored")
	assert.Equal(t, upstreamErr.Error(), storedErr)

	// Verify that GetAllServices returns it
	allServices, err := registry.GetAllServices()
	require.NoError(t, err)
	assert.Len(t, allServices, 1, "GetAllServices should return the broken service")
	assert.Equal(t, "broken-service", allServices[0].GetName())
}
