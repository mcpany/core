// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"

	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestManager_matchesProperties_Coverage(t *testing.T) {
	manager := NewManager(nil)

	tests := []struct {
		name        string
		annotations *v1.ToolAnnotations
		props       map[string]string
		expected    bool
	}{
		{
			name:        "read_only=true match",
			annotations: &v1.ToolAnnotations{ReadOnlyHint: proto.Bool(true)},
			props:       map[string]string{"read_only": "true"},
			expected:    true,
		},
		{
			name:        "read_only=true mismatch",
			annotations: &v1.ToolAnnotations{ReadOnlyHint: proto.Bool(false)},
			props:       map[string]string{"read_only": "true"},
			expected:    false,
		},
		{
			name:        "read_only nil annotations",
			annotations: nil,
			props:       map[string]string{"read_only": "false"},
			expected:    true,
		},
		{
			name:        "destructive=true match",
			annotations: &v1.ToolAnnotations{DestructiveHint: proto.Bool(true)},
			props:       map[string]string{"destructive": "true"},
			expected:    true,
		},
		{
			name:        "idempotent=true match",
			annotations: &v1.ToolAnnotations{IdempotentHint: proto.Bool(true)},
			props:       map[string]string{"idempotent": "true"},
			expected:    true,
		},
		{
			name:        "open_world=true match",
			annotations: &v1.ToolAnnotations{OpenWorldHint: proto.Bool(true)},
			props:       map[string]string{"open_world": "true"},
			expected:    true,
		},
		{
			name:        "unknown property",
			annotations: &v1.ToolAnnotations{},
			props:       map[string]string{"unknown_prop": "val"},
			expected:    false,
		},
		{
			name:        "multiple properties match",
			annotations: &v1.ToolAnnotations{ReadOnlyHint: proto.Bool(true), IdempotentHint: proto.Bool(true)},
			props:       map[string]string{"read_only": "true", "idempotent": "true"},
			expected:    true,
		},
		{
			name:        "multiple properties mismatch",
			annotations: &v1.ToolAnnotations{ReadOnlyHint: proto.Bool(true), IdempotentHint: proto.Bool(false)},
			props:       map[string]string{"read_only": "true", "idempotent": "true"},
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.matchesProperties(tt.annotations, tt.props)
			assert.Equal(t, tt.expected, result)
		})
	}
}
