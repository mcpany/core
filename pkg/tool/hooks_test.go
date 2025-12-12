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
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptr[T any](v T) *T {
	return &v
}

func TestPolicyHook_ExecutePre(t *testing.T) {
	tests := []struct {
		name       string
		policy     *configv1.CallPolicy
		toolName   string
		inputs     map[string]any
		wantAction Action
		wantError  bool
	}{
		{
			name: "Default Allow",
			policy: &configv1.CallPolicy{
				DefaultAction: ptr(configv1.CallPolicy_ALLOW),
			},
			toolName:   "any-tool",
			wantAction: ActionAllow,
		},
		{
			name: "Default Deny",
			policy: &configv1.CallPolicy{
				DefaultAction: ptr(configv1.CallPolicy_DENY),
			},
			toolName:   "any-tool",
			wantAction: ActionDeny,
			wantError:  true,
		},
		{
			name: "Explicit Allow Rule",
			policy: &configv1.CallPolicy{
				DefaultAction: ptr(configv1.CallPolicy_DENY),
				Rules: []*configv1.PolicyRule{
					{
						Action:        ptr(configv1.CallPolicy_ALLOW),
						ToolNameRegex: ptr("^allowed-.*"),
					},
				},
			},
			toolName:   "allowed-tool",
			wantAction: ActionAllow,
		},
		{
			name: "Explicit Deny Rule",
			policy: &configv1.CallPolicy{
				DefaultAction: ptr(configv1.CallPolicy_ALLOW),
				Rules: []*configv1.PolicyRule{
					{
						Action:        ptr(configv1.CallPolicy_DENY),
						ToolNameRegex: ptr("^sensitive-.*"),
					},
				},
			},
			toolName:   "sensitive-tool",
			wantAction: ActionDeny,
			wantError:  true,
		},
		{
			name: "Argument Regex Matching",
			policy: &configv1.CallPolicy{
				DefaultAction: ptr(configv1.CallPolicy_ALLOW),
				Rules: []*configv1.PolicyRule{
					{
						Action:        ptr(configv1.CallPolicy_DENY),
						ArgumentRegex: ptr(".*secret.*"),
					},
				},
			},
			toolName: "any-tool",
			inputs: map[string]any{
				"key": "this contains secret",
			},
			wantAction: ActionDeny,
			wantError:  true,
		},
		{
			name: "Argument Regex Matching Safe",
			policy: &configv1.CallPolicy{
				DefaultAction: ptr(configv1.CallPolicy_ALLOW),
				Rules: []*configv1.PolicyRule{
					{
						Action:        ptr(configv1.CallPolicy_DENY),
						ArgumentRegex: ptr(".*secret.*"),
					},
				},
			},
			toolName: "any-tool",
			inputs: map[string]any{
				"key": "safe value",
			},
			wantAction: ActionAllow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook := NewPolicyHook(tt.policy)
			inputsBytes, _ := json.Marshal(tt.inputs)
			req := &ExecutionRequest{
				ToolName:   tt.toolName,
				ToolInputs: inputsBytes,
			}

			action, _, err := hook.ExecutePre(context.Background(), req)

			assert.Equal(t, tt.wantAction, action)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTextTruncationHook_ExecutePost(t *testing.T) {
	tests := []struct {
		name     string
		maxChars int32
		input    any
		want     any
	}{
		{
			name:     "No Truncation needed",
			maxChars: 10,
			input:    "short",
			want:     "short",
		},
		{
			name:     "Truncation needed",
			maxChars: 5,
			input:    "long string",
			want:     "long ...",
		},
		{
			name:     "Map Truncation",
			maxChars: 5,
			input: map[string]any{
				"key1": "short",
				"key2": "long string",
				"nested": map[string]any{
					"key3": "very long string",
				},
			},
			want: map[string]any{
				"key1": "short",
				"key2": "long ...",
				"nested": map[string]any{
					"key3": "very ...",
				},
			},
		},
		{
			name:     "Disabled",
			maxChars: 0,
			input:    "long string",
			want:     "long string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook := NewTextTruncationHook(
				&configv1.TextTruncationConfig{MaxChars: ptr(tt.maxChars)},
			)
			req := &ExecutionRequest{} // ignored by this hook
			got, err := hook.ExecutePost(context.Background(), req, tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
