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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"

	"github.com/mcpany/core/pkg/logging"
	configv1 "github.com/mcpany/core/proto/config/v1"
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

// TextTruncationHook implements PostCallHook.
type TextTruncationHook struct {
	maxChars int
}

// NewTextTruncationHook creates a new TextTruncationHook with the given configuration.
func NewTextTruncationHook(config *configv1.TextTruncationConfig) *TextTruncationHook {
	return &TextTruncationHook{maxChars: int(config.GetMaxChars())}
}

// ExecutePost executes the text truncation logic after a tool is called.
func (h *TextTruncationHook) ExecutePost(
	_ context.Context,
	_ *ExecutionRequest,
	result any,
) (any, error) {
	if h.maxChars <= 0 {
		return result, nil
	}

	// Handle string result
	if str, ok := result.(string); ok {
		if len(str) > h.maxChars {
			return str[:h.maxChars] + "...", nil
		}
		return str, nil
	}

	// Handle map result (common for JSON)
	if m, ok := result.(map[string]any); ok {
		// Traverse and truncate suitable string fields?
		// For now, let's just serialize to check size?
		// Or maybe user implies text response truncation.
		// If the result is a Map, we probably shouldn't blindly truncate.
		// But if there is a "content" or "text" field?
		// User requirement just said "text modify hook".
		// I will implement a recursive truncation for string values in maps.
		return h.truncateMap(m), nil
	}

	return result, nil
}

func (h *TextTruncationHook) truncateMap(m map[string]any) map[string]any {
	newMap := make(map[string]any, len(m))
	for k, v := range m {
		switch val := v.(type) {
		case string:
			if len(val) > h.maxChars {
				newMap[k] = val[:h.maxChars] + "..."
			} else {
				newMap[k] = val
			}
		case map[string]any:
			newMap[k] = h.truncateMap(val)
		default:
			newMap[k] = val
		}
	}
	return newMap
}

// HtmlToMarkdownHook implements PostCallHook to convert HTML to Markdown.
type HtmlToMarkdownHook struct {
	converter *md.Converter
}

// NewHtmlToMarkdownHook creates a new HtmlToMarkdownHook.
func NewHtmlToMarkdownHook(config *configv1.HtmlToMarkdownConfig) *HtmlToMarkdownHook {
	converter := md.NewConverter("", true, nil)
	// We can configure excludes/includes if needed based on config
	return &HtmlToMarkdownHook{converter: converter}
}

// ExecutePost converts HTML strings in the result to Markdown.
func (h *HtmlToMarkdownHook) ExecutePost(
	_ context.Context,
	_ *ExecutionRequest,
	result any,
) (any, error) {
	// Handle string result
	if str, ok := result.(string); ok {
		return h.convertString(str), nil
	}

	// Handle map result
	if m, ok := result.(map[string]any); ok {
		return h.convertMap(m), nil
	}

	return result, nil
}

func (h *HtmlToMarkdownHook) convertString(s string) string {
	// Check if it looks like HTML? For now, we assume if this hook is enabled,
	// the user expects conversion.
	markdown, err := h.converter.ConvertString(s)
	if err != nil {
		logging.GetLogger().Warn("Failed to convert HTML to Markdown", "error", err)
		return s // return original if failed
	}
	return markdown
}

func (h *HtmlToMarkdownHook) convertMap(m map[string]any) map[string]any {
	newMap := make(map[string]any, len(m))
	for k, v := range m {
		switch val := v.(type) {
		case string:
			newMap[k] = h.convertString(val)
		case map[string]any:
			newMap[k] = h.convertMap(val)
		default:
			newMap[k] = val
		}
	}
	return newMap
}

// WebhookHook implements PreCallHook and PostCallHook?
// User requirement: "webhook... modify requests/responses".
// So it can be both.
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

type webhookRequest struct {
	HookType string          `json:"hook_type"`
	ToolName string          `json:"tool_name"`
	Inputs   json.RawMessage `json:"inputs,omitempty"`
	Result   any             `json:"result,omitempty"`
}

type webhookResponse struct {
	Action string          `json:"action"` // "allow", "deny"
	Error  string          `json:"error,omitempty"`
	Inputs json.RawMessage `json:"inputs,omitempty"`
	Result any             `json:"result,omitempty"`
}

// ExecutePre executes the webhook notification before a tool is called.
func (h *WebhookHook) ExecutePre(
	ctx context.Context,
	req *ExecutionRequest,
) (Action, *ExecutionRequest, error) {
	payload := webhookRequest{
		HookType: "pre",
		ToolName: req.ToolName,
		Inputs:   req.ToolInputs,
	}

	resp, err := h.callWebhook(ctx, payload)
	if err != nil {
		return ActionDeny, nil, fmt.Errorf("webhook error: %w", err)
	}

	if resp.Action == "deny" {
		return ActionDeny, nil, fmt.Errorf("denied by webhook: %s", resp.Error)
	}

	if resp.Inputs != nil {
		// Modify inputs
		newReq := *req
		newReq.ToolInputs = resp.Inputs
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
	payload := webhookRequest{
		HookType: "post",
		ToolName: req.ToolName,
		Result:   result,
	}

	resp, err := h.callWebhook(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("webhook error: %w", err)
	}

	if resp.Result != nil {
		return resp.Result, nil
	}

	return result, nil
}

func (h *WebhookHook) callWebhook(ctx context.Context, payload webhookRequest) (*webhookResponse, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
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

	var resp webhookResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &resp, nil
}
