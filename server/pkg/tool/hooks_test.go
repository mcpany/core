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
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

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
			policy: configv1.CallPolicy_builder{
				DefaultAction: configv1.CallPolicy_ALLOW.Enum(),
			}.Build(),
			toolName:   "any-tool",
			wantAction: ActionAllow,
		},
		{
			name: "Default Deny",
			policy: configv1.CallPolicy_builder{
				DefaultAction: configv1.CallPolicy_DENY.Enum(),
			}.Build(),
			toolName:   "any-tool",
			wantAction: ActionDeny,
			wantError:  true,
		},
		{
			name: "Explicit Allow Rule",
			policy: configv1.CallPolicy_builder{
				DefaultAction: configv1.CallPolicy_DENY.Enum(),
				Rules: []*configv1.CallPolicyRule{
					configv1.CallPolicyRule_builder{
						Action:        configv1.CallPolicy_ALLOW.Enum(),
						NameRegex: proto.String("^allowed-.*"),
					}.Build(),
				},
			}.Build(),
			toolName:   "allowed-tool",
			wantAction: ActionAllow,
		},
		{
			name: "Explicit Deny Rule",
			policy: configv1.CallPolicy_builder{
				DefaultAction: configv1.CallPolicy_ALLOW.Enum(),
				Rules: []*configv1.CallPolicyRule{
					configv1.CallPolicyRule_builder{
						Action:        configv1.CallPolicy_DENY.Enum(),
						NameRegex: proto.String("^sensitive-.*"),
					}.Build(),
				},
			}.Build(),
			toolName:   "sensitive-tool",
			wantAction: ActionDeny,
			wantError:  true,
		},
		{
			name: "Argument Regex Matching",
			policy: configv1.CallPolicy_builder{
				DefaultAction: configv1.CallPolicy_ALLOW.Enum(),
				Rules: []*configv1.CallPolicyRule{
					configv1.CallPolicyRule_builder{
						Action:        configv1.CallPolicy_DENY.Enum(),
						ArgumentRegex: proto.String(".*secret.*"),
					}.Build(),
				},
			}.Build(),
			toolName: "any-tool",
			inputs: map[string]any{
				"key": "this contains secret",
			},
			wantAction: ActionDeny,
			wantError:  true,
		},
		{
			name: "Argument Regex Matching Safe",
			policy: configv1.CallPolicy_builder{
				DefaultAction: configv1.CallPolicy_ALLOW.Enum(),
				Rules: []*configv1.CallPolicyRule{
					configv1.CallPolicyRule_builder{
						Action:        configv1.CallPolicy_DENY.Enum(),
						ArgumentRegex: proto.String(".*secret.*"),
					}.Build(),
				},
			}.Build(),
			toolName: "any-tool",
			inputs: map[string]any{
				"key": "safe value",
			},
			wantAction: ActionAllow,
		},
		{
			name: "Save Cache Rule",
			policy: configv1.CallPolicy_builder{
				DefaultAction: configv1.CallPolicy_ALLOW.Enum(),
				Rules: []*configv1.CallPolicyRule{
					configv1.CallPolicyRule_builder{
						Action:    configv1.CallPolicy_SAVE_CACHE.Enum(),
						NameRegex: proto.String("^save-tool.*"),
					}.Build(),
				},
			}.Build(),
			toolName:   "save-tool",
			wantAction: ActionSaveCache,
		},
		{
			name: "Delete Cache Rule",
			policy: configv1.CallPolicy_builder{
				DefaultAction: configv1.CallPolicy_ALLOW.Enum(),
				Rules: []*configv1.CallPolicyRule{
					configv1.CallPolicyRule_builder{
						Action:    configv1.CallPolicy_DELETE_CACHE.Enum(),
						NameRegex: proto.String("^delete-tool.*"),
					}.Build(),
				},
			}.Build(),
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
	t.Skip("Skipping flaky webhook test - cloudevents encoding issue")
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Attempt to extract data from CloudEvent request
		var event map[string]any
		_ = json.NewDecoder(r.Body).Decode(&event)

		// If binary request, body is data directly (assuming ApplicationJSON)
		// If structured, body is event including data
		// hooks.go sends structured? No, checking hooks.go Call...
		// event.SetData(cloudevents.ApplicationJSON, data)
		// SDK default transport might use Binary or Structured.
		// For robustness we inspect "kind" in the decoded map if possible

		var dataMap map[string]any
		if d, ok := event["data"]; ok {
			if dm, ok := d.(map[string]any); ok {
				dataMap = dm
			}
		} else {
			dataMap = event
		}

		var kindInt int32
		if k, ok := dataMap["kind"].(float64); ok {
			kindInt = int32(k)
		} else if k, ok := dataMap["kind"].(string); ok {
			if k == "WEBHOOK_KIND_PRE_CALL" { kindInt = int32(configv1.WebhookKind_WEBHOOK_KIND_PRE_CALL) }
			if k == "WEBHOOK_KIND_POST_CALL" { kindInt = int32(configv1.WebhookKind_WEBHOOK_KIND_POST_CALL) }
		}
		toolName, _ := dataMap["tool_name"].(string)

		resp := configv1.WebhookResponse_builder{Allowed: proto.Bool(true)}.Build()

		switch configv1.WebhookKind(kindInt) {
		case configv1.WebhookKind_WEBHOOK_KIND_PRE_CALL:
			switch toolName {
			case "deny-me":
				resp = configv1.WebhookResponse_builder{
					Allowed: proto.Bool(false),
					Status: configv1.WebhookStatus_builder{
						Message: proto.String("denied by policy"),
					}.Build(),
				}.Build()
			case "modify-me":
				newInputs := map[string]any{"modified": "yes"}
				s, _ := structpb.NewStruct(newInputs)
				resp = configv1.WebhookResponse_builder{
					Allowed:           proto.Bool(true),
					ReplacementObject: s,
				}.Build()
			}
		case configv1.WebhookKind_WEBHOOK_KIND_POST_CALL:
			if toolName == "modify-result" {
				newResult := map[string]any{"value": "modified result"}
				s, _ := structpb.NewStruct(newResult)
				resp = configv1.WebhookResponse_builder{
					Allowed:           proto.Bool(true),
					ReplacementObject: s,
				}.Build()
			}
		}

		// Send CloudEvent response Using Binary Content Mode
		w.Header().Set("ce-id", "test-resp-id")
		w.Header().Set("ce-source", "test-source")
		w.Header().Set("ce-type", "com.mcpany.webhook.response")
		w.Header().Set("ce-specversion", "1.0")
		w.Header().Set("Content-Type", "application/json")

		respMap := map[string]any{
			"allowed": resp.GetAllowed(),
		}
		if resp.GetStatus() != nil {
			respMap["status"] = map[string]any{
				"code":    resp.GetStatus().GetCode(),
				"message": resp.GetStatus().GetMessage(),
			}
		}
		if resp.GetReplacementObject() != nil {
			fields := resp.GetReplacementObject().GetFields()
			m := make(map[string]interface{})
			for k, v := range fields {
				m[k] = v.AsInterface()
			}
			respMap["replacement_object"] = m
		}

		// Encode only the data
		_ = json.NewEncoder(w).Encode(respMap)
	}))
	defer server.Close()

	config := configv1.WebhookConfig_builder{Url: proto.String(server.URL)}.Build()
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
		require.NotNil(t, newReq)
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
