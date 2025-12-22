// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

func ptrTo[T any](v T) *T {
	return &v
}

func TestRegistry_HTTP(t *testing.T) {
	// Register a dummy middleware
	Register("test_http", func(cfg *configv1.Middleware) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Test", cfg.GetName())
				next.ServeHTTP(w, r)
			})
		}
	})

	// Configure it
	configs := []*configv1.Middleware{
		{
			Name:     ptrTo("test_http"),
			Priority: ptrTo(int32(10)),
		},
		{
			Name: ptrTo("non_existent"), // Should be ignored
		},
		{
			Name:     ptrTo("disabled_one"),
			Disabled: ptrTo(true),
		},
	}
	// Register disabled one to ensure it exists but is disabled
	Register("disabled_one", func(cfg *configv1.Middleware) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler { return next }
	})

	middlewares := GetHTTPMiddlewares(configs)
	assert.Len(t, middlewares, 1)

	// Verify execution
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	wrapped := middlewares[0](handler)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	wrapped.ServeHTTP(w, req)
	assert.Equal(t, "test_http", w.Header().Get("X-Test"))
}

func TestRegistry_MCP(t *testing.T) {
	// Register a dummy middleware
	called := false
	RegisterMCP("test_mcp", func(cfg *configv1.Middleware) func(mcp.MethodHandler) mcp.MethodHandler {
		return func(next mcp.MethodHandler) mcp.MethodHandler {
			return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
				called = true
				return next(ctx, method, req)
			}
		}
	})

	configs := []*configv1.Middleware{
		{
			Name:     ptrTo("test_mcp"),
			Priority: ptrTo(int32(10)),
		},
	}

	middlewares := GetMCPMiddlewares(configs)
	assert.Len(t, middlewares, 1)

	// Verify execution
	handler := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return nil, nil
	}
	wrapped := middlewares[0](handler)
	_, _ = wrapped(context.Background(), "test", nil)
	assert.True(t, called)
}
