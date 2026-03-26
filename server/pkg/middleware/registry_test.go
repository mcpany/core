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

<<<<<<< HEAD
// ClearRegistryForTesting is used only in testing to clear the global registry
func ClearRegistryForTesting() {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.factories = make(map[string]Factory)
	globalRegistry.mcpFactories = make(map[string]MCPFactory)
}

func TestRegistry_HTTPMiddlewares(t *testing.T) {
	t.Cleanup(ClearRegistryForTesting)
=======
func TestRegistry_HTTPMiddlewares(t *testing.T) {
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
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

<<<<<<< HEAD
	t.Run("sorts_middlewares_by_priority", func(t *testing.T) {
		// Test multiple middlewares to ensure sort.Slice handles out-of-order priorities
		mw1 := "sort-test-1"
		mw2 := "sort-test-2"
		mw3 := "sort-test-3"

		Register(mw1, func(_ *configv1.Middleware) func(http.Handler) http.Handler {
			return func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Add("X-Order", "1")
					next.ServeHTTP(w, r)
				})
			}
		})
		Register(mw2, func(_ *configv1.Middleware) func(http.Handler) http.Handler {
			return func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Add("X-Order", "2")
					next.ServeHTTP(w, r)
				})
			}
		})
		Register(mw3, func(_ *configv1.Middleware) func(http.Handler) http.Handler {
			return func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Add("X-Order", "3")
					next.ServeHTTP(w, r)
				})
			}
		})

		configs := []*configv1.Middleware{
			configv1.Middleware_builder{
				Name:     proto.String(mw2),
				Priority: proto.Int32(20),
			}.Build(),
			configv1.Middleware_builder{
				Name:     proto.String(mw3),
				Priority: proto.Int32(30),
			}.Build(),
			configv1.Middleware_builder{
				Name:     proto.String(mw1),
				Priority: proto.Int32(10),
			}.Build(),
		}

		mws := GetHTTPMiddlewares(configs)
		assert.Len(t, mws, 3)

		// Execute the chain
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		var chain http.Handler = handler
		for i := len(mws) - 1; i >= 0; i-- {
			chain = mws[i](chain)
		}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		chain.ServeHTTP(rr, req)

		// Assert order: 10, 20, 30
		assert.Equal(t, []string{"1", "2", "3"}, rr.Header().Values("X-Order"))
	})

=======
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
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
<<<<<<< HEAD
	t.Cleanup(ClearRegistryForTesting)
=======
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
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

<<<<<<< HEAD
	t.Run("sorts_middlewares_by_priority", func(t *testing.T) {
		// Test multiple middlewares to ensure sort.Slice handles out-of-order priorities
		mw1 := "mcp-sort-test-1"
		mw2 := "mcp-sort-test-2"
		mw3 := "mcp-sort-test-3"

		RegisterMCP(mw1, func(_ *configv1.Middleware) func(mcp.MethodHandler) mcp.MethodHandler {
			return func(next mcp.MethodHandler) mcp.MethodHandler {
				return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
					ctx = context.WithValue(ctx, "trace", append(ctx.Value("trace").([]string), "1"))
					return next(ctx, method, req)
				}
			}
		})
		RegisterMCP(mw2, func(_ *configv1.Middleware) func(mcp.MethodHandler) mcp.MethodHandler {
			return func(next mcp.MethodHandler) mcp.MethodHandler {
				return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
					ctx = context.WithValue(ctx, "trace", append(ctx.Value("trace").([]string), "2"))
					return next(ctx, method, req)
				}
			}
		})
		RegisterMCP(mw3, func(_ *configv1.Middleware) func(mcp.MethodHandler) mcp.MethodHandler {
			return func(next mcp.MethodHandler) mcp.MethodHandler {
				return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
					ctx = context.WithValue(ctx, "trace", append(ctx.Value("trace").([]string), "3"))
					return next(ctx, method, req)
				}
			}
		})

		configs := []*configv1.Middleware{
			configv1.Middleware_builder{
				Name:     proto.String(mw2),
				Priority: proto.Int32(20),
			}.Build(),
			configv1.Middleware_builder{
				Name:     proto.String(mw3),
				Priority: proto.Int32(30),
			}.Build(),
			configv1.Middleware_builder{
				Name:     proto.String(mw1),
				Priority: proto.Int32(10),
			}.Build(),
		}

		mws := GetMCPMiddlewares(configs)
		assert.Len(t, mws, 3)

		// Execute the chain
		var resultTrace []string
		handler := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			resultTrace = ctx.Value("trace").([]string)
			return nil, nil
		}
		chain := handler
		for i := len(mws) - 1; i >= 0; i-- {
			chain = mws[i](chain)
		}

		ctx := context.WithValue(context.Background(), "trace", []string{})
		_, err := chain(ctx, "test", nil)
		assert.NoError(t, err)

		// Assert order: 10, 20, 30
		assert.Equal(t, []string{"1", "2", "3"}, resultTrace)
	})

=======
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
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
<<<<<<< HEAD
		nil,
=======
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
	)
	assert.NoError(t, err)
	assert.NotNil(t, stdMws.ContextOptimizer)
	assert.Equal(t, 32000, stdMws.ContextOptimizer.MaxChars)
}
