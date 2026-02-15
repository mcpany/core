package tool

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestManager_IsToolAllowed(t *testing.T) {
	t.Parallel()
	manager := NewManager(nil)

	// Define profiles
	devProfile := configv1.ProfileDefinition_builder{
		Name: proto.String("dev"),
		Selector: configv1.ProfileSelector_builder{
			Tags: []string{"dev"},
		}.Build(),
	}.Build()

	readonlyProfile := configv1.ProfileDefinition_builder{
		Name: proto.String("readonly"),
		Selector: configv1.ProfileSelector_builder{
			ToolProperties: map[string]string{
				"read_only": "true",
			},
		}.Build(),
	}.Build()

	mixedProfile := configv1.ProfileDefinition_builder{
		Name: proto.String("mixed"),
		Selector: configv1.ProfileSelector_builder{
			Tags: []string{"dev"},
			ToolProperties: map[string]string{
				"read_only": "true",
			},
		}.Build(),
	}.Build()

	manager.SetProfiles([]string{"dev", "readonly", "mixed"}, []*configv1.ProfileDefinition{devProfile, readonlyProfile, mixedProfile})

	// Helpers
	toolWithTags := func(tags ...string) *v1.Tool {
		return v1.Tool_builder{
			Tags: tags,
		}.Build()
	}

	toolWithProps := func(readonly bool) *v1.Tool {
		return v1.Tool_builder{
			Annotations: v1.ToolAnnotations_builder{
				ReadOnlyHint: proto.Bool(readonly),
			}.Build(),
		}.Build()
	}

	toolWithBoth := func(readonly bool, tags ...string) *v1.Tool {
		return v1.Tool_builder{
			Tags: tags,
			Annotations: v1.ToolAnnotations_builder{
				ReadOnlyHint: proto.Bool(readonly),
			}.Build(),
		}.Build()
	}

	tests := []struct {
		name     string
		enabled  []string
		tool     *v1.Tool
		expected bool
	}{
		{
			name:     "No profiles enabled (should allow all)",
			enabled:  []string{},
			tool:     toolWithTags("dev"),
			expected: true,
		},
		{
			name:     "Dev profile enabled, matches tag",
			enabled:  []string{"dev"},
			tool:     toolWithTags("dev"),
			expected: true,
		},
		{
			name:     "Dev profile enabled, no match",
			enabled:  []string{"dev"},
			tool:     toolWithTags("prod"),
			expected: false,
		},
		{
			name:     "Readonly profile enabled, matches property",
			enabled:  []string{"readonly"},
			tool:     toolWithProps(true),
			expected: true,
		},
		{
			name:     "Readonly profile enabled, no match",
			enabled:  []string{"readonly"},
			tool:     toolWithProps(false),
			expected: false,
		},
		{
			name:     "Mixed profile enabled, matches both",
			enabled:  []string{"mixed"},
			tool:     toolWithBoth(true, "dev"),
			expected: true,
		},
		{
			name:     "Mixed profile enabled, matches only tag",
			enabled:  []string{"mixed"},
			tool:     toolWithBoth(false, "dev"),
			expected: false,
		},
		{
			name:     "Mixed profile enabled, matches only prop",
			enabled:  []string{"mixed"},
			tool:     toolWithBoth(true, "prod"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager.enabledProfiles = tt.enabled
			allowed := manager.isToolAllowed(tt.tool)
			assert.Equal(t, tt.expected, allowed)
		})
	}
}

// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

func TestMatchesProperties_Extended(t *testing.T) {
	t.Parallel()
	tm := NewManager(nil)

	tests := []struct {
		name        string
		annotations *v1.ToolAnnotations
		props       map[string]string
		expected    bool
	}{
		{
			name: "Destructive true match",
			annotations: v1.ToolAnnotations_builder{
				DestructiveHint: proto.Bool(true),
			}.Build(),
			props:    map[string]string{"destructive": "true"},
			expected: true,
		},
		{
			name: "Destructive false match",
			annotations: v1.ToolAnnotations_builder{
				DestructiveHint: proto.Bool(false),
			}.Build(),
			props:    map[string]string{"destructive": "false"},
			expected: true,
		},
		{
			name: "Destructive mismatch",
			annotations: v1.ToolAnnotations_builder{
				DestructiveHint: proto.Bool(true),
			}.Build(),
			props:    map[string]string{"destructive": "false"},
			expected: false,
		},
		{
			name:        "Destructive nil annotations (defaults to false)",
			annotations: nil,
			props:       map[string]string{"destructive": "false"},
			expected:    true,
		},
		{
			name:        "Destructive nil annotations mismatch",
			annotations: nil,
			props:       map[string]string{"destructive": "true"},
			expected:    false,
		},
		{
			name:        "ReadOnly nil annotations (defaults to false)",
			annotations: nil,
			props:       map[string]string{"read_only": "false"},
			expected:    true,
		},
		{
			name:        "Idempotent nil annotations (defaults to false)",
			annotations: nil,
			props:       map[string]string{"idempotent": "false"},
			expected:    true,
		},
		{
			name:        "OpenWorld nil annotations (defaults to false)",
			annotations: nil,
			props:       map[string]string{"open_world": "false"},
			expected:    true,
		},
		{
			name: "Idempotent true match",
			annotations: v1.ToolAnnotations_builder{
				IdempotentHint: proto.Bool(true),
			}.Build(),
			props:    map[string]string{"idempotent": "true"},
			expected: true,
		},
		{
			name: "OpenWorld true match",
			annotations: v1.ToolAnnotations_builder{
				OpenWorldHint: proto.Bool(true),
			}.Build(),
			props:    map[string]string{"open_world": "true"},
			expected: true,
		},
		{
			name:        "Invalid property key",
			annotations: v1.ToolAnnotations_builder{}.Build(),
			props:       map[string]string{"unknown_key": "true"},
			expected:    false,
		},
		{
			name: "Multiple properties all match",
			annotations: v1.ToolAnnotations_builder{
				ReadOnlyHint:    proto.Bool(true),
				DestructiveHint: proto.Bool(false),
			}.Build(),
			props:    map[string]string{"read_only": "true", "destructive": "false"},
			expected: true,
		},
		{
			name: "Multiple properties one mismatch",
			annotations: v1.ToolAnnotations_builder{
				ReadOnlyHint:    proto.Bool(true),
				DestructiveHint: proto.Bool(true),
			}.Build(),
			props:    map[string]string{"read_only": "true", "destructive": "false"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tm.matchesProperties(tt.annotations, tt.props)
			assert.Equal(t, tt.expected, result)
		})
	}
}
