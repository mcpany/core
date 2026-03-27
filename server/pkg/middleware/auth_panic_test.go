// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"context"
	"testing"

	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/consts"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

// TestAuthMiddleware_TypedNil_Panic verifies that the middleware does not panic
// when a typed nil interface is passed.
// This test ensures robustness against nil pointers.
func TestAuthMiddleware_TypedNil_Panic(t *testing.T) {
	authManager := auth.NewManager()
	authManager.SetAPIKey("global-secret") // Setup global key to avoid early exit on extraction failure

	mw := middleware.AuthMiddleware(authManager)

	nextHandler := func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{}, nil
	}

	handler := mw(nextHandler)

	t.Run("tools/call with typed nil request", func(t *testing.T) {
		var req *mcp.CallToolRequest = nil // Typed nil
		var iReq mcp.Request = req

		// This should NOT panic
		assert.NotPanics(t, func() {
			_, _ = handler(context.Background(), consts.MethodToolsCall, iReq)
		})
	})

	t.Run("prompts/get with typed nil request", func(t *testing.T) {
		var req *mcp.GetPromptRequest = nil // Typed nil
		var iReq mcp.Request = req

		// This should NOT panic
		assert.NotPanics(t, func() {
			_, _ = handler(context.Background(), consts.MethodPromptsGet, iReq)
		})
	})
}
