// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/mcpany/core/server/pkg/tool"
)

// ScopesConfig represents the public ScopesConfig entity.
//
// Summary: Defines the structured data model representing a config.
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
type ScopesConfig struct {
	// Roles maps a role name to a list of allowed capability prefixes.
	Roles map[string][]string `json:"roles"`
}

// ScopesMiddleware represents the public ScopesMiddleware entity.
//
// Summary: Defines the structured data model representing a middleware.
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
type ScopesMiddleware struct {
	config ScopesConfig
}

// NewScopesMiddleware serves as a public interface for interacting with NewScopesMiddleware.
//
// Summary: Constructs and returns an initialized scopes middleware ready for consumption.
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
func NewScopesMiddleware(config ScopesConfig) *ScopesMiddleware {
	return &ScopesMiddleware{
		config: config,
	}
}


const agentRoleKey contextKey = "agent_role"

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
