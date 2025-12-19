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

func TestManager_IsToolAllowed(t *testing.T) {
	manager := NewManager(nil)

	// Define profiles
	devProfile := &configv1.ProfileDefinition{
		Name: proto.String("dev"),
		Selector: &configv1.ProfileSelector{
			Tags: []string{"dev"},
		},
	}

	readonlyProfile := &configv1.ProfileDefinition{
		Name: proto.String("readonly"),
		Selector: &configv1.ProfileSelector{
			ToolProperties: map[string]string{
				"read_only": "true",
			},
		},
	}

	notReadonlyProfile := &configv1.ProfileDefinition{
		Name: proto.String("not_readonly"),
		Selector: &configv1.ProfileSelector{
			ToolProperties: map[string]string{
				"read_only": "false",
			},
		},
	}

	mixedProfile := &configv1.ProfileDefinition{
		Name: proto.String("mixed"),
		Selector: &configv1.ProfileSelector{
			Tags: []string{"dev"},
			ToolProperties: map[string]string{
				"read_only": "true",
			},
		},
	}

	destructiveProfile := &configv1.ProfileDefinition{
		Name: proto.String("destructive"),
		Selector: &configv1.ProfileSelector{
			ToolProperties: map[string]string{
				"destructive": "true",
			},
		},
	}

	idempotentProfile := &configv1.ProfileDefinition{
		Name: proto.String("idempotent"),
		Selector: &configv1.ProfileSelector{
			ToolProperties: map[string]string{
				"idempotent": "true",
			},
		},
	}

	openWorldProfile := &configv1.ProfileDefinition{
		Name: proto.String("open_world"),
		Selector: &configv1.ProfileSelector{
			ToolProperties: map[string]string{
				"open_world": "true",
			},
		},
	}

	invalidPropProfile := &configv1.ProfileDefinition{
		Name: proto.String("invalid_prop"),
		Selector: &configv1.ProfileSelector{
			ToolProperties: map[string]string{
				"unknown_property": "true",
			},
		},
	}

	manager.SetProfiles(
		[]string{"dev", "readonly", "mixed", "destructive", "idempotent", "open_world", "invalid_prop", "not_readonly"},
		[]*configv1.ProfileDefinition{devProfile, readonlyProfile, mixedProfile, destructiveProfile, idempotentProfile, openWorldProfile, invalidPropProfile, notReadonlyProfile},
	)

	// Helpers
	toolWithTags := func(tags ...string) *v1.Tool {
		return &v1.Tool{
			Tags: tags,
		}
	}

	toolWithProps := func(readonly bool) *v1.Tool {
		return &v1.Tool{
			Annotations: &v1.ToolAnnotations{
				ReadOnlyHint: proto.Bool(readonly),
			},
		}
	}

	toolWithDestructive := func(destructive bool) *v1.Tool {
		return &v1.Tool{
			Annotations: &v1.ToolAnnotations{
				DestructiveHint: proto.Bool(destructive),
			},
		}
	}

	toolWithIdempotent := func(idempotent bool) *v1.Tool {
		return &v1.Tool{
			Annotations: &v1.ToolAnnotations{
				IdempotentHint: proto.Bool(idempotent),
			},
		}
	}

	toolWithOpenWorld := func(openWorld bool) *v1.Tool {
		return &v1.Tool{
			Annotations: &v1.ToolAnnotations{
				OpenWorldHint: proto.Bool(openWorld),
			},
		}
	}

	toolWithBoth := func(readonly bool, tags ...string) *v1.Tool {
		return &v1.Tool{
			Tags: tags,
			Annotations: &v1.ToolAnnotations{
				ReadOnlyHint: proto.Bool(readonly),
			},
		}
	}

	toolWithNilAnnotations := func() *v1.Tool {
		return &v1.Tool{
			Annotations: nil,
		}
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
		{
			name:     "Destructive profile enabled, matches property",
			enabled:  []string{"destructive"},
			tool:     toolWithDestructive(true),
			expected: true,
		},
		{
			name:     "Destructive profile enabled, no match",
			enabled:  []string{"destructive"},
			tool:     toolWithDestructive(false),
			expected: false,
		},
		{
			name:     "Idempotent profile enabled, matches property",
			enabled:  []string{"idempotent"},
			tool:     toolWithIdempotent(true),
			expected: true,
		},
		{
			name:     "OpenWorld profile enabled, matches property",
			enabled:  []string{"open_world"},
			tool:     toolWithOpenWorld(true),
			expected: true,
		},
		{
			name:     "Invalid property profile enabled",
			enabled:  []string{"invalid_prop"},
			tool:     toolWithProps(true),
			expected: false,
		},
		{
			name:     "Readonly profile enabled, nil annotations (default false)",
			enabled:  []string{"readonly"},
			tool:     toolWithNilAnnotations(),
			expected: false,
		},
		{
			name:     "NotReadonly profile enabled, nil annotations (default false)",
			enabled:  []string{"not_readonly"},
			tool:     toolWithNilAnnotations(),
			expected: true,
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
