// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
)

func TestRegistry_HTTP(t *testing.T) {
	// Register a test middleware
	mwName := "test-http-middleware"
	called := false
	factory := func(config *configv1.Middleware) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				next.ServeHTTP(w, r)
			})
		}
	}
	Register(mwName, factory)

	t.Run("GetHTTPMiddlewares", func(t *testing.T) {
		configs := []*configv1.Middleware{
			{
				Name: proto.String(mwName),
			},
		}

		mws := GetHTTPMiddlewares(configs)
		assert.Len(t, mws, 1)

		// Verify execution
		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()
		handler := mws[0](http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		handler.ServeHTTP(rr, req)
		assert.True(t, called)
	})

	t.Run("GetHTTPMiddlewares_Disabled", func(t *testing.T) {
		called = false
		configs := []*configv1.Middleware{
			{
				Name:     proto.String(mwName),
				Disabled: proto.Bool(true),
			},
		}

		mws := GetHTTPMiddlewares(configs)
		assert.Len(t, mws, 0)
	})

	t.Run("GetHTTPMiddlewares_Unknown", func(t *testing.T) {
		configs := []*configv1.Middleware{
			{
				Name: proto.String("unknown-middleware"),
			},
		}

		mws := GetHTTPMiddlewares(configs)
		assert.Len(t, mws, 0)
	})

	t.Run("GetHTTPMiddlewares_Priority", func(t *testing.T) {
		mwName2 := "test-http-middleware-2"
		var callOrder []string
		Register(mwName2, func(config *configv1.Middleware) func(http.Handler) http.Handler {
			return func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					callOrder = append(callOrder, mwName2)
					next.ServeHTTP(w, r)
				})
			}
		})

		// Override factory 1 to track order
		Register(mwName, func(config *configv1.Middleware) func(http.Handler) http.Handler {
			return func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					callOrder = append(callOrder, mwName)
					next.ServeHTTP(w, r)
				})
			}
		})

		configs := []*configv1.Middleware{
			{
				Name:     proto.String(mwName),
				Priority: proto.Int32(10),
			},
			{
				Name:     proto.String(mwName2),
				Priority: proto.Int32(5), // Should be first
			},
		}

		mws := GetHTTPMiddlewares(configs)
		assert.Len(t, mws, 2)

		callOrder = nil
		req1 := httptest.NewRequest("GET", "/", nil)
		rr1 := httptest.NewRecorder()
		mws[0](http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).ServeHTTP(rr1, req1)
		assert.Equal(t, []string{mwName2}, callOrder)

		callOrder = nil
		req2 := httptest.NewRequest("GET", "/", nil)
		rr2 := httptest.NewRecorder()
		mws[1](http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).ServeHTTP(rr2, req2)
		assert.Equal(t, []string{mwName}, callOrder)
	})
}

func TestRegistry_MCP(t *testing.T) {
	mwName := "test-mcp-middleware"
	called := false
	factory := func(config *configv1.Middleware) func(mcp.MethodHandler) mcp.MethodHandler {
		return func(next mcp.MethodHandler) mcp.MethodHandler {
			return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
				called = true
				return next(ctx, method, req)
			}
		}
	}
	RegisterMCP(mwName, factory)

	t.Run("GetMCPMiddlewares", func(t *testing.T) {
		configs := []*configv1.Middleware{
			{
				Name: proto.String(mwName),
			},
		}

		mws := GetMCPMiddlewares(configs)
		assert.Len(t, mws, 1)

		// Verify execution
		handler := mws[0](func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			return nil, nil
		})
		_, _ = handler(context.Background(), "method", nil)
		assert.True(t, called)
	})
}

func TestInitStandardMiddlewares(t *testing.T) {
	authManager := auth.NewManager()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	toolManager := tool.NewMockManagerInterface(ctrl)

	auditConfig := &configv1.AuditConfig{
		Enabled: proto.Bool(false),
	}

	cachingMiddleware := NewCachingMiddleware(toolManager)

	cleanup, err := InitStandardMiddlewares(authManager, toolManager, auditConfig, cachingMiddleware)
	assert.NoError(t, err)
	assert.NotNil(t, cleanup)
	defer cleanup()

	// Verify that standard middlewares are registered
	names := []string{"logging", "auth", "debug", "cors", "caching", "ratelimit", "call_policy", "audit"}

	configs := make([]*configv1.Middleware, len(names))
	for i, name := range names {
		configs[i] = &configv1.Middleware{
			Name: proto.String(name),
		}
	}

	// GetMCPMiddlewares should return factory results for all these
	mws := GetMCPMiddlewares(configs)
	assert.Len(t, mws, len(names))
}
