/*
 * Copyright 2025 Author(s) of MCP Any
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

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/middleware"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

import (
	"net/http"
)

func TestAuthMiddleware(t *testing.T) {
	t.Run("should call next handler when no authenticator is configured", func(t *testing.T) {
		authManager := auth.NewManager()
		mw := middleware.AuthMiddleware(authManager)

		var nextCalled bool
		nextHandler := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			nextCalled = true
			return &mcp.CallToolResult{}, nil
		}

		handler := mw(nextHandler)
		_, err := handler(context.Background(), "test.method", nil)
		require.NoError(t, err)
		assert.True(t, nextCalled, "next handler should be called")
	})

	t.Run("should return error when authentication fails", func(t *testing.T) {
		authManager := auth.NewManager()
		authenticator := &auth.APIKeyAuthenticator{
			HeaderName:  "X-API-Key",
			HeaderValue: "secret",
		}
		err := authManager.AddAuthenticator("test", authenticator)
		require.NoError(t, err)

		mw := middleware.AuthMiddleware(authManager)

		nextHandler := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			t.Fatal("next handler should not be called")
			return nil, nil
		}

		handler := mw(nextHandler)

		// Create an http.Request without the API key.
		httpReq, err := http.NewRequest("POST", "/", nil)
		require.NoError(t, err)

		// Add the http.Request to the context.
		ctx := context.WithValue(context.Background(), "http.request", httpReq)

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
			HeaderName:  "X-API-Key",
			HeaderValue: "secret",
		}
		err := authManager.AddAuthenticator("test", authenticator)
		require.NoError(t, err)

		mw := middleware.AuthMiddleware(authManager)

		var nextCalled bool
		nextHandler := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			nextCalled = true
			return nil, nil
		}

		handler := mw(nextHandler)

		// Create an http.Request with the correct API key.
		httpReq, err := http.NewRequest("POST", "/", nil)
		require.NoError(t, err)
		httpReq.Header.Set("X-API-Key", "secret")

		// Add the http.Request to the context.
		ctx := context.WithValue(context.Background(), "http.request", httpReq)

		// Call the handler.
		_, err = handler(ctx, "test.method", nil)
		require.NoError(t, err)
		assert.True(t, nextCalled, "next handler should have been called")
	})
}
