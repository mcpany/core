// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"fmt"

	"github.com/armon/go-metrics"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/tool"
)

// CallPolicyMiddleware is a middleware that enforces call policies (allow/deny)
// based on tool name and arguments.
type CallPolicyMiddleware struct {
	toolManager tool.ManagerInterface
}

// NewCallPolicyMiddleware creates a new CallPolicyMiddleware.
//
// Summary: Initializes the call policy enforcement middleware.
//
// Parameters:
//   - toolManager (tool.ManagerInterface): The tool manager to retrieve tool and service info.
//
// Returns:
//   - (*CallPolicyMiddleware): The initialized middleware.
func NewCallPolicyMiddleware(toolManager tool.ManagerInterface) *CallPolicyMiddleware {
	return &CallPolicyMiddleware{
		toolManager: toolManager,
	}
}

// Execute enforces call policies before proceeding to the next handler.
//
// Summary: Checks if the tool execution is allowed by the configured policies.
//
// It evaluates CompiledCallPolicies associated with the service. If any policy denies
// the execution based on the tool name or arguments, the request is rejected.
//
// Parameters:
//   - ctx (context.Context): The request context.
//   - req (*tool.ExecutionRequest): The execution request containing tool name and inputs.
//   - next (tool.ExecutionFunc): The next handler in the chain.
//
// Returns:
//   - (any): The result of the next handler if allowed.
//   - (error): An error if the tool/service is not found, or if execution is denied.
//
// Side Effects:
//   - Logs blocked requests.
//   - Increments metrics for blocked requests.
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
		// We must fail closed to prevent policy bypass.
		logging.GetLogger().Error("Service info not found for tool execution", "service_id", serviceID, "tool_name", req.ToolName)
		return nil, fmt.Errorf("service info not found for service %s", serviceID)
	}

	compiledPolicies := serviceInfo.CompiledPolicies
	if len(compiledPolicies) == 0 {
		return next(ctx, req)
	}

	// For CompiledCallPolicy, we need arguments as []byte (req.ToolInputs is already json.RawMessage which is []byte)
	allowed, err := tool.EvaluateCompiledCallPolicy(compiledPolicies, req.ToolName, "", req.ToolInputs)
	if err != nil {
		logging.GetLogger().Error("Failed to evaluate call policy", "error", err)
		return nil, err
	}

	if !allowed {
		metrics.IncrCounterWithLabels([]string{"call_policy", "blocked_total"}, 1, []metrics.Label{
			{Name: "service_id", Value: serviceID},
			{Name: "tool_name", Value: req.ToolName},
		})
		return nil, fmt.Errorf("execution denied by policy")
	}

	return next(ctx, req)
}
