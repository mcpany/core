// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/mcpany/core/pkg/logging"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// PolicyHook implements PreCallHook using CallPolicy.
type PolicyHook struct {
	policy *configv1.CallPolicy
}

// NewPolicyHook creates a new PolicyHook with the given call policy.
func NewPolicyHook(policy *configv1.CallPolicy) *PolicyHook {
	return &PolicyHook{policy: policy}
}

// ExecutePre executes the policy check before a tool is called.
func (h *PolicyHook) ExecutePre(
	_ context.Context,
	req *ExecutionRequest,
) (Action, *ExecutionRequest, error) {
	// Determine default action
	allowed := h.policy.GetDefaultAction() == configv1.CallPolicy_ALLOW

	for _, rule := range h.policy.GetRules() {
		// 1. Match Tool Name
		if rule.GetNameRegex() != "" {
			matched, err := regexp.MatchString(rule.GetNameRegex(), req.ToolName)
			if err != nil {
				logging.GetLogger().
					Error("Invalid tool name regex in policy", "regex", rule.GetNameRegex(), "error", err)
				continue // Skip invalid rule
			}
			if !matched {
				continue // Rule doesn't apply
			}
		}

		// 2. Match Arguments
		if rule.GetArgumentRegex() != "" {
			// req.ToolInputs is json.RawMessage ([]byte)
			matched, err := regexp.MatchString(rule.GetArgumentRegex(), string(req.ToolInputs))
			if err != nil {
				logging.GetLogger().
					Error("Invalid argument regex in policy", "regex", rule.GetArgumentRegex(), "error", err)
				continue
			}
			if !matched {
				continue
			}
		}

		// Rule matched!
		if rule.GetAction() == configv1.CallPolicy_ALLOW {
			return ActionAllow, nil, nil
		}
		return ActionDeny, nil, fmt.Errorf("tool execution denied by policy rule: %s", req.ToolName)
	}

	if allowed {
		return ActionAllow, nil, nil
	}
	return ActionDeny, nil, fmt.Errorf("tool execution denied by default policy: %s", req.ToolName)
}

// (Deprecated hooks removed)

// WebhookHook supports modification of requests and responses via external webhook.
type WebhookHook struct {
	url     string
	timeout time.Duration
	client  *http.Client
}

// NewWebhookHook creates a new WebhookHook.
func NewWebhookHook(config *configv1.WebhookConfig) *WebhookHook {
	timeout := 5 * time.Second
	if t := config.GetTimeout(); t != nil {
		timeout = t.AsDuration()
	}
	return &WebhookHook{
		url:     config.GetUrl(),
		timeout: timeout,
		client:  &http.Client{Timeout: timeout},
	}
}

// ExecutePre executes the webhook notification before a tool is called.
func (h *WebhookHook) ExecutePre(
	ctx context.Context,
	req *ExecutionRequest,
) (Action, *ExecutionRequest, error) {
	// Convert inputs to Struct
	inputsMap := make(map[string]any)
	if len(req.ToolInputs) > 0 {
		if err := json.Unmarshal(req.ToolInputs, &inputsMap); err != nil {
			return ActionDeny, nil, fmt.Errorf("failed to unmarshal inputs: %w", err)
		}
	}
	inputsStruct, err := structpb.NewStruct(inputsMap)
	if err != nil {
		return ActionDeny, nil, fmt.Errorf("failed to convert inputs to struct: %w", err)
	}

	reviewReq := &configv1.WebhookRequest{
		Uid:      uuid.New().String(),
		Kind:     "PreCall",
		ToolName: req.ToolName,
		Object:   inputsStruct,
	}

	reviewResp, err := h.callWebhook(ctx, reviewReq)
	if err != nil {
		return ActionDeny, nil, fmt.Errorf("webhook error: %w", err)
	}

	if !reviewResp.GetAllowed() {
		msg := "denied by webhook"
		if reviewResp.GetStatus() != nil {
			msg = fmt.Sprintf("%s: %s", msg, reviewResp.GetStatus().GetMessage())
		}
		return ActionDeny, nil, fmt.Errorf(msg)
	}

	if reviewResp.GetReplacementObject() != nil {
		// Modify inputs
		newInputsMap := reviewResp.GetReplacementObject().AsMap()
		newInputsAPI, err := json.Marshal(newInputsMap)
		if err != nil {
			return ActionDeny, nil, fmt.Errorf("failed to marshal new inputs: %w", err)
		}
		newReq := *req
		newReq.ToolInputs = newInputsAPI
		return ActionAllow, &newReq, nil
	}

	return ActionAllow, nil, nil
}

// ExecutePost executes the webhook notification after a tool is called.
func (h *WebhookHook) ExecutePost(
	ctx context.Context,
	req *ExecutionRequest,
	result any,
) (any, error) {
	// Convert result to Struct
	// Result can be string, map, slice, etc. structpb only supports map[string]any as root.
	// If result is not a map, we might need to wrap it?
	// Or maybe the webhook protocol expects an object?
	// For "simple" results (string), let's wrap in {"value": ...}?
	// Or if the standardized webhook expects Struct, we MUST provide Struct.
	
	resultMap := make(map[string]any)
	if m, ok := result.(map[string]any); ok {
		resultMap = m
	} else {
		// Wrap non-map result
		resultMap["value"] = result
	}

	resultStruct, err := structpb.NewStruct(resultMap)
	if err != nil {
		return nil, fmt.Errorf("failed to convert result to struct: %w", err)
	}

	reviewReq := &configv1.WebhookRequest{
		Uid:      uuid.New().String(),
		Kind:     "PostCall",
		ToolName: req.ToolName,
		Object:   resultStruct,
	}

	reviewResp, err := h.callWebhook(ctx, reviewReq)
	if err != nil {
		return nil, fmt.Errorf("webhook error: %w", err)
	}

	if reviewResp.GetReplacementObject() != nil {
		newResultMap := reviewResp.GetReplacementObject().AsMap()
		// Unwrap if we wrapped it?
		// If original was NOT a map, and we receive a map with "value", should we unwrap?
		// The webhook might return a full map structure.
		// If the webhook is smart, it returns what we expect.
		// If we wrapped it, we should check if we should unwrap.
		// For now, return the map unless it has only "value" and original was not map?
		// Let's rely on the structure modification. 
		// If the tool return type expects a string, and we return a map, it might break.
		// But in `tool.go`, result is `any`.
		// If we wrapped it in "value", check if "value" exists in replacement.
		if _, wasMap := result.(map[string]any); !wasMap {
			if v, ok := newResultMap["value"]; ok && len(newResultMap) == 1 {
				return v, nil
			}
		}
		return newResultMap, nil
	}

	return result, nil
}

func (h *WebhookHook) callWebhook(ctx context.Context, req *configv1.WebhookRequest) (*configv1.WebhookResponse, error) {
	review := &configv1.WebhookReview{
		Request: req,
	}
	
	body, err := json.Marshal(review)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal review: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, h.url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := h.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}
	defer func() {
		if err := httpResp.Body.Close(); err != nil {
			logging.GetLogger().Warn("Failed to close webhook response body", "error", err)
		}
	}()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", httpResp.StatusCode)
	}

	var respReview configv1.WebhookReview
	if err := json.NewDecoder(httpResp.Body).Decode(&respReview); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	if respReview.Response == nil {
		return nil, fmt.Errorf("empty response from webhook")
	}
	
	return respReview.Response, nil
}
