// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

func TestContextOptimizerMiddleware(t *testing.T) {
	opt := NewContextOptimizer(10)

	// Helper for next handler
	nextHandler := func(content string) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: content,
					},
				},
			}, nil
		}
	}

	t.Run("Long Response", func(t *testing.T) {
		handler := opt.Execute(nextHandler("This text is definitely longer than 10 characters"))

		res, err := handler(context.Background(), "tool/call", nil)
		assert.NoError(t, err)

		callRes, ok := res.(*mcp.CallToolResult)
		assert.True(t, ok)

		textContent, ok := callRes.Content[0].(*mcp.TextContent)
		assert.True(t, ok)

		assert.True(t, strings.Contains(textContent.Text, "TRUNCATED"), "Should contain truncated message")
		assert.Less(t, len(textContent.Text), 50, "Should be truncated")
	})

	t.Run("Short Response", func(t *testing.T) {
		handler := opt.Execute(nextHandler("Short"))

		res, err := handler(context.Background(), "tool/call", nil)
		assert.NoError(t, err)

		callRes, ok := res.(*mcp.CallToolResult)
		assert.True(t, ok)

		textContent, ok := callRes.Content[0].(*mcp.TextContent)
		assert.True(t, ok)

		assert.Equal(t, "Short", textContent.Text)
	})
}
