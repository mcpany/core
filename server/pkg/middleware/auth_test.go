// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"context"
	"testing"

	"encoding/json"
	"net/http"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/consts"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware(t *testing.T) {
	t.Run("should return error when no authenticator is configured", func(t *testing.T) {
		authManager := auth.NewManager()
		mw := middleware.AuthMiddleware(authManager)

		nextHandler := func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
			t.Fatal("next handler should not be called")
			return nil, nil
		}

		handler := mw(nextHandler)

		// Create an http.Request
		httpReq, err := http.NewRequest("POST", "/", nil)
		require.NoError(t, err)

		// Add the http.Request to the context
		ctx := context.WithValue(context.Background(), middleware.HTTPRequestContextKey, httpReq)

		_, err = handler(ctx, "test.method", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})

	t.Run("should return error when authentication fails", func(t *testing.T) {
		authManager := auth.NewManager()
		authenticator := &auth.APIKeyAuthenticator{
			ParamName: "X-API-Key",
			Value:     "secret",
			In:        configv1.APIKeyAuth_HEADER,
		}
		err := authManager.AddAuthenticator("test", authenticator)
		require.NoError(t, err)

		mw := middleware.AuthMiddleware(authManager)

		nextHandler := func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
			t.Fatal("next handler should not be called")
			return nil, nil
		}

		handler := mw(nextHandler)

		// Create an http.Request without the API key.
		httpReq, err := http.NewRequest("POST", "/", nil)
		require.NoError(t, err)

		// Add the http.Request to the context.
		ctx := context.WithValue(context.Background(), middleware.HTTPRequestContextKey, httpReq)

		// Call the handler. The method "test.method" implies the serviceID is "test".
		_, err = handler(ctx, "test.method", nil)

		// This is the point of failure for the existing buggy middleware.
		// It should return an error, but it will return nil.
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})

	t.Run("should call next handler when authentication succeeds", func(t *testing.T) {
		authManager := auth.NewManager()
		authenticator := &auth.APIKeyAuthenticator{
			ParamName: "X-API-Key",
			Value:     "secret",
			In:        configv1.APIKeyAuth_HEADER,
		}
		err := authManager.AddAuthenticator("test", authenticator)
		require.NoError(t, err)

		mw := middleware.AuthMiddleware(authManager)

		var nextCalled bool
		nextHandler := func(ctx context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
			nextCalled = true
			// Verify context has API key
			key, ok := auth.APIKeyFromContext(ctx)
			assert.True(t, ok, "API key should be in context")
			assert.Equal(t, "secret", key)
			return nil, nil
		}

		handler := mw(nextHandler)

		// Create an http.Request with the correct API key.
		httpReq, err := http.NewRequest("POST", "/", nil)
		require.NoError(t, err)
		httpReq.Header.Set("X-API-Key", "secret")

		// Add the http.Request to the context.
		ctx := context.WithValue(context.Background(), middleware.HTTPRequestContextKey, httpReq)

		// Call the handler.
		_, err = handler(ctx, "test.method", nil)
		require.NoError(t, err)
		assert.True(t, nextCalled, "next handler should have been called")
	})

	t.Run("should return error when method format is invalid and no global auth", func(t *testing.T) {
		authManager := auth.NewManager()
		mw := middleware.AuthMiddleware(authManager)

		nextHandler := func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
			t.Fatal("next handler should not be called")
			return nil, nil
		}

		handler := mw(nextHandler)

		httpReq, err := http.NewRequest("POST", "/", nil)
		require.NoError(t, err)
		ctx := context.WithValue(context.Background(), middleware.HTTPRequestContextKey, httpReq)

		// Method without dot
		_, err = handler(ctx, "invalid_method", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})

	t.Run("should bypass auth when http request is missing from context (stdio)", func(t *testing.T) {
		authManager := auth.NewManager()
		// Add an authenticator so it tries to authenticate
		authenticator := &auth.APIKeyAuthenticator{
			ParamName: "X-API-Key",
			Value:     "secret",
			In:        configv1.APIKeyAuth_HEADER,
		}
		err := authManager.AddAuthenticator("test", authenticator)
		require.NoError(t, err)

		mw := middleware.AuthMiddleware(authManager)

		nextCalled := false
		handler := mw(func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
			nextCalled = true
			return nil, nil
		})

		// Context without middleware.HTTPRequestContextKey
		_, err = handler(context.Background(), "test.method", nil)
		require.NoError(t, err)
		assert.True(t, nextCalled)
	})

	t.Run("should enforce auth for tools/call with protected service prefix", func(t *testing.T) {
		authManager := auth.NewManager()
		authenticator := &auth.APIKeyAuthenticator{
			ParamName: "X-API-Key",
			Value:     "secret-value",
			In:        configv1.APIKeyAuth_HEADER,
		}
		err := authManager.AddAuthenticator("protected", authenticator)
		require.NoError(t, err)

		mw := middleware.AuthMiddleware(authManager)

		// Mock next handler
		var nextCalled bool
		nextHandler := func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
			nextCalled = true
			return &mcp.CallToolResult{}, nil
		}

		handler := mw(nextHandler)

		// Create a "tools/call" request for a tool in the "protected" service
		toolName := "protected.my-tool"
		req := &mcp.CallToolRequest{
			Params: &mcp.CallToolParamsRaw{
				Name:      toolName,
				Arguments: json.RawMessage(`{}`),
			},
		}

		// Create an http.Request WITHOUT credentials
		httpReq, err := http.NewRequest("POST", "/", nil) // No X-API-Key
		require.NoError(t, err)

		ctx := context.WithValue(context.Background(), middleware.HTTPRequestContextKey, httpReq)

		// Call the handler with "tools/call" method
		_, err = handler(ctx, consts.MethodToolsCall, req)

		// EXPECTATION: Should FAIL with Unauthorized
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
		assert.False(t, nextCalled)
	})

	t.Run("should block when global auth fails for public-like method", func(t *testing.T) {
		authManager := auth.NewManager()
		authManager.SetAPIKey("global-secret") // Set global key

		mw := middleware.AuthMiddleware(authManager)

		nextHandler := func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
			t.Fatal("next handler should not be called")
			return nil, nil
		}

		handler := mw(nextHandler)

		httpReq, err := http.NewRequest("POST", "/", nil) // No X-API-Key
		require.NoError(t, err)
		ctx := context.WithValue(context.Background(), middleware.HTTPRequestContextKey, httpReq)

		// Method without dot (serviceID == "")
		_, err = handler(ctx, "invalid_method", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})

	t.Run("should allow when global auth succeeds for public-like method", func(t *testing.T) {
		authManager := auth.NewManager()
		authManager.SetAPIKey("global-secret") // Set global key

		mw := middleware.AuthMiddleware(authManager)

		var nextCalled bool
		nextHandler := func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
			nextCalled = true
			return nil, nil
		}

		handler := mw(nextHandler)

		httpReq, err := http.NewRequest("POST", "/", nil)
		require.NoError(t, err)
		httpReq.Header.Set("X-API-Key", "global-secret")
		ctx := context.WithValue(context.Background(), middleware.HTTPRequestContextKey, httpReq)

		// Method without dot (serviceID == "")
		_, err = handler(ctx, "invalid_method", nil)
		require.NoError(t, err)
		assert.True(t, nextCalled)
	})
}
