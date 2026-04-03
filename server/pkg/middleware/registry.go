// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"sync"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Registry represents the public Registry entity.
//
// Summary: Defines the structured data model representing a .
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type Registry struct {
	mu           sync.RWMutex
	factories    map[string]Factory
	mcpFactories map[string]MCPFactory
}

// Factory represents the public Factory entity.
//
// Summary: Defines the structured data model representing a .
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type Factory func(config *configv1.Middleware) func(http.Handler) http.Handler

// MCPFactory represents the public MCPFactory entity.
//
// Summary: Defines the structured data model representing a factory.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type MCPFactory func(config *configv1.Middleware) func(mcp.MethodHandler) mcp.MethodHandler

var (
	globalRegistry = &Registry{
		factories:    make(map[string]Factory),
		mcpFactories: make(map[string]MCPFactory),
	}
)

// Register serves as a public interface for interacting with Register.
//
// Summary: Register the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func Register(name string, factory Factory) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.factories[name] = factory
}

// RegisterMCP serves as a public interface for interacting with RegisterMCP.
//
// Summary: Register the mcp appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func RegisterMCP(name string, factory MCPFactory) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.mcpFactories[name] = factory
}

// GetHTTPMiddlewares serves as a public interface for interacting with GetHTTPMiddlewares.
//
// Summary: Fetches and returns the underlying http middlewares from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// GetMCPMiddlewares serves as a public interface for interacting with GetMCPMiddlewares.
//
// Summary: Fetches and returns the underlying mcp middlewares from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// StandardMiddlewares represents the public StandardMiddlewares entity.
//
// Summary: Defines the structured data model representing a middlewares.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type StandardMiddlewares struct {
	Audit            *AuditMiddleware
	GlobalRateLimit  *GlobalRateLimitMiddleware
	ContextOptimizer *ContextOptimizer
	Debugger         *Debugger
	SmartRecovery    *SmartRecoveryMiddleware
	RecursiveContext *RecursiveContextManager
	A2ABridge        *A2ABridgeMiddleware
	ESB              *ESBMiddleware
	CFIA             *CFIAMiddleware
	DiscoverySandbox *DiscoverySandboxMiddleware
	Cleanup          func() error
}

// InitStandardMiddlewares serves as a public interface for interacting with InitStandardMiddlewares.
//
// Summary: Init the standard middlewares appropriately based on current system conditions.
//
// Parameters:
//   - None.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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
	cfiaConfig *CFIAConfig,
	discoverySandboxConfig *DiscoverySandboxConfig,
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

	recursiveContext := NewRecursiveContextManager()
	Register("recursive_context", func(_ *configv1.Middleware) func(http.Handler) http.Handler {
		return recursiveContext.HandleContext
	})

	a2aBridge := NewA2ABridgeMiddleware(recursiveContext)
	RegisterMCP("a2a_bridge", func(_ *configv1.Middleware) func(mcp.MethodHandler) mcp.MethodHandler {
		return func(next mcp.MethodHandler) mcp.MethodHandler {
			return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
				return a2aBridge.Execute(ctx, method, req, next)
			}
		}
	})

	esbMiddleware := NewESBMiddleware(nil)
	RegisterMCP("esb", func(cfg *configv1.Middleware) func(mcp.MethodHandler) mcp.MethodHandler {
		if cfg != nil {
			esbMiddleware = NewESBMiddleware(cfg)
		} else {
			esbMiddleware = NewESBMiddleware(nil)
		}
		return func(next mcp.MethodHandler) mcp.MethodHandler {
			return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
				// Call the ESB execute which wraps the next handler
				return esbMiddleware.Execute(ctx, method, req, next)
			}
		}
	})

	// Context-File Integrity Attestation (CFIA)
	var cfiaMiddleware *CFIAMiddleware
	if cfiaConfig != nil && cfiaConfig.Enabled {
		cfiaMiddleware = NewCFIAMiddleware(*cfiaConfig)
		RegisterMCP("cfia", func(_ *configv1.Middleware) func(mcp.MethodHandler) mcp.MethodHandler {
			return func(next mcp.MethodHandler) mcp.MethodHandler {
				return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
					if method != "tools/call" {
						return next(ctx, method, req)
					}

					callReq, ok := req.(*mcp.CallToolRequest)
					if !ok {
						return next(ctx, method, req)
					}

					var args map[string]interface{}
					if callReq.Params.Arguments != nil {
						// Attempt to unmarshal json.RawMessage to map[string]interface{}
						_ = json.Unmarshal(callReq.Params.Arguments, &args)
					}

					executionReq := &tool.ExecutionRequest{
						ToolName:  callReq.Params.Name,
						Arguments: args,
					}

					result, err := cfiaMiddleware.Execute(ctx, executionReq, func(ctx context.Context, _ *tool.ExecutionRequest) (any, error) {
						return next(ctx, method, req)
					})
					if err != nil {
						return nil, err
					}

					// Convert `any` back to `mcp.Result` safely
					mcpResult, ok := result.(mcp.Result)
					if ok {
						return mcpResult, nil
					}
					return &mcp.CallToolResult{}, nil
				}
			}
		})
	}

	// Discovery-Phase Sandbox
	var discoverySandboxMiddleware *DiscoverySandboxMiddleware
	if discoverySandboxConfig != nil && discoverySandboxConfig.Enabled {
		discoverySandboxMiddleware = NewDiscoverySandboxMiddleware(*discoverySandboxConfig)
		RegisterMCP("discovery_sandbox", func(_ *configv1.Middleware) func(mcp.MethodHandler) mcp.MethodHandler {
			return func(next mcp.MethodHandler) mcp.MethodHandler {
				return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
					if method != "tools/call" {
						return next(ctx, method, req)
					}

					callReq, ok := req.(*mcp.CallToolRequest)
					if !ok {
						return next(ctx, method, req)
					}

					if err := discoverySandboxMiddleware.PreExecute(ctx, callReq, nil); err != nil {
						return nil, err
					}

					res, err := next(ctx, method, req)
					if err != nil {
						return nil, err
					}

					callRes, ok := res.(*mcp.CallToolResult)
					if !ok {
						return res, nil
					}

					return discoverySandboxMiddleware.PostExecute(ctx, callReq, callRes, nil)
				}
			}
		})
	}

	return &StandardMiddlewares{
		Audit:            audit,
		GlobalRateLimit:  globalRateLimit,
		ContextOptimizer: contextOptimizer,
		Debugger:         debugger,
		SmartRecovery:    smartRecovery,
		RecursiveContext: recursiveContext,
		A2ABridge:        a2aBridge,
		ESB:              esbMiddleware,
		CFIA:             cfiaMiddleware,
		DiscoverySandbox: discoverySandboxMiddleware,
		Cleanup:          audit.Close,
	}, nil
}
