// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ContextOptimizer optimises the context size of responses.
type ContextOptimizer struct {
	MaxChars int
}

// NewContextOptimizer creates a new ContextOptimizer.
func NewContextOptimizer(maxChars int) *ContextOptimizer {
	return &ContextOptimizer{
		MaxChars: maxChars,
	}
}

// Execute is a middleware that truncates large text outputs in tool results.
func (co *ContextOptimizer) Execute(next mcp.MethodHandler) mcp.MethodHandler {
	return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		result, err := next(ctx, method, req)
		if err != nil {
			return result, err
		}

		if co.MaxChars <= 0 {
			return result, nil
		}

		// Check if result is CallToolResult
		if callResult, ok := result.(*mcp.CallToolResult); ok {
			for _, content := range callResult.Content {
				if textContent, ok := content.(*mcp.TextContent); ok {
					if len(textContent.Text) > co.MaxChars {
						originalLen := len(textContent.Text)
						textContent.Text = textContent.Text[:co.MaxChars] + fmt.Sprintf("...[TRUNCATED %d chars]", originalLen-co.MaxChars)
					}
				}
			}
		}

		return result, nil
	}
}
