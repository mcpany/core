// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"net/http"
	"sort"
	"sync"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Registry manages available middlewares.
type Registry struct {
	mu           sync.RWMutex
	factories    map[string]Factory
	mcpFactories map[string]MCPFactory
}

// Factory is a function that creates a HTTP middleware from configuration.
type Factory func(config *configv1.Middleware) func(http.Handler) http.Handler

// MCPFactory is a function that creates an MCP middleware from configuration.
type MCPFactory func(config *configv1.Middleware) func(mcp.MethodHandler) mcp.MethodHandler

var (
	globalRegistry = &Registry{
		factories:    make(map[string]Factory),
		mcpFactories: make(map[string]MCPFactory),
	}
)

// Register registers a HTTP middleware factory.
func Register(name string, factory Factory) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.factories[name] = factory
}

// RegisterMCP registers an MCP middleware factory.
func RegisterMCP(name string, factory MCPFactory) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.mcpFactories[name] = factory
}

// GetHTTPMiddlewares returns a sorted list of HTTP middlewares based on configuration.
func GetHTTPMiddlewares(configs []*configv1.Middleware) []func(http.Handler) http.Handler {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	active := make([]*configv1.Middleware, 0, len(configs))
	for _, cfg := range configs {
		if !cfg.GetDisabled() && globalRegistry.factories[cfg.GetName()] != nil {
			active = append(active, cfg)
		}
	}

	sort.Slice(active, func(i, j int) bool {
		return active[i].GetPriority() < active[j].GetPriority()
	})

	middlewares := make([]func(http.Handler) http.Handler, 0, len(active))
	for _, cfg := range active {
		factory := globalRegistry.factories[cfg.GetName()]
		middlewares = append(middlewares, factory(cfg))
	}
	return middlewares
}

// GetMCPMiddlewares returns a sorted list of MCP middlewares based on configuration.
func GetMCPMiddlewares(configs []*configv1.Middleware) []func(mcp.MethodHandler) mcp.MethodHandler {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	active := make([]*configv1.Middleware, 0, len(configs))
	for _, cfg := range configs {
		if !cfg.GetDisabled() && globalRegistry.mcpFactories[cfg.GetName()] != nil {
			active = append(active, cfg)
		}
	}

	sort.Slice(active, func(i, j int) bool {
		return active[i].GetPriority() < active[j].GetPriority()
	})

	middlewares := make([]func(mcp.MethodHandler) mcp.MethodHandler, 0, len(active))
	for _, cfg := range active {
		factory := globalRegistry.mcpFactories[cfg.GetName()]
		middlewares = append(middlewares, factory(cfg))
	}
	return middlewares
}

