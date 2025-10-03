/*
 * Copyright 2025 Author(s) of MCPX
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package middleware_test

import (
	"context"
	"testing"

	"github.com/mcpxy/mcpx/pkg/auth"
	"github.com/mcpxy/mcpx/pkg/middleware"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware(t *testing.T) {
	t.Run("should call next handler", func(t *testing.T) {
		authManager := auth.NewAuthManager()
		mw := middleware.AuthMiddleware(authManager)

		var nextCalled bool
		nextHandler := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
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
