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
// compiledRule holds the pre-compiled regexes for a policy rule.
// NewPolicyHook creates a new PolicyHook with the given call policy.
//
// Summary: Initializes a new PolicyHook.
//
// Parameters:
//   - policy: *configv1.CallPolicy. The policy configuration to enforce.
//
// Returns:
//   - *PolicyHook: The initialized hook.
// Errors:
//   - None.
// Side Effects:
//   - None.
//
// Side Effects:
//   - Compiles regex patterns from the policy rules.
//   - Logs errors for invalid regexes.
// Errors:
//   - triggers relevant error states on failure.
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
// ExecutePre executes the policy check before a tool is called.
//
// Summary: Evaluates the tool request against the compiled policy rules.
//
// Parameters:
//   - _: context.Context. Unused.
//   - req: *ExecutionRequest. The tool execution request.
//
// Returns:
//   - Action: The action to take (Allow, Deny, etc.).
//   - *ExecutionRequest: The modified request (nil if no modification).
//   - error: An error if the policy denies execution.
//
// Errors:
//   - Returns error if an explicit DENY rule is matched.
//   - Returns error if the default policy is DENY and no ALLOW rule matches.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
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
// NewWebhookClient creates a new WebhookClient.
//
// Summary: Initializes a new WebhookClient.
//
// Parameters:
//   - config: *configv1.WebhookConfig. The webhook configuration.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
//
// Returns:
//   - *WebhookClient: The initialized client.
//
// Side Effects:
//   - Initializes HTTP client and optional signer.
// Errors:
//   - triggers relevant error states on failure.
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

// Call sends a cloud event to the webhook and returns the response event.
//
// Summary: Sends a synchronous CloudEvent to the webhook URL.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - eventType: string. The CloudEvent type.
//   - data: any. The event payload.
//
// Returns:
//   - *cloudevents.Event: The response CloudEvent.
//   - error: An error if the request fails or response is missing.
//
// Errors:
//   - Returns error if event creation or serialization fails.
//   - Returns error if the HTTP request fails or is undelivered.
//   - Returns error if no response event is received.
//
// Side Effects:
//   - Makes an external HTTP POST request.
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
// NewWebhookHook creates a new WebhookHook.
//
// Summary: Initializes a new WebhookHook.
//
// Parameters:
//   - config: *configv1.WebhookConfig. The webhook configuration.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
//
// Returns:
//   - *WebhookHook: The initialized hook.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func NewWebhookHook(config *configv1.WebhookConfig) *WebhookHook {
	return &WebhookHook{
		client: NewWebhookClient(config),
	}
}

// ExecutePre executes the webhook notification before a tool is called.
//
// Summary: Sends a pre-call event to the webhook and handles the response.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - req: *ExecutionRequest. The execution request.
//
// Returns:
//   - Action: Allow or Deny based on webhook response.
//   - *ExecutionRequest: Modified request if webhook returned replacements.
//   - error: An error if webhook denies or fails.
//
// Errors:
//   - Returns error if input marshaling fails.
//   - Returns error if webhook call fails.
//   - Returns error if webhook response parsing fails.
//   - Returns "denied by webhook" if explicitly denied.
//
// Side Effects:
//   - Invokes external webhook.
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

// ExecutePost executes the webhook notification after a tool is called.
//
// Summary: Sends a post-call event to the webhook and potentially modifies the result.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - req: *ExecutionRequest. The original request.
//   - result: any. The result of the tool execution.
//
// Returns:
//   - any: The (potentially modified) result.
//   - error: An error if the webhook call fails.
//
// Errors:
//   - Returns error if webhook call or response processing fails.
//
// Side Effects:
//   - Invokes external webhook.
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
// WebhookStatus represents the status returned by the webhook.
//
// Summary: Status information included in the webhook response.
// RoundTrip executes the HTTP request with a signature.
//
// Summary: Intercepts the request to add Webhook-Id, Webhook-Timestamp, and Webhook-Signature headers.
//
// Parameters:
//   - req: *http.Request. The outgoing request.
//
// Returns:
//   - *http.Response: The received response.
//   - error: An error if signing or transport fails.
// Errors:
//   - None.
// Side Effects:
//   - None.
//
// Side Effects:
//   - Reads and buffers the request body for signing.
//   - Modifies request headers.
// Errors:
//   - triggers relevant error states on failure.
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
