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
	"google.golang.org/protobuf/types/known/structpb"
)

func ptr[T any](v T) *T {
	return &v
}

func TestPolicyHook_ExecutePre(t *testing.T) {
	t.Parallel()
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
		{
			name: "Save Cache Rule",
			policy: &configv1.CallPolicy{
				DefaultAction: ptr(configv1.CallPolicy_ALLOW),
				Rules: []*configv1.CallPolicyRule{
					{
						Action:    ptr(configv1.CallPolicy_SAVE_CACHE),
						NameRegex: ptr("^save-tool.*"),
					},
				},
			},
			toolName:   "save-tool",
			wantAction: ActionSaveCache,
		},
		{
			name: "Delete Cache Rule",
			policy: &configv1.CallPolicy{
				DefaultAction: ptr(configv1.CallPolicy_ALLOW),
				Rules: []*configv1.CallPolicyRule{
					{
						Action:    ptr(configv1.CallPolicy_DELETE_CACHE),
						NameRegex: ptr("^delete-tool.*"),
					},
				},
			},
			toolName:   "delete-tool",
			wantAction: ActionDeleteCache,
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


func TestWebhookHook(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req configv1.WebhookRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp := &configv1.WebhookResponse{Allowed: true}

		switch req.Kind {
		case configv1.WebhookKind_WEBHOOK_KIND_PRE_CALL:
			switch req.ToolName {
			case "deny-me":
				resp.Allowed = false
				resp.Status = &configv1.WebhookStatus{Message: "denied by policy"}
			case "modify-me":
				// Modify inputs
				newInputs := map[string]any{"modified": "yes"}
				s, _ := structpb.NewStruct(newInputs)
				resp.ReplacementObject = s
			}
		case configv1.WebhookKind_WEBHOOK_KIND_POST_CALL:
			if req.ToolName == "modify-result" {
				// hooks.go unwraps "value" if original result was likely a primitive
				newResult := map[string]any{"value": "modified result"}
				s, _ := structpb.NewStruct(newResult)
				resp.ReplacementObject = s
			}
		}

		// Send CloudEvent response (Binary Mode)
		w.Header().Set("ce-id", "test-resp-id")
		w.Header().Set("ce-source", "test-source")
		w.Header().Set("ce-specversion", "1.0")
		w.Header().Set("ce-type", "com.mcpany.webhook.response")
		w.Header().Set("Content-Type", "application/json")

		respMap := map[string]any{
			"allowed": resp.Allowed,
		}
		if resp.Status != nil {
			respMap["status"] = map[string]any{
				"code":    resp.Status.Code,
				"message": resp.Status.Message,
			}
		}
		if resp.ReplacementObject != nil {
			// structpb.Struct to map
			respMap["replacement_object"] = resp.ReplacementObject
		}

		_ = json.NewEncoder(w).Encode(respMap)
	}))
	defer server.Close()

	config := &configv1.WebhookConfig{Url: server.URL}
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
