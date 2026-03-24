// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/mcpany/core/server/pkg/tool"
)

// ScopesConfig defines the configuration for capability-based scoping.
//
// Summary: Represents the configuration for capability-based role scoping.
type ScopesConfig struct {
	// Roles maps a role name to a list of allowed capability prefixes.
	Roles map[string][]string `json:"roles"`
}

// ScopesMiddleware enforces granular capability-based tokens for tool execution.
//
// Summary: Implements middleware that verifies if an active session role possesses the required scopes to call a tool.
type ScopesMiddleware struct {
	config ScopesConfig
}

// NewScopesMiddleware creates a new ScopesMiddleware.
//
// Summary: Initializes and returns a new capability-based scopes middleware instance.
//
// Parameters:
//   - config (ScopesConfig): The scoping configuration mapping roles to prefixes.
//
// Returns:
//   - *ScopesMiddleware: A new instance of the middleware.
func NewScopesMiddleware(config ScopesConfig) *ScopesMiddleware {
	return &ScopesMiddleware{
		config: config,
	}
}

type contextKey string

const agentRoleKey contextKey = "agent_role"

// Execute checks if the tool name matches any capability token prefix granted to the agent's role.
//
// Summary: Validates that the active session role has permission to execute the requested tool based on scope prefixes.
//
// Parameters:
//   - ctx (context.Context): The context of the request containing role information.
//   - req (*tool.ExecutionRequest): The details of the tool execution request.
//   - next (tool.ExecutionFunc): The next handler in the execution chain.
//
// Returns:
//   - any: The result of the next handler, typically the tool's response.
//   - error: An error if validation fails or the next handler errors out.
//
// Errors:
//   - Returns an error containing "access denied: role '[role]' missing capability for tool '[tool]'" if the active session is unauthorized.
func (m *ScopesMiddleware) Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error) {
	// For testing and mock purposes, we assume the agent role is passed in the context
	// or we default to a "default" role if not found.
	role := "default"
	if r, ok := ctx.Value(agentRoleKey).(string); ok && r != "" {
		role = r
	}

	allowedPrefixes, roleExists := m.config.Roles[role]
	if !roleExists {
		return nil, fmt.Errorf("access denied: no scope configuration for role '%s'", role)
	}

	isAllowed := false
	for _, prefix := range allowedPrefixes {
		if strings.HasPrefix(req.ToolName, prefix) {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		return nil, fmt.Errorf("access denied: tool '%s' is outside granted scopes", req.ToolName)
	}

	return next(ctx, req)
}
