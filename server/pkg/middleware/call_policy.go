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

// CallPolicyMiddleware represents the public CallPolicyMiddleware entity.
//
// Summary: Defines the structured data model representing a policy middleware.
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
type CallPolicyMiddleware struct {
	toolManager tool.ManagerInterface
}

// NewCallPolicyMiddleware serves as a public interface for interacting with NewCallPolicyMiddleware.
//
// Summary: Constructs and returns an initialized call policy middleware ready for consumption.
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
func NewCallPolicyMiddleware(toolManager tool.ManagerInterface) *CallPolicyMiddleware {
	return &CallPolicyMiddleware{
		toolManager: toolManager,
	}
}

// Execute serves as a public interface for interacting with Execute.
//
// Summary: Execute the  appropriately based on current system conditions.
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
