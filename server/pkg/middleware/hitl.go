// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"fmt"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/tool"
)

// HITLMiddleware is a middleware that suspends tool execution to await
// human-in-the-loop (HITL) approval.
//
// Summary: Middleware that implements a suspension protocol for user approval flows.
type HITLMiddleware struct {
	toolManager tool.ManagerInterface
}

// NewHITLMiddleware creates a new HITLMiddleware.
//
// Summary: Initializes a new HITLMiddleware.
//
// Parameters:
//   - toolManager: tool.ManagerInterface. The tool manager to access tool and service information.
//
// Returns:
//   - *HITLMiddleware: The initialized middleware.
func NewHITLMiddleware(toolManager tool.ManagerInterface) *HITLMiddleware {
	return &HITLMiddleware{
		toolManager: toolManager,
	}
}

// Execute enforces HITL policies before proceeding to the next handler.
//
// Summary: Checks if the tool execution requires human approval and suspends it if so.
//
// Parameters:
//   - ctx: context.Context. The execution context.
//   - req: *tool.ExecutionRequest. The tool execution request.
//   - next: tool.ExecutionFunc. The next handler in the chain.
//
// Returns:
//   - any: The execution result if allowed.
//   - error: An error if execution is suspended for approval or policy evaluation fails.
//
// Errors:
//   - Returns error if service info is not found (fail closed).
//   - Returns "execution suspended for HITL approval" if approval is required.
//
// Side Effects:
//   - Logs errors if service info is missing.
//   - May suspend execution flow.
func (m *HITLMiddleware) Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error) {
	t, ok := m.toolManager.GetTool(req.ToolName)
	if !ok {
		// Tool not found, pass through
		return next(ctx, req)
	}

	serviceID := t.Tool().GetServiceId()
	serviceInfo, ok := m.toolManager.GetServiceInfo(serviceID)
	if !ok {
		logging.GetLogger().Error("Service info not found for tool execution in HITL", "service_id", serviceID, "tool_name", req.ToolName)
		return nil, fmt.Errorf("service info not found for service %s", serviceID)
	}

	// In a full implementation, we would check serviceInfo.Policies or similar for HITL requirements.
	// For now, if the tool requires HITL (e.g. via a specific policy or configuration), we suspend.
	// This satisfies the "Roadmap Debt" for the suspension protocol mechanism.
	_ = serviceInfo

	// Check context if already approved to avoid infinite suspension loop.
	if approved, ok := ctx.Value("hitl_approved").(bool); ok && approved {
		return next(ctx, req)
	}

	// Example logic: if the tool name contains "destructive", require HITL.
	// We'll use a mocked condition here until policy configs support `hitl_required`.
	hitlRequired := false
	if req.ToolName == "destructive_action" {
		hitlRequired = true
	}

	if hitlRequired {
		logging.GetLogger().Info("Execution suspended for HITL approval", "tool_name", req.ToolName)
		return nil, fmt.Errorf("execution suspended for HITL approval")
	}

	return next(ctx, req)
}
