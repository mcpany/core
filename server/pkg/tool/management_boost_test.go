// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	mcp_router_v1 "github.com/mcpany/core/proto/mcp_router/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestManager_ListMCPTools(t *testing.T) {
	m := NewManager(nil)

    svcConfig := configv1.UpstreamServiceConfig_builder{
        Id: proto.String("svc1"),
    }.Build()
    m.AddServiceInfo("svc1", &ServiceInfo{
        Name: "svc1",
        Config: svcConfig,
    })

    toolDef := mcp_router_v1.Tool_builder{
        Name: proto.String("tool1"),
        ServiceId: proto.String("svc1"),
    }.Build()
    tool := &MockTool{
        ToolFunc: func() *v1.Tool { return toolDef },
        MCPToolFunc: func() *mcp.Tool { return &mcp.Tool{Name: "svc1.tool1"} },
    }
    err := m.AddTool(tool)
    assert.NoError(t, err)

    // Add another tool
    toolDef2 := mcp_router_v1.Tool_builder{
        Name: proto.String("tool2"),
        ServiceId: proto.String("svc1"),
    }.Build()
    tool2 := &MockTool{
        ToolFunc: func() *v1.Tool { return toolDef2 },
        MCPToolFunc: func() *mcp.Tool { return &mcp.Tool{Name: "svc1.tool2"} },
    }
    err = m.AddTool(tool2)
    assert.NoError(t, err)

    // List tools
    tools := m.ListMCPTools()
    assert.Len(t, tools, 2)
}

func TestManager_ProfileMatching(t *testing.T) {
	m := NewManager(nil)

	svcConfig := configv1.UpstreamServiceConfig_builder{Id: proto.String("svc1")}.Build()
	m.AddServiceInfo("svc1", &ServiceInfo{Name: "svc1", Config: svcConfig})

	toolDef := v1.Tool_builder{
		Name:      proto.String("tool1"),
		ServiceId: proto.String("svc1"),
		Tags:      []string{"tag1"},
	}.Build()
	tool := &MockTool{ToolFunc: func() *v1.Tool { return toolDef }}
    err := m.AddTool(tool)
    assert.NoError(t, err)

	// Define profile matching tag1
	profile := configv1.ProfileDefinition_builder{
		Name: proto.String("p1"),
		ServiceConfig: map[string]*configv1.ProfileServiceConfig{
			"svc1": configv1.ProfileServiceConfig_builder{
				Enabled: proto.Bool(true),
			}.Build(),
		},
		Selector: configv1.ProfileSelector_builder{
			Tags: []string{"tag1"},
		}.Build(),
	}.Build()
	m.SetProfiles([]string{"p1"}, []*configv1.ProfileDefinition{profile})

	// Should match
	assert.True(t, m.ToolMatchesProfile(tool, "p1"))

	// Should not match non-existent profile
	assert.False(t, m.ToolMatchesProfile(tool, "p2"))

    // Test mismatched tag
	toolDef2 := v1.Tool_builder{
		Name:      proto.String("tool2"),
		ServiceId: proto.String("svc1"),
		Tags:      []string{"tag2"},
	}.Build()
	tool2 := &MockTool{ToolFunc: func() *v1.Tool { return toolDef2 }}
    // Don't add to manager, just check logic if possible?
    // ToolMatchesProfile calls GetServiceInfo via manager.
    // So we can pass tool2.
    assert.False(t, m.ToolMatchesProfile(tool2, "p1"))
}

func TestWebrtcTool_Coverage(t *testing.T) {
    toolDef := v1.Tool_builder{Name: proto.String("webrtc-tool")}.Build()
    // CallDefinition is required or handled gracefully?
    // We pass nil for callDef, which might panic in NewWebrtcTool if it accesses fields.
    // NewWebrtcTool accesses callDefinition.GetParameters() etc.
    callDef := configv1.WebrtcCallDefinition_builder{}.Build()
    wt, err := NewWebrtcTool(toolDef, nil, "svc1", nil, callDef)
    assert.NoError(t, err)
    assert.NotNil(t, wt)
    assert.Equal(t, toolDef, wt.Tool())

    // Execute without pool (should fail because newPeerConnection fails or connection fails)
    // Actually newPeerConnection uses STUN.
    // We can disable STUN via env.
    t.Setenv("MCPANY_WEBRTC_DISABLE_STUN", "true")

    req := &ExecutionRequest{ToolName: "webrtc-tool", ToolInputs: []byte("{}")}
    _, err = wt.Execute(context.Background(), req)
    // Execute will try to create connection, then create offer, then send HTTP POST.
    // Address is "webrtc-tool" (from UnderlyingMethodFqn if not set? No, it expects WEBRTC prefix).
    // It will likely fail at HTTP POST or earlier.
    assert.Error(t, err)

    // IsHealthy is not method of WebrtcTool, but peerConnectionWrapper.
    // WebrtcTool doesn't expose IsHealthy directly.
}
