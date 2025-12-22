// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/mcpany/core/pkg/middleware"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestRegistry_HTTPMiddleware(t *testing.T) {
	middlewareName := "test-http-middleware"
	called := false

	factory := func(config *configv1.Middleware) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				next.ServeHTTP(w, r)
			})
		}
	}

	middleware.Register(middlewareName, factory)

	configs := []*configv1.Middleware{
		{
			Name:     proto.String(middlewareName),
			Priority: proto.Int32(10),
			Disabled: proto.Bool(false),
		},
	}

	middlewares := middleware.GetHTTPMiddlewares(configs)
	require.Len(t, middlewares, 1)

	handler := middlewares[0](http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	handler.ServeHTTP(nil, nil)

	assert.True(t, called)
}

func TestRegistry_MCPMiddleware(t *testing.T) {
	middlewareName := "test-mcp-middleware"
	called := false

	factory := func(config *configv1.Middleware) func(mcp.MethodHandler) mcp.MethodHandler {
		return func(next mcp.MethodHandler) mcp.MethodHandler {
			return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
				called = true
				return next(ctx, method, req)
			}
		}
	}

	middleware.RegisterMCP(middlewareName, factory)

	configs := []*configv1.Middleware{
		{
			Name:     proto.String(middlewareName),
			Priority: proto.Int32(10),
			Disabled: proto.Bool(false),
		},
	}

	middlewares := middleware.GetMCPMiddlewares(configs)
	require.Len(t, middlewares, 1)

	next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return nil, nil
	}
	handler := middlewares[0](next)
	_, _ = handler(context.Background(), "test", nil)

	assert.True(t, called)
}

func TestRegistry_PriorityAndDisabled(t *testing.T) {
	name1 := "p1"
	name2 := "p2"
	var order []string

	factory1 := func(config *configv1.Middleware) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name1)
				next.ServeHTTP(w, r)
			})
		}
	}
	factory2 := func(config *configv1.Middleware) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name2)
				next.ServeHTTP(w, r)
			})
		}
	}

	middleware.Register(name1, factory1)
	middleware.Register(name2, factory2)

	// 1. Check priority
	configs := []*configv1.Middleware{
		{Name: proto.String(name2), Priority: proto.Int32(20)},
		{Name: proto.String(name1), Priority: proto.Int32(10)},
	}

	middlewares := middleware.GetHTTPMiddlewares(configs)
	require.Len(t, middlewares, 2)

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	// Iterate in reverse to wrap: p1(p2(handler))
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	order = []string{}
	handler.ServeHTTP(nil, nil)
	assert.Equal(t, []string{name1, name2}, order)

	// 2. Check Disabled
	configsWithDisabled := []*configv1.Middleware{
		{Name: proto.String(name1), Priority: proto.Int32(10), Disabled: proto.Bool(true)},
		{Name: proto.String(name2), Priority: proto.Int32(20)},
	}

	middlewares = middleware.GetHTTPMiddlewares(configsWithDisabled)
	require.Len(t, middlewares, 1) // Only p2

	handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	handler = middlewares[0](handler)

	order = []string{}
	handler.ServeHTTP(nil, nil)
	assert.Equal(t, []string{name2}, order)
}
