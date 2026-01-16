// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"fmt"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ContextOptimizer optimises the context size of responses.
type ContextOptimizer struct {
	MaxChars int
}

// NewContextOptimizer creates a new ContextOptimizer.
func NewContextOptimizer(config *configv1.ContextOptimizerConfig) *ContextOptimizer {
	maxChars := 1000 // Default
	if config != nil && config.GetMaxChars() > 0 {
		maxChars = int(config.GetMaxChars())
	}
	return &ContextOptimizer{
		MaxChars: maxChars,
	}
}

// UpdateConfig updates the configuration.
func (co *ContextOptimizer) UpdateConfig(config *configv1.ContextOptimizerConfig) {
	if config != nil && config.GetMaxChars() > 0 {
		co.MaxChars = int(config.GetMaxChars())
	}
}

// Middleware returns the MCP middleware handler.
func (co *ContextOptimizer) Middleware(next mcp.MethodHandler) mcp.MethodHandler {
	return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		res, err := next(ctx, method, req)
		if err != nil {
			return res, err
		}

		if ctr, ok := res.(*mcp.CallToolResult); ok {
			for _, content := range ctr.Content {
				if tc, ok := content.(*mcp.TextContent); ok {
					if len(tc.Text) > co.MaxChars {
						tc.Text = tc.Text[:co.MaxChars] + fmt.Sprintf("...[TRUNCATED %d chars]", len(tc.Text)-co.MaxChars)
					}
				}
			}
		}

		return res, nil
	}
}
