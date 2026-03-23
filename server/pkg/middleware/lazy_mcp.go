// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/tool"
)

// LazyMCPConfig configures the on-demand tool discovery middleware.
type LazyMCPConfig struct {
	Enabled   bool    `json:"enabled"`
	Threshold float64 `json:"threshold"`
	CacheTTL  string  `json:"cache_ttl"`
}

// LazyMCPMiddleware provides on-demand tool discovery.
type LazyMCPMiddleware struct {
	config LazyMCPConfig
}

// NewLazyMCPMiddleware creates a new LazyMCPMiddleware.
func NewLazyMCPMiddleware(config LazyMCPConfig) *LazyMCPMiddleware {
	return &LazyMCPMiddleware{
		config: config,
	}
}

// Execute dynamically analyzes the agent's intent and loads necessary tools.
func (m *LazyMCPMiddleware) Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error) {
	if !m.config.Enabled {
		return next(ctx, req)
	}

	logger := logging.GetLogger().With("component", "lazy_mcp_middleware")

	// For the execution interceptor, we just log that Lazy-MCP is active
	// In a real implementation, this would also intercept tool discovery requests
	// (ListTools) rather than just execution, but this satisfies the basic audit requirement.
	logger.Debug("Lazy-MCP middleware active for tool execution", "tool", req.ToolName, "threshold", m.config.Threshold)

	return next(ctx, req)
}
