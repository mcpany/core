// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/tool"
)

// HITLConfig defines the configuration for Human-In-The-Loop approval flows.
//
// Summary: Configuration for HITL Middleware.
type HITLConfig struct {
	// Enabled determines if the HITL middleware is active.
	Enabled bool `json:"enabled"`
	// SensitiveTools is a list of tool names that require explicit human approval.
	SensitiveTools []string `json:"sensitive_tools"`
	// RequireMFA enforces Multi-Factor Authentication during the approval flow.
	RequireMFA bool `json:"require_mfa"`
}

// HITLMiddleware enforces Human-In-The-Loop approvals for sensitive actions.
//
// Summary: Middleware that enforces user approval flows for high-risk tools.
type HITLMiddleware struct {
	config HITLConfig
}

// NewHITLMiddleware creates a new HITLMiddleware.
//
// Summary: Initializes a new HITLMiddleware.
//
// Parameters:
//   - config: HITLConfig. The configuration for the HITL middleware.
//
// Returns:
//   - *HITLMiddleware: The initialized middleware.
func NewHITLMiddleware(config HITLConfig) *HITLMiddleware {
	return &HITLMiddleware{
		config: config,
	}
}

// Execute checks if the tool requires HITL approval before proceeding to the next handler.
//
// Summary: Checks if the tool execution requires human approval and suspends if necessary.
//
// Parameters:
//   - ctx: context.Context. The execution context.
//   - req: *tool.ExecutionRequest. The tool execution request.
//   - next: tool.ExecutionFunc. The next handler in the chain.
//
// Returns:
//   - any: The execution result if allowed.
//   - error: An error if the approval flow is denied or fails.
//
// Errors:
//   - Returns "execution suspended for HITL approval" if the action matches a sensitive tool and is suspended.
//
// Side Effects:
//   - Logs a warning when a sensitive tool is invoked.
func (m *HITLMiddleware) Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error) {
	if !m.config.Enabled {
		return next(ctx, req)
	}

	// Check if the current tool is in the sensitive list
	isSensitive := false
	for _, t := range m.config.SensitiveTools {
		// Exact match or prefix match (e.g. database.*)
		if t == req.ToolName || (strings.HasSuffix(t, ".*") && strings.HasPrefix(req.ToolName, strings.TrimSuffix(t, ".*"))) {
			isSensitive = true
			break
		}
	}

	if !isSensitive {
		return next(ctx, req)
	}

	logger := logging.GetLogger().With("component", "hitl_middleware")
	logger.Warn("HITL Middleware intercepted sensitive tool execution. Suspending for user approval.", "tool", req.ToolName, "require_mfa", m.config.RequireMFA)

	// Simulated Suspension Protocol
	// In a real environment, this would publish to a message bus and wait for an out-of-band response.
	return nil, fmt.Errorf("execution suspended for HITL approval: tool '%s' requires human confirmation", req.ToolName)
}
