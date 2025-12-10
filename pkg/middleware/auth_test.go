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
	"net/http"
	"testing"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/middleware"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	authManager := auth.NewAuthManager()
	handler := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{}, nil
	}

	t.Run("valid api key", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set("X-API-Key", "test-api-key")
		ctx := context.WithValue(context.Background(), "http.request", req)
		authManager.SetAPIKey("test-api-key")
		_, err := middleware.AuthMiddleware(authManager)(handler)(ctx, "test.method", nil)
		assert.NoError(t, err)
	})

	t.Run("invalid api key", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set("X-API-Key", "invalid-api-key")
		ctx := context.WithValue(context.Background(), "http.request", req)
		authManager.SetAPIKey("test-api-key")
		_, err := middleware.AuthMiddleware(authManager)(handler)(ctx, "test.method", nil)
		assert.Error(t, err)
	})

	t.Run("missing api key", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)
		ctx := context.WithValue(context.Background(), "http.request", req)
		authManager.SetAPIKey("test-api-key")
		_, err := middleware.AuthMiddleware(authManager)(handler)(ctx, "test.method", nil)
		assert.Error(t, err)
	})

	t.Run("no api key configured", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)
		ctx := context.WithValue(context.Background(), "http.request", req)
		authManager.SetAPIKey("")
		_, err := middleware.AuthMiddleware(authManager)(handler)(ctx, "test.method", nil)
		assert.NoError(t, err)
	})
}