// InitStandardMiddlewares registers standard middlewares.
func InitStandardMiddlewares(
	authManager *auth.Manager,
	toolManager tool.ManagerInterface,
	auditConfig *configv1.AuditConfig,
	cachingMiddleware *CachingMiddleware,
) error {
	// 1. Logging
	RegisterMCP("logging", func(_ *configv1.Middleware) func(mcp.MethodHandler) mcp.MethodHandler {
		return LoggingMiddleware(nil)
	})

	// 2. Auth
	RegisterMCP("auth", func(_ *configv1.Middleware) func(mcp.MethodHandler) mcp.MethodHandler {
		return AuthMiddleware(authManager)
	})

	// 3. Debug
	RegisterMCP("debug", func(_ *configv1.Middleware) func(mcp.MethodHandler) mcp.MethodHandler {
		return DebugMiddleware()
	})

	// 4. CORS
	RegisterMCP("cors", func(_ *configv1.Middleware) func(mcp.MethodHandler) mcp.MethodHandler {
		return CORSMiddleware()
	})

	// Tool-specific middlewares: Caching, RateLimit, CallPolicy, Audit
	// These need to wrap the tool execution logic.

	// Caching
	RegisterMCP("caching", func(_ *configv1.Middleware) func(mcp.MethodHandler) mcp.MethodHandler {
		return func(next mcp.MethodHandler) mcp.MethodHandler {
			return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
				if r, ok := req.(*mcp.CallToolRequest); ok {
					executionReq := &tool.ExecutionRequest{
						ToolName:   r.Params.Name,
						ToolInputs: r.Params.Arguments,
					}
					// Caching middleware expects a 'next' that returns (any, error)
					result, err := cachingMiddleware.Execute(ctx, executionReq, func(ctx context.Context, _ *tool.ExecutionRequest) (any, error) {
						return next(ctx, method, req)
					})
					if err != nil {
						return nil, err
					}
					if res, ok := result.(*mcp.CallToolResult); ok {
						return res, nil
					}
					return nil, nil // Should not happen if caching returns correct type
				}
				return next(ctx, method, req)
			}
		}
	})

	// Rate Limit
	rateLimit := NewRateLimitMiddleware(toolManager)
	RegisterMCP("ratelimit", func(_ *configv1.Middleware) func(mcp.MethodHandler) mcp.MethodHandler {
		return func(next mcp.MethodHandler) mcp.MethodHandler {
			return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
				if r, ok := req.(*mcp.CallToolRequest); ok {
					executionReq := &tool.ExecutionRequest{
						ToolName:   r.Params.Name,
						ToolInputs: r.Params.Arguments,
					}
					result, err := rateLimit.Execute(ctx, executionReq, func(ctx context.Context, _ *tool.ExecutionRequest) (any, error) {
						return next(ctx, method, req)
					})
					if err != nil {
						return nil, err
					}
					if res, ok := result.(*mcp.CallToolResult); ok {
						return res, nil
					}
				}
				return next(ctx, method, req)
			}
		}
	})

	// Call Policy
	callPolicy := NewCallPolicyMiddleware(toolManager)
	RegisterMCP("call_policy", func(_ *configv1.Middleware) func(mcp.MethodHandler) mcp.MethodHandler {
		return func(next mcp.MethodHandler) mcp.MethodHandler {
			return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
				if r, ok := req.(*mcp.CallToolRequest); ok {
					executionReq := &tool.ExecutionRequest{
						ToolName:   r.Params.Name,
						ToolInputs: r.Params.Arguments,
					}
					result, err := callPolicy.Execute(ctx, executionReq, func(ctx context.Context, _ *tool.ExecutionRequest) (any, error) {
						return next(ctx, method, req)
					})
					if err != nil {
						return nil, err
					}
					if res, ok := result.(*mcp.CallToolResult); ok {
						return res, nil
					}
				}
				return next(ctx, method, req)
			}
		}
	})

	// Audit
	// Audit middleware might need to be closed? server.go deferred Close().
	// We can't easy defer Close here.
	// The AuditMiddleware struct has a Close method.
	// We should probably allow the caller to manage the lifecycle or make it a singleton in App.
	// For now, let's assume it's safe OR we refactor how it's created.
	// Actually, `NewAuditMiddleware` returns (*AuditMiddleware, error).
	audit, err := NewAuditMiddleware(auditConfig)
	if err != nil {
		return err
	}
	// TODO: Handle closing audit middleware?
	// Maybe we just don't close it globally here, relying on app shutdown?
	// Or we add a Shutdown/Close method to registry?

	RegisterMCP("audit", func(_ *configv1.Middleware) func(mcp.MethodHandler) mcp.MethodHandler {
		return func(next mcp.MethodHandler) mcp.MethodHandler {
			return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
				if r, ok := req.(*mcp.CallToolRequest); ok {
					executionReq := &tool.ExecutionRequest{
						ToolName:   r.Params.Name,
						ToolInputs: r.Params.Arguments,
					}
					result, err := audit.Execute(ctx, executionReq, func(ctx context.Context, _ *tool.ExecutionRequest) (any, error) {
						return next(ctx, method, req)
					})
					if err != nil {
						return nil, err
					}
					if res, ok := result.(*mcp.CallToolResult); ok {
						return res, nil
					}
				}
				return next(ctx, method, req)
			}
		}
	})

	return nil
}
