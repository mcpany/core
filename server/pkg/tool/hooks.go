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
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	webhook "github.com/standard-webhooks/standard-webhooks/libraries/go"
)

// compiledRule holds the pre-compiled regexes for a policy rule.
type compiledRule struct {
	nameRegex     *regexp.Regexp
	argumentRegex *regexp.Regexp
	rule          *configv1.CallPolicyRule
}

// PolicyHook represents the public PolicyHook entity.
//
// Summary: Defines the structured data model representing a hook.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type PolicyHook struct {
	policy        *configv1.CallPolicy
	compiledRules []compiledRule
}

// NewPolicyHook serves as a public interface for interacting with NewPolicyHook.
//
// Summary: Constructs and returns an initialized policy hook ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func NewPolicyHook(policy *configv1.CallPolicy) *PolicyHook {
	compiledRules := make([]compiledRule, len(policy.GetRules()))
	for i, rule := range policy.GetRules() {
		var nameRe, argRe *regexp.Regexp
		var err error

		if rule.GetNameRegex() != "" {
			nameRe, err = regexp.Compile(rule.GetNameRegex())
			if err != nil {
				logging.GetLogger().
					Error("Invalid tool name regex in policy", "regex", rule.GetNameRegex(), "error", err)
			}
		}

		if rule.GetArgumentRegex() != "" {
			argRe, err = regexp.Compile(rule.GetArgumentRegex())
			if err != nil {
				logging.GetLogger().
					Error("Invalid argument regex in policy", "regex", rule.GetArgumentRegex(), "error", err)
			}
		}

		compiledRules[i] = compiledRule{
			nameRegex:     nameRe,
			argumentRegex: argRe,
			rule:          rule,
		}
	}

	return &PolicyHook{
		policy:        policy,
		compiledRules: compiledRules,
	}
}

// ExecutePre serves as a public interface for interacting with ExecutePre.
//
// Summary: Execute the pre appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (h *PolicyHook) ExecutePre(
	_ context.Context,
	req *ExecutionRequest,
) (Action, *ExecutionRequest, error) {
	// Determine default action
	allowed := h.policy.GetDefaultAction() == configv1.CallPolicy_ALLOW

	for _, cRule := range h.compiledRules {
		rule := cRule.rule

		// 1. Match Tool Name
		if rule.GetNameRegex() != "" {
			if cRule.nameRegex == nil {
				continue // Skip invalid rule
			}
			if !cRule.nameRegex.MatchString(req.ToolName) {
				continue // Rule doesn't apply
			}
		}

		// 2. Match Arguments
		if rule.GetArgumentRegex() != "" {
			if cRule.argumentRegex == nil {
				continue
			}
			// req.ToolInputs is json.RawMessage ([]byte)
			if !cRule.argumentRegex.MatchString(string(req.ToolInputs)) {
				continue
			}
		}

		// Rule matched!
		switch rule.GetAction() {
		case configv1.CallPolicy_ALLOW:
			return ActionAllow, nil, nil
		case configv1.CallPolicy_SAVE_CACHE:
			return ActionSaveCache, nil, nil
		case configv1.CallPolicy_DELETE_CACHE:
			return ActionDeleteCache, nil, nil
		}
		return ActionDeny, nil, fmt.Errorf("tool execution denied by policy rule: %s", req.ToolName)
	}

	if allowed {
		return ActionAllow, nil, nil
	}
	return ActionDeny, nil, fmt.Errorf("tool execution denied by default policy: %s", req.ToolName)
}

// (Deprecated hooks removed)

// WebhookClient represents the public WebhookClient entity.
//
// Summary: Defines the structured data model representing a client.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type WebhookClient struct {
	url     string
	timeout time.Duration
	client  *http.Client
	webhook *webhook.Webhook
}

// NewWebhookClient serves as a public interface for interacting with NewWebhookClient.
//
// Summary: Constructs and returns an initialized webhook client ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func NewWebhookClient(config *configv1.WebhookConfig) *WebhookClient {
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

	return &WebhookClient{
		url:     config.GetUrl(),
		timeout: timeout,
		client:  client,
		webhook: wh,
	}
}

