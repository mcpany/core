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
						Action:    configv1.CallPolicy_ALLOW.Enum(),
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
						Action:    configv1.CallPolicy_DENY.Enum(),
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
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqMap map[string]any
		if err := json.NewDecoder(r.Body).Decode(&reqMap); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp := configv1.WebhookResponse_builder{Allowed: true}.Build()

		kindValue := int32(reqMap["kind"].(float64))
		toolName := reqMap["tool_name"].(string)

		switch configv1.WebhookKind(kindValue) {
		case configv1.WebhookKind_WEBHOOK_KIND_PRE_CALL:
			switch toolName {
			case "deny-me":
				resp.SetAllowed(false)
				resp.SetStatus(configv1.WebhookStatus_builder{Message: "denied by policy"}.Build())
			case "modify-me":
				// Modify inputs
				newInputs := map[string]any{"modified": "yes"}
				s, _ := structpb.NewStruct(newInputs)
				resp.SetReplacementObject(s)
			}
		case configv1.WebhookKind_WEBHOOK_KIND_POST_CALL:
			if toolName == "modify-result" {
				// hooks.go unwraps "value" if original result was likely a primitive
				newResult := map[string]any{"value": "modified result"}
				s, _ := structpb.NewStruct(newResult)
				resp.SetReplacementObject(s)
			}
		}

		// Send CloudEvent response (Binary Mode)
		w.Header().Set("ce-id", "test-resp-id")
		w.Header().Set("ce-source", "test-source")
		w.Header().Set("ce-specversion", "1.0")
		w.Header().Set("ce-type", "com.mcpany.webhook.response")
		w.Header().Set("Content-Type", "application/json")

		respMap := map[string]any{
			"allowed": resp.GetAllowed(),
		}
		if resp.HasStatus() {
			respMap["status"] = map[string]any{
				"code":    resp.GetStatus().GetCode(),
				"message": resp.GetStatus().GetMessage(),
			}
		}
		if resp.HasReplacementObject() {
			// Convert structpb.Struct to map for JSON encoding
			replacementMap := resp.GetReplacementObject().AsMap()
			respMap["replacement_object"] = replacementMap
		}

		_ = json.NewEncoder(w).Encode(respMap)
	}))
	defer server.Close()

	config := configv1.WebhookConfig_builder{Url: server.URL}.Build()
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
