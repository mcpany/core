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
	"fmt"
	"regexp"

	"github.com/mcpany/core/pkg/logging"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

// PolicyHook implements PreCallHook using CallPolicy.
type PolicyHook struct {
	policy *configv1.CallPolicy
}

func NewPolicyHook(policy *configv1.CallPolicy) *PolicyHook {
	return &PolicyHook{policy: policy}
}

func (h *PolicyHook) ExecutePre(
	ctx context.Context,
	req *ExecutionRequest,
) (Action, *ExecutionRequest, error) {
	// Determine default action
	allowed := h.policy.GetDefaultAction() == configv1.CallPolicy_ALLOW

	for _, rule := range h.policy.GetRules() {
		// 1. Match Tool Name
		if rule.GetToolNameRegex() != "" {
			matched, err := regexp.MatchString(rule.GetToolNameRegex(), req.ToolName)
			if err != nil {
				logging.GetLogger().
					Error("Invalid tool name regex in policy", "regex", rule.GetToolNameRegex(), "error", err)
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

func NewTextTruncationHook(config *configv1.TextTruncationConfig) *TextTruncationHook {
	return &TextTruncationHook{maxChars: int(config.GetMaxChars())}
}

func (h *TextTruncationHook) ExecutePost(
	ctx context.Context,
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

// WebhookHook implements PreCallHook and PostCallHook?
// User requirement: "webhook... modify requests/responses".
// So it can be both.
type WebhookHook struct {
	config *configv1.WebhookConfig //nolint:unused
	// client http.Client
}

func (h *WebhookHook) ExecutePre(
	ctx context.Context,
	_ *ExecutionRequest,
) (Action, *ExecutionRequest, error) {
	// TODO: Implement webhook call
	return ActionAllow, nil, nil
}
