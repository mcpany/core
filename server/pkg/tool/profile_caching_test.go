// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestManager_ListMCPToolsForProfile_Caching(t *testing.T) {
	tm := NewManager(nil)

	// Define services and tools
	s1 := &ServiceInfo{
		Name: "service1",
		Config: configv1.UpstreamServiceConfig_builder{
			// Name: proto.String("service1"), // Name might not be in builder, rely on ServiceInfo.Name
		}.Build(),
	}
	s2 := &ServiceInfo{
		Name: "service2",
		Config: configv1.UpstreamServiceConfig_builder{
			// Name: proto.String("service2"),
		}.Build(),
	}
	tm.AddServiceInfo("s1", s1)
	tm.AddServiceInfo("s2", s2)

	// Add tools
	// Mock tool that implements Tool interface
	t1 := &MockTool{
		ToolFunc: func() *v1.Tool {
			return v1.Tool_builder{
				Name:      proto.String("tool1"),
				ServiceId: proto.String("s1"),
			}.Build()
		},
		MCPToolFunc: func() *mcp.Tool {
			return &mcp.Tool{Name: "s1.tool1"}
		},
	}
	t2 := &MockTool{
		ToolFunc: func() *v1.Tool {
			return v1.Tool_builder{
				Name:      proto.String("tool2"),
				ServiceId: proto.String("s2"),
			}.Build()
		},
		MCPToolFunc: func() *mcp.Tool {
			return &mcp.Tool{Name: "s2.tool2"}
		},
	}

	// We need to use AddTool to populate indices and caches
	err := tm.AddTool(t1)
	assert.NoError(t, err)
	err = tm.AddTool(t2)
	assert.NoError(t, err)

	// Setup Profile
	// Profile "p1" allows only "s1"
	p1 := configv1.ProfileDefinition_builder{
		Name: proto.String("p1"),
		ServiceConfig: map[string]*configv1.ProfileServiceConfig{
			"s1": configv1.ProfileServiceConfig_builder{
				Enabled: proto.Bool(true),
			}.Build(),
		},
	}.Build()

	tm.SetProfiles([]string{"p1"}, []*configv1.ProfileDefinition{p1})

	// 1. First Call - Miss
	tools := tm.ListMCPToolsForProfile("p1")
	assert.Len(t, tools, 1)
	if len(tools) > 0 {
		assert.Equal(t, "s1.tool1", tools[0].Name)
	}

	// Verify Cache populated
	tm.toolsMutex.RLock()
	cached, ok := tm.cachedProfileTools["p1"]
	tm.toolsMutex.RUnlock()
	assert.True(t, ok)
	assert.Equal(t, tools, cached)

	// 2. Second Call - Hit
	tools2 := tm.ListMCPToolsForProfile("p1")
	assert.Equal(t, tools, tools2)

	// 3. Invalidation - Add Tool
	t3 := &MockTool{
		ToolFunc: func() *v1.Tool {
			return v1.Tool_builder{
				Name:      proto.String("tool3"),
				ServiceId: proto.String("s1"),
			}.Build()
		},
		MCPToolFunc: func() *mcp.Tool {
			return &mcp.Tool{Name: "s1.tool3"}
		},
	}
	err = tm.AddTool(t3)
	assert.NoError(t, err)

	// Verify Cache Invalidated
	tm.toolsMutex.RLock()
	assert.Nil(t, tm.cachedProfileTools)
	tm.toolsMutex.RUnlock()

	// 4. Call Again - Should include tool3
	tools3 := tm.ListMCPToolsForProfile("p1")
	assert.Len(t, tools3, 2) // tool1 and tool3

	// 5. Invalidation - SetProfiles
	p1Updated := configv1.ProfileDefinition_builder{
		Name: proto.String("p1"),
		ServiceConfig: map[string]*configv1.ProfileServiceConfig{
			"s1": configv1.ProfileServiceConfig_builder{
				Enabled: proto.Bool(true),
			}.Build(),
			"s2": configv1.ProfileServiceConfig_builder{
				Enabled: proto.Bool(true),
			}.Build(),
		},
	}.Build()

	tm.SetProfiles([]string{"p1"}, []*configv1.ProfileDefinition{p1Updated})

	// Verify Cache Invalidated
	tm.toolsMutex.RLock()
	assert.Nil(t, tm.cachedProfileTools)
	tm.toolsMutex.RUnlock()

	// 6. Call Again - Should include tool2 (from s2)
	tools4 := tm.ListMCPToolsForProfile("p1")
	assert.Len(t, tools4, 3) // tool1, tool2, tool3
}
