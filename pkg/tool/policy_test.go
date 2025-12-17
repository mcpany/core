// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func actionPtr(a configv1.ExportPolicy_Action) *configv1.ExportPolicy_Action {
	return &a
}

func TestShouldExport(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		policy   *configv1.ExportPolicy
		want     bool
	}{
		{
			name:     "Nil Policy",
			toolName: "any",
			policy:   nil,
			want:     true,
		},
		{
			name:     "Default Action Unspecified",
			toolName: "any",
			policy:   &configv1.ExportPolicy{},
			want:     true,
		},
		{
			name:     "Default Action Export",
			toolName: "any",
			policy: &configv1.ExportPolicy{
				DefaultAction: actionPtr(configv1.ExportPolicy_EXPORT),
			},
			want: true,
		},
		{
			name:     "Default Action Unexport",
			toolName: "any",
			policy: &configv1.ExportPolicy{
				DefaultAction: actionPtr(configv1.ExportPolicy_UNEXPORT),
			},
			want: false,
		},
		{
			name:     "Rule Match Export",
			toolName: "allowed_tool",
			policy: &configv1.ExportPolicy{
				DefaultAction: actionPtr(configv1.ExportPolicy_UNEXPORT),
				Rules: []*configv1.ExportRule{
					{
						NameRegex: strPtr("^allowed_.*"),
						Action:    actionPtr(configv1.ExportPolicy_EXPORT),
					},
				},
			},
			want: true,
		},
		{
			name:     "Rule Match Unexport",
			toolName: "hidden_tool",
			policy: &configv1.ExportPolicy{
				DefaultAction: actionPtr(configv1.ExportPolicy_EXPORT),
				Rules: []*configv1.ExportRule{
					{
						NameRegex: strPtr("^hidden_.*"),
						Action:    actionPtr(configv1.ExportPolicy_UNEXPORT),
					},
				},
			},
			want: false,
		},
		{
			name:     "Rule No Match Fallthrough",
			toolName: "other_tool",
			policy: &configv1.ExportPolicy{
				DefaultAction: actionPtr(configv1.ExportPolicy_UNEXPORT),
				Rules: []*configv1.ExportRule{
					{
						NameRegex: strPtr("^allowed_.*"),
						Action:    actionPtr(configv1.ExportPolicy_EXPORT),
					},
				},
			},
			want: false,
		},
		{
			name:     "Invalid Regex",
			toolName: "any",
			policy: &configv1.ExportPolicy{
				Rules: []*configv1.ExportRule{
					{
						NameRegex: strPtr("["), // Invalid regex
						Action:    actionPtr(configv1.ExportPolicy_UNEXPORT),
					},
				},
			},
			want: true, // Should continue and use default (true)
		},
		{
			name:     "Empty Regex",
			toolName: "any",
			policy: &configv1.ExportPolicy{
				Rules: []*configv1.ExportRule{
					{
						NameRegex: strPtr(""),
						Action:    actionPtr(configv1.ExportPolicy_UNEXPORT),
					},
				},
			},
			want: true, // Should skipped empty regex
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldExport(tt.toolName, tt.policy)
			assert.Equal(t, tt.want, got)
		})
	}
}

func strPtr(s string) *string { return &s }
