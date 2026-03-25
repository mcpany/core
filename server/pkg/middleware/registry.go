// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"net/http"
	"sort"
	"sync"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Registry manages available middlewares in a thread-safe manner.
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

// Register registers a HTTP middleware factory in the global registry.
//
// Parameters:
//   - name (string): The unique identifier for the middleware.
//   - factory (Factory): The function to create the middleware instance.
//
// Side Effects:
//   - Modifies the global factories map.
func Register(name string, factory Factory) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.factories[name] = factory
}

// RegisterMCP registers an MCP middleware factory in the global registry.
//
// Parameters:
//   - name (string): The unique identifier for the MCP middleware.
//   - factory (MCPFactory): The function to create the middleware instance.
//
// Side Effects:
//   - Modifies the global mcpFactories map.
func RegisterMCP(name string, factory MCPFactory) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.mcpFactories[name] = factory
}

// GetHTTPMiddlewares returns a sorted list of HTTP middlewares based on configuration priority.
//
// Parameters:
//   - configs ([]*configv1.Middleware): The middleware configurations to filter and sort.
//
// Returns:
//   - []func(http.Handler) http.Handler: A slice of active HTTP middleware functions.
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

// GetMCPMiddlewares returns a sorted list of MCP middlewares based on configuration priority.
//
// Parameters:
//   - configs ([]*configv1.Middleware): The middleware configurations to filter and sort.
//
// Returns:
//   - []func(mcp.MethodHandler) mcp.MethodHandler: A slice of active MCP middleware functions.
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

// StandardMiddlewares holds the standard middlewares that might need to be updated.
type StandardMiddlewares struct {
	Audit            *AuditMiddleware
	GlobalRateLimit  *GlobalRateLimitMiddleware
	ContextOptimizer *ContextOptimizer
	Debugger         *Debugger
	SmartRecovery    *SmartRecoveryMiddleware
	RecursiveContext *RecursiveContextManager
	A2ABridge        *A2ABridgeMiddleware
	Cleanup          func() error
}

// InitStandardMiddlewares initializes and registers the default set of system middlewares.
//
// Parameters:
//   - authManager (*auth.Manager): The manager handling user authentication.
//   - toolManager (tool.ManagerInterface): The interface for tool lifecycle management.
//   - auditConfig (*configv1.AuditConfig): Configuration for request auditing.
//   - cachingMiddleware (*CachingMiddleware): The middleware instance for result caching.
//   - globalRateLimitConfig (*configv1.RateLimitConfig): Configuration for global request throttling.
//   - dlpConfig (*configv1.DLPConfig): Configuration for Data Loss Prevention scanning.
//   - contextOptimizerConfig (*configv1.ContextOptimizerConfig): Configuration for context window optimization.
//   - debuggerConfig (*configv1.DebuggerConfig): Configuration for the request debugger.
//   - smartRecoveryConfig (*configv1.SmartRecoveryConfig): Configuration for autonomous error recovery.
//
// Returns:
//   - *StandardMiddlewares: A container holding the initialized middleware instances.
//   - error: An error if any standard middleware fails to initialize.
func InitStandardMiddlewares(
	authManager *auth.Manager,
	toolManager tool.ManagerInterface,
	auditConfig *configv1.AuditConfig,
	cachingMiddleware *CachingMiddleware,
	globalRateLimitConfig *configv1.RateLimitConfig,
	dlpConfig *configv1.DLPConfig,
	contextOptimizerConfig *configv1.ContextOptimizerConfig,
	debuggerConfig *configv1.DebuggerConfig,
	smartRecoveryConfig *configv1.SmartRecoveryConfig,
) (*StandardMiddlewares, error) {
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
	// Audit middleware needs to be closed to ensure file handles are released.
	audit, err := NewAuditMiddleware(auditConfig)
	if err != nil {
		return nil, err
	}

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

	// Global Rate Limit
	globalRateLimit := NewGlobalRateLimitMiddleware(globalRateLimitConfig)
	RegisterMCP("global_ratelimit", func(_ *configv1.Middleware) func(mcp.MethodHandler) mcp.MethodHandler {
		return func(next mcp.MethodHandler) mcp.MethodHandler {
			return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
				return globalRateLimit.Execute(ctx, method, req, next)
			}
		}
	})

	// DLP
	RegisterMCP("dlp", func(_ *configv1.Middleware) func(mcp.MethodHandler) mcp.MethodHandler {
		return DLPMiddleware(dlpConfig, nil)
	})

	// Context Optimizer
	var contextOptimizer *ContextOptimizer
	if contextOptimizerConfig != nil {
		maxChars := int(contextOptimizerConfig.GetMaxChars())
		if maxChars == 0 {
			maxChars = 32000 // Default to approx 8000 tokens
		}
		contextOptimizer = NewContextOptimizer(maxChars)
		Register("context_optimizer", func(_ *configv1.Middleware) func(http.Handler) http.Handler {
			return contextOptimizer.Handler
		})
	}

	// Debugger
	var debugger *Debugger
	if debuggerConfig != nil && debuggerConfig.GetEnabled() {
		size := int(debuggerConfig.GetSize())
		if size == 0 {
			size = 100 // Default
		}
		debugger = NewDebugger(size)
		Register("debugger", func(_ *configv1.Middleware) func(http.Handler) http.Handler {
			return debugger.Handler
		})
	}

	// Smart Recovery
	smartRecovery := NewSmartRecoveryMiddleware(smartRecoveryConfig, toolManager)

	recursiveContext := NewRecursiveContextManager()
	Register("recursive_context", func(_ *configv1.Middleware) func(http.Handler) http.Handler {
		return recursiveContext.HandleContext
	})
	RegisterMCP("smart_recovery", func(_ *configv1.Middleware) func(mcp.MethodHandler) mcp.MethodHandler {
		return func(next mcp.MethodHandler) mcp.MethodHandler {
			return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
				if r, ok := req.(*mcp.CallToolRequest); ok {
					executionReq := &tool.ExecutionRequest{
						ToolName:   r.Params.Name,
						ToolInputs: r.Params.Arguments,
					}
					result, err := smartRecovery.Execute(ctx, executionReq, func(ctx context.Context, updatedReq *tool.ExecutionRequest) (any, error) {
						// Propagate updated arguments to the MCP request
						r.Params.Arguments = updatedReq.ToolInputs
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

	a2aBridge := NewA2ABridgeMiddleware(recursiveContext)
	RegisterMCP("a2a_bridge", func(_ *configv1.Middleware) func(mcp.MethodHandler) mcp.MethodHandler {
		return func(next mcp.MethodHandler) mcp.MethodHandler {
			return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
				return a2aBridge.Execute(ctx, method, req, next)
			}
		}
	})

	return &StandardMiddlewares{
		Audit:            audit,
		GlobalRateLimit:  globalRateLimit,
		ContextOptimizer: contextOptimizer,
		Debugger:         debugger,
		SmartRecovery:    smartRecovery,
		RecursiveContext: recursiveContext,
		A2ABridge:        a2aBridge,
		Cleanup:          audit.Close,
	}, nil
}
