// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// LazyMCPConfig defines the configuration for Lazy-MCP on-demand tool discovery.
//
// Summary: Configuration for Lazy-MCP Middleware.
type LazyMCPConfig struct {
	Enabled   bool    `json:"enabled"`
	Threshold float64 `json:"threshold"`
	CacheTTL  string  `json:"cache_ttl"`
}

// LazyMCPMiddleware provides on-demand tool discovery mechanism, preventing context pollution.
//
// Summary: Middleware that filters tools dynamically.
type LazyMCPMiddleware struct {
	config LazyMCPConfig
}

// NewLazyMCPMiddleware creates a new LazyMCPMiddleware.
//
// Summary: Initializes a new LazyMCPMiddleware.
//
// Parameters:
//   - config: LazyMCPConfig. The configuration for the middleware.
//
// Returns:
//   - *LazyMCPMiddleware: The initialized middleware.
func NewLazyMCPMiddleware(config LazyMCPConfig) *LazyMCPMiddleware {
	return &LazyMCPMiddleware{
		config: config,
	}
}

// Execute checks if the request is tools/list and filters tools.
//
// Summary: Intercepts tool discovery.
func (m *LazyMCPMiddleware) Execute(ctx context.Context, method string, req mcp.Request, next mcp.MethodHandler) (mcp.Result, error) {
	if !m.config.Enabled || method != "tools/list" {
		return next(ctx, method, req)
	}

	result, err := next(ctx, method, req)
	if err != nil {
		return nil, err
	}

	if listResult, ok := result.(*mcp.ListToolsResult); ok {
		if m.config.Threshold > 0 {
			var filtered []*mcp.Tool
			for i, tool := range listResult.Tools {
				if float64(i) < float64(len(listResult.Tools))*m.config.Threshold {
					filtered = append(filtered, tool)
				}
			}
			listResult.Tools = filtered
		}
		return listResult, nil
	}

	return result, nil
}
