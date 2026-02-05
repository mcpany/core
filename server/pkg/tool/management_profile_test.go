// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestManager_Profiles(t *testing.T) {
	// Setup
	tm := NewManager(nil)

	// Define profiles
	p1 := configv1.ProfileDefinition_builder{
		Name: proto.String("p1"),
		ServiceConfig: map[string]*configv1.ProfileServiceConfig{
			"s1": configv1.ProfileServiceConfig_builder{Enabled: proto.Bool(true)}.Build(),
			"s2": configv1.ProfileServiceConfig_builder{Enabled: proto.Bool(false)}.Build(),
		},
	}.Build()
	p2 := configv1.ProfileDefinition_builder{
		Name: proto.String("p2"),
		Selector: configv1.ProfileSelector_builder{
			Tags: []string{"allowed"},
		}.Build(),
	}.Build()

	tm.SetProfiles([]string{"p1", "p2"}, []*configv1.ProfileDefinition{p1, p2})

	// Test IsServiceAllowed
	assert.True(t, tm.IsServiceAllowed("s1", "p1"))
	assert.False(t, tm.IsServiceAllowed("s2", "p1"))
	assert.False(t, tm.IsServiceAllowed("s3", "p1")) // Not in config
	assert.False(t, tm.IsServiceAllowed("s1", "unknown"))

	// Test GetAllowedServiceIDs
	allowed, ok := tm.GetAllowedServiceIDs("p1")
	assert.True(t, ok)
	assert.True(t, allowed["s1"])
	assert.False(t, allowed["s2"])
	assert.Len(t, allowed, 1)

	allowed, ok = tm.GetAllowedServiceIDs("unknown")
	assert.False(t, ok)
	assert.Nil(t, allowed)

	// Test ToolMatchesProfile
	tool1 := v1.Tool_builder{
		ServiceId: proto.String("s1"),
		Tags: []string{"tag1"},
	}.Build()
	// Mock tool wrapper
	mTool1 := &MockTool{ToolFunc: func() *v1.Tool { return tool1 }}

	// p1 allows s1
	assert.True(t, tm.ToolMatchesProfile(mTool1, "p1"))

	tool2 := v1.Tool_builder{
		ServiceId: proto.String("s3"),
		Tags: []string{"allowed"},
	}.Build()
	mTool2 := &MockTool{ToolFunc: func() *v1.Tool { return tool2 }}

	// p2 selects by tag "allowed"
	assert.True(t, tm.ToolMatchesProfile(mTool2, "p2"))

	tool3 := v1.Tool_builder{
		ServiceId: proto.String("s3"),
		Tags: []string{"forbidden"},
	}.Build()
	mTool3 := &MockTool{ToolFunc: func() *v1.Tool { return tool3 }}

	// p2 does not select "forbidden"
	assert.False(t, tm.ToolMatchesProfile(mTool3, "p2"))

    // Test matchesProperties
    p3 := configv1.ProfileDefinition_builder{
		Name: proto.String("p3"),
		Selector: configv1.ProfileSelector_builder{
			ToolProperties: map[string]string{
                "read_only": "true",
            },
		}.Build(),
	}.Build()
    // Update profiles
    tm.SetProfiles([]string{"p1", "p2", "p3"}, []*configv1.ProfileDefinition{p1, p2, p3})

    tool4 := v1.Tool_builder{
        Annotations: v1.ToolAnnotations_builder{
            ReadOnlyHint: proto.Bool(true),
        }.Build(),
    }.Build()
    mTool4 := &MockTool{ToolFunc: func() *v1.Tool { return tool4 }}
    assert.True(t, tm.ToolMatchesProfile(mTool4, "p3"))

    tool5 := v1.Tool_builder{
        Annotations: v1.ToolAnnotations_builder{
            ReadOnlyHint: proto.Bool(false),
        }.Build(),
    }.Build()
    mTool5 := &MockTool{ToolFunc: func() *v1.Tool { return tool5 }}
    assert.False(t, tm.ToolMatchesProfile(mTool5, "p3"))
}

