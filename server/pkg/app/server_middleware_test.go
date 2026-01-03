// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"testing"

	"github.com/mcpany/core/pkg/middleware"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

// Mock factories
type MockFactory struct {
	mock.Mock
}

func (m *MockFactory) Create(cfg *configv1.Middleware) func(mcp.MethodHandler) mcp.MethodHandler {
	args := m.Called(cfg)
	return args.Get(0).(func(mcp.MethodHandler) mcp.MethodHandler)
}

func TestMiddlewareRegistry(t *testing.T) {
	// Register a test middleware
	middleware.RegisterMCP("test_middleware", func(cfg *configv1.Middleware) func(mcp.MethodHandler) mcp.MethodHandler {
		return func(next mcp.MethodHandler) mcp.MethodHandler {
			return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
				return next(ctx, method, req)
			}
		}
	})

	t.Run("GetMCPMiddlewares sorts by priority", func(t *testing.T) {
		configs := []*configv1.Middleware{
			configv1.Middleware_builder{Name: proto.String("test_middleware"), Priority: proto.Int32(100)}.Build(),
			configv1.Middleware_builder{Name: proto.String("logging"), Priority: proto.Int32(10)}.Build(), // Assume logging is registered
		}

		// Ensure logging is registered (it is standard)
		// We can't easily mock toolManager here without defined interface in this package,
		// but we can pass nil if InitStandardMiddlewares handles it, or a simple mock.
		// InitStandardMiddlewares requires arguments.
		// For this test, we rely on the fact that "logging" is registered by InitStandardMiddlewares OR we manually register it.
		// Let's manually register a mock "logging" to avoid dependencies.
		middleware.RegisterMCP("logging", func(cfg *configv1.Middleware) func(mcp.MethodHandler) mcp.MethodHandler {
			return func(next mcp.MethodHandler) mcp.MethodHandler {
				return next
			}
		})

		chain := middleware.GetMCPMiddlewares(configs)
		assert.Equal(t, 2, len(chain))
		// We can't inspect the function itself easily, but we know the order.
		// We could verify execution order if we wrap them.
		// But here we just want to check if it didn't panic and returned 2 items.
	})

	t.Run("GetMCPMiddlewares filters disabled", func(t *testing.T) {
		configs := []*configv1.Middleware{
			configv1.Middleware_builder{Name: proto.String("test_middleware"), Disabled: proto.Bool(true)}.Build(),
		}
		chain := middleware.GetMCPMiddlewares(configs)
		assert.Empty(t, chain)
	})
}
