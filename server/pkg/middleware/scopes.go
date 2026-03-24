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
type ScopesConfig struct {
	// Roles maps a role name to a list of allowed capability prefixes.
	Roles map[string][]string `json:"roles"`
}

// ScopesMiddleware enforces granular capability-based tokens for tool execution.
type ScopesMiddleware struct {
	config ScopesConfig
}

// NewScopesMiddleware creates a new ScopesMiddleware.
func NewScopesMiddleware(config ScopesConfig) *ScopesMiddleware {
	return &ScopesMiddleware{
		config: config,
	}
}

const agentRoleKey contextKey = "agent_role"

// AgentRoleKey returns the context key used for the agent role.
func AgentRoleKey() any {
	return agentRoleKey
}

// CheckScope checks if the given scope is allowed for the agent's role.
func (m *ScopesMiddleware) CheckScope(ctx context.Context, scope string) error {
	role := "default"
	if r, ok := ctx.Value(agentRoleKey).(string); ok && r != "" {
		role = r
	}

	allowedPrefixes, roleExists := m.config.Roles[role]
	if !roleExists {
		return fmt.Errorf("access denied: no scope configuration for role '%s'", role)
	}

	isAllowed := false
	for _, prefix := range allowedPrefixes {
		if strings.HasPrefix(scope, prefix) {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		return fmt.Errorf("access denied: tool '%s' is outside granted scopes", scope)
	}

	return nil
}

// Execute checks if the tool name matches any capability token prefix granted to the agent's role.
func (m *ScopesMiddleware) Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error) {
	if err := m.CheckScope(ctx, req.ToolName); err != nil {
		return nil, err
	}
	return next(ctx, req)
}
