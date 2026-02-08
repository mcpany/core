// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"net/http"
	"sort"
	"sync"

	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Registry manages available middlewares.
//
// Summary: manages available middlewares.
type Registry struct {
	mu           sync.RWMutex
	factories    map[string]Factory
	mcpFactories map[string]MCPFactory
}

// Factory is a function that creates a HTTP middleware from configuration.
//
// Summary: is a function that creates a HTTP middleware from configuration.
type Factory func(config *configv1.Middleware) func(http.Handler) http.Handler

// MCPFactory is a function that creates an MCP middleware from configuration.
//
// Summary: is a function that creates an MCP middleware from configuration.
type MCPFactory func(config *configv1.Middleware) func(mcp.MethodHandler) mcp.MethodHandler

var (
	globalRegistry = &Registry{
		factories:    make(map[string]Factory),
		mcpFactories: make(map[string]MCPFactory),
	}
)

// Register registers a HTTP middleware factory.
//
// Summary: registers a HTTP middleware factory.
//
// Parameters:
//   - name: string. The name.
//   - factory: Factory. The factory.
//
// Returns:
//   None.
func Register(name string, factory Factory) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.factories[name] = factory
}

// RegisterMCP registers an MCP middleware factory.
//
// Summary: registers an MCP middleware factory.
//
// Parameters:
//   - name: string. The name.
//   - factory: MCPFactory. The factory.
//
// Returns:
//   None.
func RegisterMCP(name string, factory MCPFactory) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.mcpFactories[name] = factory
}

// GetHTTPMiddlewares returns a sorted list of HTTP middlewares based on configuration.
//
// Summary: returns a sorted list of HTTP middlewares based on configuration.
//
// Parameters:
//   - configs: []*configv1.Middleware. The configs.
//
// Returns:
//   - []func(http.Handler) http.Handler: The []func(http.Handler) http.Handler.
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
//
// Summary: returns a sorted list of MCP middlewares based on configuration.
//
// Parameters:
//   - configs: []*configv1.Middleware. The configs.
//
// Returns:
//   - []func(mcp.MethodHandler) mcp.MethodHandler: The []func(mcp.MethodHandler) mcp.MethodHandler.
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
//
// Summary: holds the standard middlewares that might need to be updated.
type StandardMiddlewares struct {
	Audit            *AuditMiddleware
	GlobalRateLimit  *GlobalRateLimitMiddleware
	ContextOptimizer *ContextOptimizer
	Debugger         *Debugger
	SmartRecovery    *SmartRecoveryMiddleware
	Cleanup          func() error
}

// InitStandardMiddlewares registers standard middlewares.
//
// Summary: registers standard middlewares.
//
// Parameters:
//   - authManager: *auth.Manager. The authManager.
//   - toolManager: tool.ManagerInterface. The toolManager.
//   - auditConfig: *configv1.AuditConfig. The auditConfig.
//   - cachingMiddleware: *CachingMiddleware. The cachingMiddleware.
//   - globalRateLimitConfig: *configv1.RateLimitConfig. The globalRateLimitConfig.
//   - dlpConfig: *configv1.DLPConfig. The dlpConfig.
//   - contextOptimizerConfig: *configv1.ContextOptimizerConfig. The contextOptimizerConfig.
//   - debuggerConfig: *configv1.DebuggerConfig. The debuggerConfig.
//   - smartRecoveryConfig: *configv1.SmartRecoveryConfig. The smartRecoveryConfig.
//
// Returns:
//   - *StandardMiddlewares: The *StandardMiddlewares.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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
		// Logger will be injected by DLPMiddleware constructor or we use default?
		// DLPMiddleware takes (*configv1.DLPConfig, *slog.Logger)
		// We use package level logger or similar.
		// NOTE: DLPMiddleware signature is: func DLPMiddleware(config *configv1.DLPConfig, log *slog.Logger) mcp.Middleware
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

	return &StandardMiddlewares{
		Audit:            audit,
		GlobalRateLimit:  globalRateLimit,
		ContextOptimizer: contextOptimizer,
		Debugger:         debugger,
		SmartRecovery:    smartRecovery,
		Cleanup:          audit.Close,
	}, nil
}
