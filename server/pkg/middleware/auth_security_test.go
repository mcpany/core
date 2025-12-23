// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/middleware"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGlobalAPIKeyProtection(t *testing.T) {
	t.Run("should enforce global API key even when no service authenticator is present", func(t *testing.T) {
		// 1. Setup Auth Manager with Global API Key
		authManager := auth.NewManager()
		authManager.SetAPIKey("super-secret-global-key")

		// 2. Setup Middleware
		mw := middleware.AuthMiddleware(authManager)

		// 3. Mock Next Handler (should NOT be reached if auth fails)
		var nextCalled bool
		nextHandler := func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
			nextCalled = true
			return &mcp.CallToolResult{}, nil
		}
		handler := mw(nextHandler)

		// 4. Create Request WITHOUT API Key
		httpReq, err := http.NewRequest("POST", "/", nil)
		require.NoError(t, err)

		// "http.request" key as per convention in codebase
		ctx := context.WithValue(context.Background(), "http.request", httpReq)

		// 5. Invoke Handler for a random service "myservice.method"
		// This service has NO specific authenticator registered.
		_, err = handler(ctx, "myservice.method", nil)

		// 6. Assertions
		// Expect "unauthorized: missing API key" because global key is enforced
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
		assert.False(t, nextCalled, "Next handler should not be called")
	})

	t.Run("should allow request when global API key is provided and valid", func(t *testing.T) {
		// 1. Setup Auth Manager with Global API Key
		authManager := auth.NewManager()
		authManager.SetAPIKey("super-secret-global-key")

		// 2. Setup Middleware
		mw := middleware.AuthMiddleware(authManager)

		// 3. Mock Next Handler
		var nextCalled bool
		nextHandler := func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
			nextCalled = true
			return &mcp.CallToolResult{}, nil
		}
		handler := mw(nextHandler)

		// 4. Create Request WITH API Key
		httpReq, err := http.NewRequest("POST", "/", nil)
		require.NoError(t, err)
		httpReq.Header.Set("X-API-Key", "super-secret-global-key")

		// "http.request" key as per convention in codebase
		ctx := context.WithValue(context.Background(), "http.request", httpReq)

		// 5. Invoke Handler
		_, err = handler(ctx, "myservice.method", nil)

		// 6. Assertions
		require.NoError(t, err)
		assert.True(t, nextCalled, "Next handler should be called")
	})
}
