// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
				Rules: []*configv1.CallPolicyRule{
					{
						Action:        ptr(configv1.CallPolicy_ALLOW),
						NameRegex: ptr("^allowed-.*"),
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
				Rules: []*configv1.CallPolicyRule{
					{
						Action:        ptr(configv1.CallPolicy_DENY),
						NameRegex: ptr("^sensitive-.*"),
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
				Rules: []*configv1.CallPolicyRule{
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
				Rules: []*configv1.CallPolicyRule{
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

func TestWebhookHook(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req webhookRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp := webhookResponse{Action: "allow"}

		switch req.HookType {
		case "pre":
			switch req.ToolName {
			case "deny-me":
				resp.Action = "deny"
				resp.Error = "denied by policy"
			case "modify-me":
				// Modify inputs
				newInputs := map[string]string{"modified": "yes"}
				b, _ := json.Marshal(newInputs)
				resp.Inputs = b
			}
		case "post":
			if req.ToolName == "modify-result" {
				resp.Result = "modified result"
			}
		}

		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	config := &configv1.WebhookConfig{Url: ptr(server.URL)}
	hook := NewWebhookHook(config)

	t.Run("Pre Allow", func(t *testing.T) {
		req := &ExecutionRequest{ToolName: "allowed-tool", ToolInputs: json.RawMessage("{}")}
		action, newReq, err := hook.ExecutePre(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, ActionAllow, action)
		assert.Nil(t, newReq)
	})

	t.Run("Pre Deny", func(t *testing.T) {
		req := &ExecutionRequest{ToolName: "deny-me", ToolInputs: json.RawMessage("{}")}
		action, _, err := hook.ExecutePre(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "denied by webhook")
		assert.Equal(t, ActionDeny, action)
	})

	t.Run("Pre Modify", func(t *testing.T) {
		req := &ExecutionRequest{ToolName: "modify-me", ToolInputs: json.RawMessage("{}")}
		action, newReq, err := hook.ExecutePre(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, ActionAllow, action)
		assert.NotNil(t, newReq)
		assert.Contains(t, string(newReq.ToolInputs), "modified")
	})

	t.Run("Post No Modification", func(t *testing.T) {
		req := &ExecutionRequest{ToolName: "allowed-tool"}
		res, err := hook.ExecutePost(context.Background(), req, "original")
		require.NoError(t, err)
		assert.Equal(t, "original", res)
	})

	t.Run("Post Modify", func(t *testing.T) {
		req := &ExecutionRequest{ToolName: "modify-result"}
		res, err := hook.ExecutePost(context.Background(), req, "original")
		require.NoError(t, err)
		assert.Equal(t, "modified result", res)
	})
}

func TestHTMLToMarkdownHook_ExecutePost(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		want     any
	}{
		{
			name:  "String HTML",
			input: "<b>bold</b>",
			want:  "**bold**",
		},
		{
			name:  "Map HTML",
			input: map[string]any{"content": "<i>italic</i>"},
			want:  map[string]any{"content": "_italic_"},
		},
		{
			name:  "Nested Map HTML",
			input: map[string]any{"nested": map[string]any{"val": "<p>para</p>"}},
			want:  map[string]any{"nested": map[string]any{"val": "para"}},
		},
		{
			name:  "Mixed Content",
			input: map[string]any{"plain": "text", "html": "<h1>Header</h1>"},
			want:  map[string]any{"plain": "text", "html": "# Header"},
		},
	}

	hook := NewHTMLToMarkdownHook(&configv1.HtmlToMarkdownConfig{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := hook.ExecutePost(context.Background(), nil, tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
