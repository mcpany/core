// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/google/uuid"
	"github.com/mcpany/core/pkg/logging"
	configv1 "github.com/mcpany/core/proto/config/v1"
	webhook "github.com/standard-webhooks/standard-webhooks/libraries/go"
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

// WebhookHook supports modification of requests and responses via external webhook using CloudEvents.
type WebhookHook struct {
	url     string
	timeout time.Duration
	client  *http.Client
	webhook *webhook.Webhook // Keep for signature if needed, or replace with CloudEvents signature
}

// NewWebhookHook creates a new WebhookHook.
func NewWebhookHook(config *configv1.WebhookConfig) *WebhookHook {
	timeout := 5 * time.Second
	if t := config.GetTimeout(); t != nil {
		timeout = t.AsDuration()
	}
	var wh *webhook.Webhook
	if secret := config.GetWebhookSecret(); secret != "" {
		var err error
		wh, err = webhook.NewWebhook(secret)
		if err != nil {
			logging.GetLogger().Error("Failed to create webhook signer", "error", err)
		}
	}

	// Create client with signing transport if webhook signer is present
	client := &http.Client{Timeout: timeout}
	if wh != nil {
		client.Transport = &SigningRoundTripper{
			signer: wh,
			base:   http.DefaultTransport,
		}
	}

	return &WebhookHook{
		url:     config.GetUrl(),
		timeout: timeout,
		client:  client,
		webhook: wh,
	}
}

// SigningRoundTripper signs the request body.
type SigningRoundTripper struct {
	signer *webhook.Webhook
	base   http.RoundTripper
}

// RoundTrip implements http.RoundTripper.
func (s *SigningRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body for signing: %w", err)
		}
		_ = req.Body.Close()
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		msgID := uuid.New().String()
		now := time.Now()
		signature, err := s.signer.Sign(msgID, now, bodyBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to sign request: %w", err)
		}
		req.Header.Set("webhook-signature", signature)
		req.Header.Set("webhook-id", msgID)
		req.Header.Set("webhook-timestamp", fmt.Sprintf("%d", now.Unix()))
	}

	base := s.base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(req)
}

// WebhookStatus represents the status of a webhook response.
type WebhookStatus struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

