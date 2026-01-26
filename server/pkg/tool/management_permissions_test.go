// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestIsToolAllowed(t *testing.T) {
	tm := NewManager(nil)

	// Define Tools using builders if opaque, or struct if not.
    // Assuming v1.Tool is NOT opaque because previous errors complained about *string mismatch, not unknown field.
    // But let's check configv1.ProfileDefinition which DEFINITELY complained about unknown field.

    // Tools
	toolA := &v1.Tool{Name: proto.String("tool_a"), ServiceId: proto.String("service_1")}
	toolB := &v1.Tool{Name: proto.String("tool_b"), ServiceId: proto.String("service_1")}
	toolC := &v1.Tool{Name: proto.String("tool_c"), ServiceId: proto.String("service_2")}

	// Register tools (using MockTool)
	tA := &MockTool{ToolFunc: func() *v1.Tool { return toolA }}
	tB := &MockTool{ToolFunc: func() *v1.Tool { return toolB }}
	tC := &MockTool{ToolFunc: func() *v1.Tool { return toolC }}

	// Define Profiles using Builders
	profiles := []*configv1.ProfileDefinition{
		configv1.ProfileDefinition_builder{
			Name: proto.String("allow_all"),
			ServiceConfig: map[string]*configv1.ProfileServiceConfig{
				"service_1": configv1.ProfileServiceConfig_builder{Enabled: proto.Bool(true)}.Build(),
			},
		}.Build(),
		configv1.ProfileDefinition_builder{
			Name: proto.String("allow_subset"),
			ServiceConfig: map[string]*configv1.ProfileServiceConfig{
				"service_1": configv1.ProfileServiceConfig_builder{
					Enabled:      proto.Bool(true),
					AllowedTools: []string{"tool_a"},
				}.Build(),
			},
		}.Build(),
		configv1.ProfileDefinition_builder{
			Name: proto.String("block_specific"),
			ServiceConfig: map[string]*configv1.ProfileServiceConfig{
				"service_1": configv1.ProfileServiceConfig_builder{
					Enabled:      proto.Bool(true),
					BlockedTools: []string{"tool_b"},
				}.Build(),
			},
		}.Build(),
		configv1.ProfileDefinition_builder{
			Name: proto.String("disabled_service"),
			ServiceConfig: map[string]*configv1.ProfileServiceConfig{
				"service_1": configv1.ProfileServiceConfig_builder{Enabled: proto.Bool(false)}.Build(),
			},
		}.Build(),
	}

	tm.SetProfiles([]string{"allow_all", "allow_subset", "block_specific", "disabled_service"}, profiles)

	tests := []struct {
		name      string
		tool      Tool
		profileID string
		want      bool
	}{
		{"Allow All - Tool A", tA, "allow_all", true},
		{"Allow All - Tool B", tB, "allow_all", true},
		{"Allow Subset - Tool A (Allowed)", tA, "allow_subset", true},
		{"Allow Subset - Tool B (Not Allowed)", tB, "allow_subset", false},
		{"Block Specific - Tool A", tA, "block_specific", true},
		{"Block Specific - Tool B (Blocked)", tB, "block_specific", false},
		{"Disabled Service", tA, "disabled_service", false},
		{"Service Not Configured", tC, "allow_all", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tm.IsToolAllowed(tt.tool, tt.profileID)
			assert.Equal(t, tt.want, got)
		})
	}
}
