// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"context"
	"testing"

	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCORSMiddleware(t *testing.T) {
	t.Run("should call next handler", func(t *testing.T) {
		mw := middleware.CORSMiddleware()

		var nextCalled bool
		nextHandler := func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
			nextCalled = true
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "test"},
				},
			}, nil
		}

		handler := mw(nextHandler)
		result, err := handler(context.Background(), "test.method", nil)
		require.NoError(t, err)
		assert.True(t, nextCalled, "next handler should be called")

		callToolResult, ok := result.(*mcp.CallToolResult)
		require.True(t, ok)

		require.Len(t, callToolResult.Content, 1)
		textContent, ok := callToolResult.Content[0].(*mcp.TextContent)
		require.True(t, ok)
		assert.Equal(t, "test", textContent.Text)
	})
}