// Add test for isToolAllowed (internal method)
func TestManager_IsToolAllowed_Indirect(t *testing.T) {
     tm := NewManager(nil)

     p1 := configv1.ProfileDefinition_builder{
		Name: proto.String("p1"),
        ServiceConfig: map[string]*configv1.ProfileServiceConfig{
			"s1": configv1.ProfileServiceConfig_builder{Enabled: proto.Bool(true)}.Build(),
		},
	}.Build()
    tm.SetProfiles([]string{"p1"}, []*configv1.ProfileDefinition{p1})

    // Tool allowed
    t1 := v1.Tool_builder{ServiceId: proto.String("s1"), Name: proto.String("t1")}.Build()
    _ = &MockTool{ToolFunc: func() *v1.Tool { return t1 }}

    assert.True(t, tm.isToolAllowed(t1))

    t2 := v1.Tool_builder{ServiceId: proto.String("s2"), Name: proto.String("t2")}.Build()
    assert.False(t, tm.isToolAllowed(t2))
}

func TestManager_GetTool_Inconsistent(t *testing.T) {
	tm := NewManager(nil)
    // Manually corrupt the map (using implementation details)
    tm.nameMap.Store("alias", "realID")
    // But "realID" is NOT in tm.tools

    tool, ok := tm.GetTool("alias")
    assert.False(t, ok)
    assert.Nil(t, tool)
}

func TestIsSensitiveHeader(t *testing.T) {
    assert.True(t, isSensitiveHeader("Authorization"))
    assert.True(t, isSensitiveHeader("X-My-Token"))
    assert.False(t, isSensitiveHeader("Content-Type"))
}

func TestManager_Profile_GranularToolDisabling(t *testing.T) {
	tm := NewManager(nil)

	p1 := configv1.ProfileDefinition_builder{
		Name: proto.String("p1"),
		ServiceConfig: map[string]*configv1.ProfileServiceConfig{
			"s1": configv1.ProfileServiceConfig_builder{
				Enabled: proto.Bool(true),
				Tools: map[string]*configv1.ToolConfig{
					"disabled_tool": configv1.ToolConfig_builder{Disabled: true}.Build(),
					"enabled_tool":  configv1.ToolConfig_builder{Disabled: false}.Build(),
				},
			}.Build(),
		},
	}.Build()

	tm.SetProfiles([]string{"p1"}, []*configv1.ProfileDefinition{p1})

	// Enabled Service, Disabled Tool
	t1 := v1.Tool_builder{ServiceId: proto.String("s1"), Name: proto.String("disabled_tool")}.Build()
	mTool1 := &MockTool{ToolFunc: func() *v1.Tool { return t1 }}
	assert.False(t, tm.ToolMatchesProfile(mTool1, "p1"), "Tool should be disabled by config")

	// Enabled Service, Explicitly Enabled Tool
	t2 := v1.Tool_builder{ServiceId: proto.String("s1"), Name: proto.String("enabled_tool")}.Build()
	mTool2 := &MockTool{ToolFunc: func() *v1.Tool { return t2 }}
	assert.True(t, tm.ToolMatchesProfile(mTool2, "p1"), "Tool should be enabled by config")

	// Enabled Service, Unspecified Tool (Default allowed)
	t3 := v1.Tool_builder{ServiceId: proto.String("s1"), Name: proto.String("other_tool")}.Build()
	mTool3 := &MockTool{ToolFunc: func() *v1.Tool { return t3 }}
	assert.True(t, tm.ToolMatchesProfile(mTool3, "p1"), "Tool should be allowed by default service enablement")
}

func TestManager_Profile_TagBasedServiceAccess(t *testing.T) {
	tm := NewManager(nil)
	pTag := configv1.ProfileDefinition_builder{
		Name: proto.String("pTag"),
		Selector: configv1.ProfileSelector_builder{
			Tags: []string{"finance"},
		}.Build(),
	}.Build()
	tm.SetProfiles([]string{"pTag"}, []*configv1.ProfileDefinition{pTag})

	// Register service with tag
	svc := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("finance-service"),
		Tags: []string{"finance"},
		McpService: configv1.McpUpstreamService_builder{}.Build(), // minimal
	}.Build()
	tm.AddServiceInfo("finance-service", &ServiceInfo{Config: svc})

	// Assert IsServiceAllowed
	assert.True(t, tm.IsServiceAllowed("finance-service", "pTag"))

	// Assert GetAllowedServiceIDs
	allowed, ok := tm.GetAllowedServiceIDs("pTag")
	assert.True(t, ok)
	assert.True(t, allowed["finance-service"])
}
