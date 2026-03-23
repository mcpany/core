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

// GranularScopesConfig defines the configuration for granular capability tokens.
//
// Summary: Configuration for Granular Scopes Middleware.
type GranularScopesConfig struct {
	// Default policy if no specific token matches (e.g., "read", "deny")
	Default string `json:"default"`
	// Tokens is a list of allowed capabilities (e.g., "fs:read:/tmp", "db:write:users")
	Tokens []string `json:"tokens"`
}

// GranularScopesMiddleware enforces token-based scoping for tool execution.
//
// Summary: Middleware that enforces granular capability tokens.
type GranularScopesMiddleware struct {
	config GranularScopesConfig
}

// NewGranularScopesMiddleware creates a new GranularScopesMiddleware.
//
// Summary: Initializes a new GranularScopesMiddleware.
//
// Parameters:
//   - config: GranularScopesConfig. The configuration for the middleware.
//
// Returns:
//   - *GranularScopesMiddleware: The initialized middleware.
func NewGranularScopesMiddleware(config GranularScopesConfig) *GranularScopesMiddleware {
	return &GranularScopesMiddleware{
		config: config,
	}
}

// Execute checks if the tool requires specific capabilities and whether the agent has them.
//
// Summary: Checks capability tokens before execution.
func (m *GranularScopesMiddleware) Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error) {
	logger := logging.GetLogger().With("component", "granular_scopes_middleware")

	// Determine required capabilities based on the tool name
	toolScope := strings.ReplaceAll(req.ToolName, ".", ":")

	allowed := false
	for _, token := range m.config.Tokens {
		if strings.HasPrefix(toolScope, token) || strings.HasPrefix(token, toolScope) {
			allowed = true
			break
		}
	}

	if !allowed && strings.ToLower(m.config.Default) != "allow" {
		if strings.ToLower(m.config.Default) == "read" && strings.Contains(strings.ToLower(req.ToolName), "read") {
			allowed = true
		} else {
			logger.Warn("Granular Scopes blocked execution", "tool", req.ToolName, "required_scope", toolScope)
			return nil, fmt.Errorf("access denied: lacking required capability token for tool '%s'", req.ToolName)
		}
	}

	return next(ctx, req)
}
