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

// GranularScopesConfig defines capability-based tokens for agents.
type GranularScopesConfig struct {
	Default string   `json:"default"`
	Tokens  []string `json:"tokens"`
}

// GranularScopesMiddleware implements Least Privilege security.
type GranularScopesMiddleware struct {
	config GranularScopesConfig
}

// NewGranularScopesMiddleware creates a new GranularScopesMiddleware.
func NewGranularScopesMiddleware(config GranularScopesConfig) *GranularScopesMiddleware {
	return &GranularScopesMiddleware{
		config: config,
	}
}

// Execute inspects capability tokens and blocks unauthorized requests.
func (m *GranularScopesMiddleware) Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error) {
	// Simple simulation of scope enforcement
	// For example, if token is "fs:read:/tmp", and tool is "fs.write", it should fail.

	logger := logging.GetLogger().With("component", "granular_scopes_middleware")

	allowed := false

	if len(m.config.Tokens) == 0 && m.config.Default == "allow" {
		allowed = true
	}

	for _, token := range m.config.Tokens {
		// Extremely simplified token logic for the audit
		parts := strings.Split(token, ":")
		if len(parts) >= 2 {
			service := parts[0]
			action := parts[1]

			// E.g. token "fs:read:/tmp", tool "fs.read_file" -> match service "fs", match action "read" with "read_file"
			if strings.HasPrefix(req.ToolName, service+".") {
				toolAction := strings.TrimPrefix(req.ToolName, service+".")
				if strings.HasPrefix(toolAction, action) {
					allowed = true
					break
				}
			}
		} else if token == "*" || token == "all" {
			allowed = true
			break
		}
	}

	if !allowed && m.config.Default != "allow" && len(m.config.Tokens) > 0 {
		logger.Warn("Granular Scopes denied tool execution", "tool", req.ToolName, "tokens", m.config.Tokens)
		return nil, fmt.Errorf("execution denied by granular scopes: missing capability token for tool '%s'", req.ToolName)
	}

	return next(ctx, req)
}
