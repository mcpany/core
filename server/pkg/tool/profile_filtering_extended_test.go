package tool

import (
	"testing"

	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestMatchesProperties_Extended(t *testing.T) {
	tm := NewManager(nil)

	tests := []struct {
		name        string
		annotations *v1.ToolAnnotations
		props       map[string]string
		expected    bool
	}{
		{
			name: "Destructive true match",
			annotations: &v1.ToolAnnotations{
				DestructiveHint: proto.Bool(true),
			},
			props:    map[string]string{"destructive": "true"},
			expected: true,
		},
		{
			name: "Destructive false match",
			annotations: &v1.ToolAnnotations{
				DestructiveHint: proto.Bool(false),
			},
			props:    map[string]string{"destructive": "false"},
			expected: true,
		},
		{
			name: "Destructive mismatch",
			annotations: &v1.ToolAnnotations{
				DestructiveHint: proto.Bool(true),
			},
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
			annotations: &v1.ToolAnnotations{
				IdempotentHint: proto.Bool(true),
			},
			props:    map[string]string{"idempotent": "true"},
			expected: true,
		},
		{
			name: "OpenWorld true match",
			annotations: &v1.ToolAnnotations{
				OpenWorldHint: proto.Bool(true),
			},
			props:    map[string]string{"open_world": "true"},
			expected: true,
		},
		{
			name: "Invalid property key",
			annotations: &v1.ToolAnnotations{},
			props:    map[string]string{"unknown_key": "true"},
			expected: false,
		},
		{
			name: "Multiple properties all match",
			annotations: &v1.ToolAnnotations{
				ReadOnlyHint:    proto.Bool(true),
				DestructiveHint: proto.Bool(false),
			},
			props:    map[string]string{"read_only": "true", "destructive": "false"},
			expected: true,
		},
		{
			name: "Multiple properties one mismatch",
			annotations: &v1.ToolAnnotations{
				ReadOnlyHint:    proto.Bool(true),
				DestructiveHint: proto.Bool(true),
			},
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
