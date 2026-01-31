// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package serviceregistry

import (
	"context"
	"testing"

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

// TestToolCountOptimization tests that CountToolsForService returns the correct count
// and is used by GetAllServices/GetServiceConfig.
func TestToolCountOptimization(t *testing.T) {
	// Setup Mocks
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
	// Use the threadSafeToolManager which we updated to support CountToolsForService
	tm := newThreadSafeToolManager()
	registry := New(f, tm, prompt.NewManager(), resource.NewManager(), auth.NewManager())

	// Register a service
	serviceName := "perf-test-service"
	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String(serviceName),
	}.Build()

	serviceID, _, _, err := registry.RegisterService(context.Background(), serviceConfig)
	require.NoError(t, err)

	// Initially count should be 0
	config, ok := registry.GetServiceConfig(serviceID)
	require.True(t, ok)
	assert.Equal(t, int32(0), config.GetToolCount())

	// Add 5 tools
	for i := 0; i < 5; i++ {
		toolName := proto.String(util.BytesToString([]byte{byte('a' + i)})) // "a", "b", ...
		tool := &mockTool{
			tool: mcp_routerv1.Tool_builder{
				Name:      toolName,
				ServiceId: proto.String(serviceID),
			}.Build(),
		}
		err := tm.AddTool(tool)
		require.NoError(t, err)
	}

	// Verify count is 5 via GetServiceConfig
	config, ok = registry.GetServiceConfig(serviceID)
	require.True(t, ok)
	assert.Equal(t, int32(5), config.GetToolCount())

	// Verify count is 5 via GetAllServices
	services, err := registry.GetAllServices()
	require.NoError(t, err)
	assert.Len(t, services, 1)
	assert.Equal(t, int32(5), services[0].GetToolCount())

	// Add another service with 3 tools to ensure isolation
	serviceName2 := "perf-test-service-2"
	serviceConfig2 := configv1.UpstreamServiceConfig_builder{
		Name: proto.String(serviceName2),
	}.Build()
	serviceID2, _, _, err := registry.RegisterService(context.Background(), serviceConfig2)
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		toolName := proto.String(util.BytesToString([]byte{byte('x' + i)})) // "x", "y", ...
		tool := &mockTool{
			tool: mcp_routerv1.Tool_builder{
				Name:      toolName,
				ServiceId: proto.String(serviceID2),
			}.Build(),
		}
		err := tm.AddTool(tool)
		require.NoError(t, err)
	}

	// Check first service still has 5
	config, ok = registry.GetServiceConfig(serviceID)
	require.True(t, ok)
	assert.Equal(t, int32(5), config.GetToolCount())

	// Check second service has 3
	config2, ok := registry.GetServiceConfig(serviceID2)
	require.True(t, ok)
	assert.Equal(t, int32(3), config2.GetToolCount())
}
