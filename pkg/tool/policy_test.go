/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
