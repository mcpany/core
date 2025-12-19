// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/armon/go-metrics"
	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

// CallPolicyMiddleware is a middleware that enforces call policies (allow/deny)
// based on tool name and arguments.
type CallPolicyMiddleware struct {
	serviceID string
	policy    *configv1.CallPolicy
}

// NewCallPolicyMiddleware creates a new CallPolicyMiddleware.
func NewCallPolicyMiddleware(serviceID string, policy *configv1.CallPolicy) *CallPolicyMiddleware {
	return &CallPolicyMiddleware{
		serviceID: serviceID,
		policy:    policy,
	}
}

// Execute enforces call policies before proceeding to the next handler.
func (m *CallPolicyMiddleware) Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error) {
	if m.policy == nil {
		return next(ctx, req)
	}

	// Convert arguments to JSON string for regex matching
	argsBytes, err := json.Marshal(req.ToolInputs)
	if err != nil {
		logging.GetLogger().Error("Failed to marshal tool inputs for policy check", "error", err)
		return nil, fmt.Errorf("failed to process arguments for policy check")
	}
	argsStr := string(argsBytes)

	if err := m.checkPolicy(req.ToolName, argsStr); err != nil {
		metrics.IncrCounterWithLabels([]string{"call_policy", "blocked_total"}, 1, []metrics.Label{
			{Name: "service_id", Value: m.serviceID},
			{Name: "tool_name", Value: req.ToolName},
		})
		return nil, err
	}

	return next(ctx, req)
}

func (m *CallPolicyMiddleware) checkPolicy(toolName, argsStr string) error {
	matchedRule := false
	for _, rule := range m.policy.GetRules() {
		matched := true
		if rule.GetNameRegex() != "" {
			if matchedTool, _ := regexp.MatchString(rule.GetNameRegex(), toolName); !matchedTool {
				matched = false
			}
		}

		if matched && rule.GetArgumentRegex() != "" {
			if matchedArgs, _ := regexp.MatchString(rule.GetArgumentRegex(), argsStr); !matchedArgs {
				matched = false
			}
		}

		if matched {
			matchedRule = true
			if rule.GetAction() == configv1.CallPolicy_DENY {
				return fmt.Errorf("execution denied by policy")
			}
			break // First match wins
		}
	}
	if !matchedRule {
		if m.policy.GetDefaultAction() == configv1.CallPolicy_DENY {
			return fmt.Errorf("execution denied by default policy")
		}
	}
	return nil
}