// Call serves as a public interface for interacting with Call.
//
// Summary: Call the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (c *WebhookClient) Call(ctx context.Context, eventType string, data any) (*cloudevents.Event, error) {
	event := cloudevents.NewEvent()
	event.SetID(uuid.New().String())
	event.SetSource("https://github.com/mcpany/core")
	event.SetType(eventType)
	event.SetTime(time.Now())

	if err := event.SetData(cloudevents.ApplicationJSON, data); err != nil {
		return nil, fmt.Errorf("failed to set cloud event data: %w", err)
	}

	p, err := cehttp.New(
		cehttp.WithTarget(c.url),
		cehttp.WithClient(*c.client),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create protocol: %w", err)
	}

	cl, err := cloudevents.NewClient(p, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	respEvent, result := cl.Request(ctx, event)
	if cloudevents.IsUndelivered(result) {
		return nil, fmt.Errorf("failed to send webhook event: %w", result)
	}

	if respEvent == nil {
		logging.GetLogger().Error("No response event received", "result", result)
		return nil, fmt.Errorf("webhook error: no response event received (result: %v)", result)
	}

	return respEvent, nil
}

// WebhookHook represents the public WebhookHook entity.
//
// Summary: Defines the structured data model representing a hook.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type WebhookHook struct {
	client *WebhookClient
}

// NewWebhookHook serves as a public interface for interacting with NewWebhookHook.
//
// Summary: Constructs and returns an initialized webhook hook ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func NewWebhookHook(config *configv1.WebhookConfig) *WebhookHook {
	return &WebhookHook{
		client: NewWebhookClient(config),
	}
}

// ExecutePre serves as a public interface for interacting with ExecutePre.
//
// Summary: Execute the pre appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

	data := map[string]any{
		"kind":      configv1.WebhookKind_WEBHOOK_KIND_PRE_CALL,
		"tool_name": req.ToolName,
		"inputs":    inputsMap,
	}

	respEvent, err := h.client.Call(ctx, "com.mcpany.tool.pre_call", data)
	if err != nil {
		return ActionDeny, nil, fmt.Errorf("webhook error: %w", err)
	}

	// ResponseData is a helper struct for parsing the webhook response.
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

// ExecutePost serves as a public interface for interacting with ExecutePost.
//
// Summary: Execute the post appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (h *WebhookHook) ExecutePost(
	ctx context.Context,
	req *ExecutionRequest,
	result any,
) (any, error) {
	logging.GetLogger().Info("ExecutePost called", "tool", req.ToolName)

	data := map[string]any{
		"kind":      configv1.WebhookKind_WEBHOOK_KIND_POST_CALL,
		"tool_name": req.ToolName,
		"result":    result,
	}

	respEvent, err := h.client.Call(ctx, "com.mcpany.tool.post_call", data)
	if err != nil {
		return nil, fmt.Errorf("webhook error: %w", err)
	}

	// ResponseData is a helper struct for parsing the webhook response.
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

// WebhookStatus represents the public WebhookStatus entity.
//
// Summary: Defines the structured data model representing a status.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type WebhookStatus struct {
	// Code is the status code returned by the webhook.
	Code int `json:"code"`
	// Message is a descriptive message returned by the webhook.
	Message string `json:"message"`
}

// SigningRoundTripper represents the public SigningRoundTripper entity.
//
// Summary: Defines the structured data model representing a round tripper.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type SigningRoundTripper struct {
	signer *webhook.Webhook
	base   http.RoundTripper
}

// RoundTrip serves as a public interface for interacting with RoundTrip.
//
// Summary: Round the trip appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *SigningRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if s.signer != nil {
		payload := []byte{} // Signing requires payload, but request body might be stream.

		if req.Body != nil {
			var err error
			payload, err = io.ReadAll(req.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read request body for signing: %w", err)
			}
			req.Body = io.NopCloser(bytes.NewReader(payload))
		}

		msgID := uuid.New().String()
		now := time.Now()
		signature, err := s.signer.Sign(msgID, now, payload)
		if err != nil {
			return nil, fmt.Errorf("failed to sign request: %w", err)
		}

		req.Header.Set("Webhook-Id", msgID)
		req.Header.Set("Webhook-Timestamp", fmt.Sprintf("%d", now.Unix()))
		req.Header.Set("Webhook-Signature", signature)
	}

	base := s.base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(req)
}
