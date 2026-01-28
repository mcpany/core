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

// Register registers a new HTTP middleware factory with the global registry.
//
// Parameters:
//   - name: The unique name to identify the middleware (e.g., "gzip", "cors").
//   - factory: The factory function that creates the middleware instance from configuration.
//
// Side Effects:
//   - Updates the global middleware registry.
//   - This function is thread-safe.
func Register(name string, factory Factory) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.factories[name] = factory
}

// RegisterMCP registers a new MCP middleware factory with the global registry.
//
// Parameters:
//   - name: The unique name to identify the middleware (e.g., "logging", "auth").
//   - factory: The factory function that creates the MCP middleware instance from configuration.
//
// Side Effects:
//   - Updates the global middleware registry.
//   - This function is thread-safe.
func RegisterMCP(name string, factory MCPFactory) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.mcpFactories[name] = factory
}

// GetHTTPMiddlewares constructs a sorted list of HTTP middlewares based on the provided configuration.
//
// It filters out disabled middlewares and those not present in the registry, then sorts
// the remaining ones by priority before instantiating them using their registered factories.
//
// Parameters:
//   - configs: A slice of middleware configurations from the server config.
//
// Returns:
//   - []func(http.Handler) http.Handler: A slice of HTTP middleware functions ready to be applied.
//
// Side Effects:
//   - Acquires a read lock on the global registry.
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

// GetMCPMiddlewares constructs a sorted list of MCP middlewares based on the provided configuration.
//
// It filters out disabled middlewares and those not present in the registry, then sorts
// the remaining ones by priority before instantiating them using their registered factories.
//
// Parameters:
//   - configs: A slice of middleware configurations from the server config.
//
// Returns:
//   - []func(mcp.MethodHandler) mcp.MethodHandler: A slice of MCP middleware functions ready to be applied.
//
// Side Effects:
//   - Acquires a read lock on the global registry.
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
	Cleanup          func() error
}

// InitStandardMiddlewares initializes and registers the set of core middlewares included with the server.
//
// This function sets up middlewares for Logging, Authentication, Debugging, CORS, Caching, Rate Limiting,
// Call Policy, Audit Logging, DLP, and more. It wires them up with their respective dependencies
// and configurations.
//
// Parameters:
//   - authManager: The authentication manager for verifying credentials.
//   - toolManager: The tool manager for inspecting tool execution requests.
//   - auditConfig: Configuration for the audit logging subsystem.
//   - cachingMiddleware: The pre-configured caching middleware instance.
//   - globalRateLimitConfig: Configuration for the global rate limiter.
//   - dlpConfig: Configuration for Data Loss Prevention (DLP).
//   - contextOptimizerConfig: Configuration for the context optimizer.
//   - debuggerConfig: Configuration for the request debugger.
//
// Returns:
//   - *StandardMiddlewares: A struct containing references to the initialized middlewares.
//   - error: An error if initialization of any middleware fails (e.g., Audit store).
//
// Side Effects:
//   - Registers multiple middlewares in the global registry via Register() and RegisterMCP().
//   - May open connections to external services (e.g., Audit DB).
func InitStandardMiddlewares(
	authManager *auth.Manager,
	toolManager tool.ManagerInterface,
	auditConfig *configv1.AuditConfig,
	cachingMiddleware *CachingMiddleware,
	globalRateLimitConfig *configv1.RateLimitConfig,
	dlpConfig *configv1.DLPConfig,
	contextOptimizerConfig *configv1.ContextOptimizerConfig,
	debuggerConfig *configv1.DebuggerConfig,
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
		contextOptimizer = NewContextOptimizer(int(contextOptimizerConfig.GetMaxChars()))
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

	return &StandardMiddlewares{
		Audit:            audit,
		GlobalRateLimit:  globalRateLimit,
		ContextOptimizer: contextOptimizer,
		Debugger:         debugger,
		Cleanup:          audit.Close,
	}, nil
}
