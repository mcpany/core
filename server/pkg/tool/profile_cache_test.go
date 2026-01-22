// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"

	"github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	buspkg "github.com/mcpany/core/server/pkg/bus"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListMCPToolsForProfile_Caching(t *testing.T) {
	// Setup
	messageBus := bus.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus.InMemoryBus_builder{}.Build())
	busProvider, err := buspkg.NewProvider(messageBus)
	require.NoError(t, err)

	tm := NewManager(busProvider)

	// Add Service Info
	tm.AddServiceInfo("service-a", &ServiceInfo{
		Config: &configv1.UpstreamServiceConfig{
			Name: strPtr("service-a"),
		},
	})
	tm.AddServiceInfo("service-b", &ServiceInfo{
		Config: &configv1.UpstreamServiceConfig{
			Name: strPtr("service-b"),
		},
	})

	// Add Tools
	toolA := &MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{
				Name:      strPtr("tool-a"),
				ServiceId: strPtr("service-a"),
			}
		},
		MCPToolFunc: func() *mcp.Tool {
			return &mcp.Tool{
				Name: "tool-a",
			}
		},
	}
	toolB := &MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{
				Name:      strPtr("tool-b"),
				ServiceId: strPtr("service-b"),
			}
		},
		MCPToolFunc: func() *mcp.Tool {
			return &mcp.Tool{
				Name: "tool-b",
			}
		},
	}
	require.NoError(t, tm.AddTool(toolA))
	require.NoError(t, tm.AddTool(toolB))

	// Setup Profiles
	profiles := []*configv1.ProfileDefinition{
		{
			Name: strPtr("profile-1"),
			ServiceConfig: map[string]*configv1.ProfileServiceConfig{
				"service-a": {Enabled: boolPtr(true)},
			},
		},
		{
			Name: strPtr("profile-2"),
			ServiceConfig: map[string]*configv1.ProfileServiceConfig{
				"service-b": {Enabled: boolPtr(true)},
			},
		},
	}
	tm.SetProfiles([]string{"profile-1", "profile-2"}, profiles)

	// Verify setup
	tools := tm.ListMCPTools()
	assert.Len(t, tools, 2)

	// Verify caching logic
	// 1st call - should compute and cache
	p1Tools := tm.ListMCPToolsForProfile("profile-1")
	assert.Len(t, p1Tools, 1)
	assert.Equal(t, "service-a.tool-a", p1Tools[0].Name)

	// 2nd call - should return from cache
	p1ToolsCached := tm.ListMCPToolsForProfile("profile-1")
	assert.Len(t, p1ToolsCached, 1)
	assert.Equal(t, p1Tools, p1ToolsCached) // Should be same slice (pointer equality if we checked pointers, but equal content is fine)

	p2Tools := tm.ListMCPToolsForProfile("profile-2")
	assert.Len(t, p2Tools, 1)
	assert.Equal(t, "service-b.tool-b", p2Tools[0].Name)

	// Invalid profile - should return empty
	invalidTools := tm.ListMCPToolsForProfile("invalid-profile")
	assert.Empty(t, invalidTools)
}
