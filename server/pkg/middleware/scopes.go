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
// Summary: Represents the configuration for capability-based tool scoping.
type ScopesConfig struct {
	// Roles maps a role name to a list of allowed capability prefixes.
	Roles map[string][]string `json:"roles"`
}

// ScopesMiddleware enforces granular capability-based tokens for tool execution.
//
// Summary: Represents middleware that enforces tool execution scopes based on agent roles.
type ScopesMiddleware struct {
	config ScopesConfig
}

// NewScopesMiddleware creates a new ScopesMiddleware.
//
// Summary: Creates a new instance of ScopesMiddleware.
//
// Parameters:
//   - config (ScopesConfig): The configuration settings specifying allowed scopes per role.
//
// Returns:
//   - *ScopesMiddleware: A new instance of the middleware.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewScopesMiddleware(config ScopesConfig) *ScopesMiddleware {
	return &ScopesMiddleware{
		config: config,
	}
}

type contextKey string

const agentRoleKey contextKey = "agent_role"

// Execute checks if the tool name matches any capability token prefix granted to the agent's role.
//
// Summary: Validates that the requested tool is allowed for the agent's role before execution.
//
// Parameters:
//   - ctx (context.Context): The context for the request, optionally containing the agent's role.
//   - req (*tool.ExecutionRequest): The execution request detailing the tool to be called.
//   - next (tool.ExecutionFunc): The next execution function in the middleware chain.
//
// Returns:
//   - any: The result of the tool execution if permitted.
//   - error: An error if access is denied or execution fails.
//
// Errors:
//   - Returns an error if no scope configuration exists for the agent's role.
//   - Returns an error if the requested tool is outside the granted scopes.
//
// Side Effects:
//   - Executes the next function in the chain if access is granted.
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
	for _, token := range allowedPrefixes {
		parts := strings.SplitN(token, ":", 3)
		if len(parts) == 1 {
			// e.g. "fs:" -> allow any tool starting with "fs:"
			if strings.HasPrefix(req.ToolName, parts[0]) {
				isAllowed = true
				break
			}
		} else if len(parts) == 2 {
			// e.g. "fs:read"
			if req.ToolName == parts[0]+":"+parts[1] {
				isAllowed = true
				break
			}
		} else if len(parts) == 3 {
			// e.g. "fs:read:/tmp"
			toolNamePrefix := parts[0] + ":" + parts[1]
			if req.ToolName == toolNamePrefix {
				// Validate the path argument
				if req.Arguments != nil {
					// We heuristically check common path arguments: "path", "dir", "file"
					pathArgKeys := []string{"path", "dir", "file"}
					pathValue := ""
					for _, k := range pathArgKeys {
						if val, ok := req.Arguments[k]; ok {
							if sval, ok := val.(string); ok {
								pathValue = sval
								break
							}
						}
					}

					// If the tool has a path argument, it must start with the token's path.
					if pathValue != "" {
						if strings.HasPrefix(pathValue, parts[2]) {
							isAllowed = true
							break
						}
					}
					// If no path is provided but the token mandates a path constraint,
					// we MUST fail closed to prevent path-bypass attacks.
				}
			}
		}
	}

	if !isAllowed {
		return nil, fmt.Errorf("access denied: tool '%s' is outside granted scopes", req.ToolName)
	}

	return next(ctx, req)
}
