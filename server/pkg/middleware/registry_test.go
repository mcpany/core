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

func TestRegister(t *testing.T) {
	factory := func(config *configv1.Middleware) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	Register("test-http", factory)

	globalRegistry.mu.RLock()
	_, ok := globalRegistry.factories["test-http"]
	globalRegistry.mu.RUnlock()

	assert.True(t, ok)
}

func TestRegisterMCP(t *testing.T) {
	factory := func(config *configv1.Middleware) func(mcp.MethodHandler) mcp.MethodHandler {
		return func(next mcp.MethodHandler) mcp.MethodHandler {
			return next
		}
	}

	RegisterMCP("test-mcp", factory)

	globalRegistry.mu.RLock()
	_, ok := globalRegistry.mcpFactories["test-mcp"]
	globalRegistry.mu.RUnlock()

	assert.True(t, ok)
}

func TestGetHTTPMiddlewares(t *testing.T) {
	// Register a test middleware
	headerMiddleware := func(config *configv1.Middleware) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Test", "true")
				next.ServeHTTP(w, r)
			})
		}
	}
	Register("test-header", headerMiddleware)

	configs := []*configv1.Middleware{
		{
			Name: proto.String("test-header"),
		},
		{
			Name: proto.String("non-existent"),
		},
		{
			Name:     proto.String("test-header"),
			Disabled: proto.Bool(true),
		},
	}

	middlewares := GetHTTPMiddlewares(configs)
	assert.Len(t, middlewares, 1)

	// Verify middleware execution
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	chain := middlewares[0](handler)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	chain.ServeHTTP(rec, req)
	assert.Equal(t, "true", rec.Header().Get("X-Test"))
}

func TestGetMCPMiddlewares(t *testing.T) {
	// Register a test middleware
	contextMiddleware := func(config *configv1.Middleware) func(mcp.MethodHandler) mcp.MethodHandler {
		return func(next mcp.MethodHandler) mcp.MethodHandler {
			return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
				// We can't easily modify context in the chain without wrapping result,
				// but we can check if it's called.
				return next(ctx, method, req)
			}
		}
	}
	RegisterMCP("test-context", contextMiddleware)

	configs := []*configv1.Middleware{
		{
			Name: proto.String("test-context"),
		},
	}

	middlewares := GetMCPMiddlewares(configs)
	assert.Len(t, middlewares, 1)

	// Execute chain
	handler := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return nil, nil
	}

	chain := middlewares[0](handler)
	_, err := chain(context.Background(), "test", nil)
	assert.NoError(t, err)
}

func TestGetHTTPMiddlewares_Sorting(t *testing.T) {
	// Register test middlewares
	Register("p10", func(config *configv1.Middleware) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("X-Order", "10")
				next.ServeHTTP(w, r)
			})
		}
	})
	Register("p20", func(config *configv1.Middleware) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("X-Order", "20")
				next.ServeHTTP(w, r)
			})
		}
	})

	configs := []*configv1.Middleware{
		{Name: proto.String("p20"), Priority: proto.Int32(20)},
		{Name: proto.String("p10"), Priority: proto.Int32(10)},
	}

	middlewares := GetHTTPMiddlewares(configs)
	assert.Len(t, middlewares, 2)

	// Execution order: outer wraps inner.
	// GetHTTPMiddlewares returns [p10, p20] because of sort.
	// So execution chain is p10(p20(handler)).
	// p10 writes, calls p20. p20 writes.

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Construct chain manually to test the order returned
	var chain http.Handler = handler
	// Apply in reverse order so first in list is outermost
	for i := len(middlewares) - 1; i >= 0; i-- {
		chain = middlewares[i](chain)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	chain.ServeHTTP(rec, req)

	// Headers order depends on writer implementation but let's assume Add appends.
	assert.Equal(t, []string{"10", "20"}, rec.Header().Values("X-Order"))
}

func TestInitStandardMiddlewares(t *testing.T) {
	auditConfig := &configv1.AuditConfig{
		Enabled: proto.Bool(true),
	}

	// We pass nil for dependencies as we don't invoke the factories in this test
	// except for audit middleware creation which happens inside.
	err := InitStandardMiddlewares(nil, nil, auditConfig, nil)
	assert.NoError(t, err)

	expected := []string{"auth", "logging", "debug", "cors", "caching", "ratelimit", "call_policy", "audit"}

	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	for _, name := range expected {
		_, ok := globalRegistry.mcpFactories[name]
		assert.True(t, ok, "Middleware %s should be registered", name)
	}
}
