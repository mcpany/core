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
// based on tool name and arguments. It integrates with the tool manager to
// retrieve service-specific policy configurations.
type CallPolicyMiddleware struct {
	toolManager tool.ManagerInterface
}

// NewCallPolicyMiddleware creates a new CallPolicyMiddleware instance.
//
// Parameters:
//   - toolManager (tool.ManagerInterface): The tool manager interface used to look up tool and service information.
//
// Returns:
//   - (*CallPolicyMiddleware): A pointer to the initialized middleware.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewCallPolicyMiddleware(toolManager tool.ManagerInterface) *CallPolicyMiddleware {
	return &CallPolicyMiddleware{
		toolManager: toolManager,
	}
}

// Execute enforces call policies before proceeding to the next handler in the chain.
// It checks if the requested tool execution is permitted by the active service policies.
//
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - req (*tool.ExecutionRequest): The tool execution request containing tool name and arguments.
//   - next (tool.ExecutionFunc): The next handler to call if the policy allows execution.
//
// Returns:
//   - (any): The result of the tool execution from the next handler.
//   - (error): An error if the policy check fails, the service info is missing, or the execution is denied.
//
// Errors:
//   - Returns error if the service info cannot be found (fail closed).
//   - Returns error if policy evaluation fails.
//   - Returns "execution denied by policy" if the policy blocks the request.
//
// Side Effects:
//   - Increments "call_policy.blocked_total" metric if a request is blocked.
//   - Logs errors if service info is missing or policy evaluation fails.
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
