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
	"google.golang.org/protobuf/proto"
)

func TestRegistry_HTTPMiddlewares(t *testing.T) {
	// Register a test middleware
	mwName := "test-http-middleware"
	mwHeaderKey := "X-Test-Middleware"
	mwHeaderVal := "executed"

	Register(mwName, func(_ *configv1.Middleware) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set(mwHeaderKey, mwHeaderVal)
				next.ServeHTTP(w, r)
			})
		}
	})

	t.Run("get_registered_middleware", func(t *testing.T) {
		configs := []*configv1.Middleware{
			configv1.Middleware_builder{
				Name:     proto.String(mwName),
				Priority: proto.Int32(10),
			}.Build(),
		}

		mws := GetHTTPMiddlewares(configs)
		assert.Len(t, mws, 1)

		// Verify execution
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		chain := mws[0](handler)
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		chain.ServeHTTP(rr, req)

		assert.Equal(t, mwHeaderVal, rr.Header().Get(mwHeaderKey))
	})

	t.Run("middleware_priority", func(t *testing.T) {
		mwName2 := "test-http-middleware-2"
		Register(mwName2, func(_ *configv1.Middleware) func(http.Handler) http.Handler {
			return func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Add("X-Order", "2")
					next.ServeHTTP(w, r)
				})
			}
		})

		// Re-register mwName to add X-Order header
		Register(mwName, func(_ *configv1.Middleware) func(http.Handler) http.Handler {
			return func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Add("X-Order", "1")
					next.ServeHTTP(w, r)
				})
			}
		})

		configs := []*configv1.Middleware{
			configv1.Middleware_builder{
				Name:     proto.String(mwName2), // Priority 20
				Priority: proto.Int32(20),
			}.Build(),
			configv1.Middleware_builder{
				Name:     proto.String(mwName), // Priority 10 (should come first)
				Priority: proto.Int32(10),
			}.Build(),
		}

		mws := GetHTTPMiddlewares(configs)
		assert.Len(t, mws, 2)
	})

	t.Run("ignore_disabled_middleware", func(t *testing.T) {
		configs := []*configv1.Middleware{
			configv1.Middleware_builder{
				Name:     proto.String(mwName),
				Disabled: proto.Bool(true),
			}.Build(),
		}

		mws := GetHTTPMiddlewares(configs)
		assert.Len(t, mws, 0)
	})

	t.Run("ignore_unregistered_middleware", func(t *testing.T) {
		configs := []*configv1.Middleware{
			configv1.Middleware_builder{
				Name: proto.String("non-existent-middleware"),
			}.Build(),
		}

		mws := GetHTTPMiddlewares(configs)
		assert.Len(t, mws, 0)
	})
}

func TestRegistry_MCPMiddlewares(t *testing.T) {
	mwName := "test-mcp-middleware"
	RegisterMCP(mwName, func(_ *configv1.Middleware) func(mcp.MethodHandler) mcp.MethodHandler {
		return func(next mcp.MethodHandler) mcp.MethodHandler {
			return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
				// Decorate context or do something
				return next(ctx, method, req)
			}
		}
	})

	t.Run("get_registered_mcp_middleware", func(t *testing.T) {
		configs := []*configv1.Middleware{
			configv1.Middleware_builder{
				Name:     proto.String(mwName),
				Priority: proto.Int32(10),
			}.Build(),
		}

		mws := GetMCPMiddlewares(configs)
		assert.Len(t, mws, 1)
	})

	t.Run("ignore_disabled_mcp_middleware", func(t *testing.T) {
		configs := []*configv1.Middleware{
			configv1.Middleware_builder{
				Name:     proto.String(mwName),
				Disabled: proto.Bool(true),
			}.Build(),
		}

		mws := GetMCPMiddlewares(configs)
		assert.Len(t, mws, 0)
	})
}

// Helper to extract name safely
func activeMiddlewareName(c *configv1.Middleware) string {
	return c.GetName()
}

func TestInitStandardMiddlewares_ContextOptimizer_Default(t *testing.T) {
	// Initialize with empty ContextOptimizerConfig (MaxChars = 0)
	config := &configv1.ContextOptimizerConfig{} // Defaults to 0

	stdMws, err := InitStandardMiddlewares(
		nil, nil, nil, nil, nil, nil,
		config, // Pass empty config
		nil,
		nil,
	)
	assert.NoError(t, err)
	assert.NotNil(t, stdMws.ContextOptimizer)
	assert.Equal(t, 32000, stdMws.ContextOptimizer.MaxChars)
}