// ExecutePre executes the webhook notification before a tool is called.
func (h *WebhookHook) ExecutePre(
	ctx context.Context,
	req *ExecutionRequest,
) (Action, *ExecutionRequest, error) {
	// Convert inputs to Map for clearer JSON
	inputsMap := make(map[string]any)
	if len(req.ToolInputs) > 0 {
		if err := json.Unmarshal(req.ToolInputs, &inputsMap); err != nil {
			return ActionDeny, nil, fmt.Errorf("failed to unmarshal inputs: %w", err)
		}
	}

	event := cloudevents.NewEvent()
	event.SetID(uuid.New().String())
	event.SetSource("https://github.com/mcpany/core")
	event.SetType("com.mcpany.tool.pre_call")
	event.SetTime(time.Now())

	data := map[string]any{
		"kind":      configv1.WebhookKind_WEBHOOK_KIND_PRE_CALL,
		"tool_name": req.ToolName,
		"inputs":    inputsMap,
	}
	if err := event.SetData(cloudevents.ApplicationJSON, data); err != nil {
		return ActionDeny, nil, fmt.Errorf("failed to set cloud event data: %w", err)
	}

	respEvent, err := h.callWebhook(ctx, event)
	if err != nil {
		return ActionDeny, nil, fmt.Errorf("webhook error: %w", err)
	}

	// Helper struct for response data
	type ResponseData struct {
		Allowed           bool            `json:"allowed"`
		Status            *WebhookStatus  `json:"status,omitempty"`
		ReplacementObject json.RawMessage `json:"replacement_object,omitempty"`
	}

	var respData ResponseData
	if err := respEvent.DataAs(&respData); err != nil {
		return ActionDeny, nil, fmt.Errorf("failed to decode response event data: %w", err)
	}

	if !respData.Allowed {
		msg := "denied by webhook"
		if respData.Status != nil {
			msg = fmt.Sprintf("%s: %s", msg, respData.Status.Message)
		}
		return ActionDeny, nil, fmt.Errorf("%s", msg)
	}

	if respData.ReplacementObject != nil {
		newInputsMap := make(map[string]any)
		if err := json.Unmarshal(respData.ReplacementObject, &newInputsMap); err != nil {
			return ActionDeny, nil, fmt.Errorf("failed to unmarshal replacement inputs: %w", err)
		}
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
	logging.GetLogger().Info("ExecutePost called", "tool", req.ToolName)
	event := cloudevents.NewEvent()
	event.SetID(uuid.New().String())
	event.SetSource("https://github.com/mcpany/core")
	event.SetType("com.mcpany.tool.post_call")
	event.SetTime(time.Now())

	data := map[string]any{
		"kind":      configv1.WebhookKind_WEBHOOK_KIND_POST_CALL,
		"tool_name": req.ToolName,
		"result":    result,
	}
	if err := event.SetData(cloudevents.ApplicationJSON, data); err != nil {
		return nil, fmt.Errorf("failed to set cloud event data: %w", err)
	}

	respEvent, err := h.callWebhook(ctx, event)
	if err != nil {
		return nil, fmt.Errorf("webhook error: %w", err)
	}

	type ResponseData struct {
		Allowed           bool            `json:"allowed"`
		Status            *WebhookStatus  `json:"status,omitempty"`
		ReplacementObject json.RawMessage `json:"replacement_object,omitempty"`
	}

	var respData ResponseData
	if err := respEvent.DataAs(&respData); err != nil {
		return nil, fmt.Errorf("failed to decode response event data: %w", err)
	}

	if respData.ReplacementObject != nil {
		// If replacement object is present, return it.
		// We need to determine if we should return map or value.
		// Similar strategy: if original was string, try to extract "value" from replacement if it is map?
		// But CloudEvents data is cleaner.
		// If replacement object is just "some string" (quoted in JSON), Unmarshal handles it.
		// Since ReplacementObject is RawMessage, we can unmarshal it to any.
		var newResult any
		if err := json.Unmarshal(respData.ReplacementObject, &newResult); err != nil {
			return nil, fmt.Errorf("failed to unmarshal replacement result: %w", err)
		}

		// Unwrap "value" if it is the only key, to support returning primitives via Struct
		if m, ok := newResult.(map[string]any); ok {
			if v, ok := m["value"]; ok && len(m) == 1 {
				return v, nil
			}
		}

		return newResult, nil
	}

	return result, nil
}

func (h *WebhookHook) callWebhook(ctx context.Context, event cloudevents.Event) (*cloudevents.Event, error) {
	p, err := cehttp.New(
		cehttp.WithTarget(h.url),
		cehttp.WithClient(*h.client),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create protocol: %w", err)
	}

	c, err := cloudevents.NewClient(p, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	// We can manually add signature headers here if we modify the context or protocol wrapper?
	// The standard-webhooks library expects raw body. CloudEvents handles encoding.
	// To combine them, we might need to intercept the request.
	// Or we can assume CloudEvents has its own security or use standard headers.
	// The user asked to use standard-webhooks signature previously.
	// But standard-webhooks signs the *raw body*.
	// cloudevents-go handles body serialization internally.
	// We can use a middleware or just let cloudevents do its thing?
	// For now, let's stick to CloudEvents structure.
	// If we MUST use standard-webhooks signature, we need to capture the body.
	// Let's implement without explicit signature first, or rely on HTTPS/Auth header.
	// The user prompt mentioned "use standard-webhooks ... then format should follow cloudevents".
	// It's a bit mixed.
	// Standard webhooks usually implies a specific payload format too.
	// If we use CloudEvents, we are compliant with CloudEvents.
	// We can add a custom header extension `webhook-signature`?
	// Let's proceed with pure CloudEvents first as it is the "latest" instruction override.

	// Issue: We need to receive a response event!
	// cloudevents Request(ctx, event) returns (*Event, Result)

	respEvent, result := c.Request(ctx, event)
	if cloudevents.IsUndelivered(result) {
		return nil, fmt.Errorf("failed to send webhook event: %w", result)
	}

	if respEvent == nil {
		// Log the result which might contain HTTP status error
		logging.GetLogger().Error("No response event received", "result", result)
		return nil, fmt.Errorf("webhook error: no response event received (result: %v)", result)
	}
	// If status is not 2xx, result is error.

	if respEvent != nil {
		return respEvent, nil
	}

	// If no event returned (e.g. 202 Accepted or empty body 200), that's an issue for us if we expect review.
	// But maybe we allow it (no change).
	if result != nil && !cloudevents.IsACK(result) {
             // It might be an error result
             return nil, fmt.Errorf("request failed: %w", result)
	}

	// If empty response, assume allowed, no change?
	// Or should we synthesize a response?
	// Let's assume we REQUIRE a response event for now since it's a "Review" system.
	// But typical webhooks are fire-and-forget unless "pre-call".
	// Pre-call MUST reply.

	return nil, fmt.Errorf("no response event received")
}
