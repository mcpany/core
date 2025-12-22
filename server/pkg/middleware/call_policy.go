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
	toolManager tool.ManagerInterface
}

// NewCallPolicyMiddleware creates a new CallPolicyMiddleware.
// toolManager is the toolManager.
// Returns the result.
func NewCallPolicyMiddleware(toolManager tool.ManagerInterface) *CallPolicyMiddleware {
	return &CallPolicyMiddleware{
		toolManager: toolManager,
	}
}

// Execute enforces call policies before proceeding to the next handler.
// ctx is the context.
// req is the req.
// Returns the result, an error.
func (m *CallPolicyMiddleware) Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error) {
	t, ok := m.toolManager.GetTool(req.ToolName)
	if !ok {
		// Tool not found, pass through (let other layers handle 404/error)
		return next(ctx, req)
	}

	serviceID := t.Tool().GetServiceId()
	serviceInfo, ok := m.toolManager.GetServiceInfo(serviceID)
	if !ok {
		// Service info not found, cannot enforce policies.
		// Defaulting to allow or maybe log warning?
		// Since we can't find policy, we assume allow (fail open) or deny (fail closed)?
		// Standard practice for missing config is usually allow if it's not a critical auth failure.
		// Given we found the tool but not service info, something is weird.
		// Proceeding.
		return next(ctx, req)
	}

	policies := serviceInfo.Config.GetCallPolicies()
	if len(policies) == 0 {
		return next(ctx, req)
	}

	// Convert arguments to JSON string for regex matching
	argsBytes, err := json.Marshal(req.ToolInputs)
	if err != nil {
		logging.GetLogger().Error("Failed to marshal tool inputs for policy check", "error", err)
		// If we can't check arguments, we should probably fail safe (deny) if policies exist?
		// Or maybe allow?
		// Let's block to be safe.
		return nil, fmt.Errorf("failed to process arguments for policy check")
	}
	argsStr := string(argsBytes)

	if err := m.checkPolicy(policies, req.ToolName, argsStr); err != nil {
		metrics.IncrCounterWithLabels([]string{"call_policy", "blocked_total"}, 1, []metrics.Label{
			{Name: "service_id", Value: serviceID},
			{Name: "tool_name", Value: req.ToolName},
		})
		return nil, err
	}

	return next(ctx, req)
}

func (m *CallPolicyMiddleware) checkPolicy(policies []*configv1.CallPolicy, toolName, argsStr string) error {
	for _, policy := range policies {
		if policy == nil {
			continue
		}
		matchedRule := false
		for _, rule := range policy.GetRules() {
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

			// We currently do not support CallIdRegex and UrlRegex here as we lack that context
			// in the generic ExecutionRequest easily (CallId is often part of name, URL is upstream specific).
			// If needed, we'd need to extend ExecutionRequest or look up ToolDefinition more deeply.
			// However, ArgumentRegex is the primary missing feature.

			if matched {
				matchedRule = true
				if rule.GetAction() == configv1.CallPolicy_DENY {
					return fmt.Errorf("execution denied by policy")
				}
				break // First match wins
			}
		}
		if !matchedRule {
			if policy.GetDefaultAction() == configv1.CallPolicy_DENY {
				return fmt.Errorf("execution denied by default policy")
			}
		}
	}
	return nil
}
