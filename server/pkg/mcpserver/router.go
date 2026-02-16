// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MethodHandler defines the signature for a function that handles an MCP method call.
//
// Summary: A function type for handling MCP method calls.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - req: mcp.Request. The request object.
//
// Returns:
//   - mcp.Result: The result of the operation.
//   - error: An error if the operation fails.
type MethodHandler func(ctx context.Context, req mcp.Request) (mcp.Result, error)

// Router is responsible for mapping MCP method names to their corresponding handler functions.
//
// Summary: A router for dispatching MCP requests to handlers.
//
// It maintains a registry of method names and their associated handler functions,
// allowing the server to route incoming JSON-RPC requests to the correct logic.
//
// Side Effects:
//   - Stores handlers in an internal map.
type Router struct {
	handlers map[string]MethodHandler
}

// NewRouter creates and returns a new, empty Router.
//
// Summary: Initializes a new Router instance.
//
// It allocates the internal map for storing method handlers.
//
// Returns:
//   - *Router: A pointer to a new, initialized Router.
func NewRouter() *Router {
	return &Router{
		handlers: make(map[string]MethodHandler),
	}
}

// Register associates a handler function with a specific MCP method name.
//
// Summary: Registers a handler for a given MCP method.
//
// It stores the provided handler function in the router's internal map, keyed by the method name.
// When a request with this method name is received, the router will dispatch it to this handler.
//
// Parameters:
//   - method: string. The name of the MCP method (e.g., "tools/call").
//   - handler: MethodHandler. The function that will handle the method call.
//
// Returns:
//   None.
//
// Side Effects:
//   - If a handler for the method already exists, it will be overwritten.
func (r *Router) Register(method string, handler MethodHandler) {
	r.handlers[method] = handler
}

// GetHandler retrieves the handler function for a given MCP method name.
//
// Summary: Retrieves the handler registered for a specific method.
//
// It looks up the method name in the router's internal map and returns the associated handler function.
//
// Parameters:
//   - method: string. The name of the MCP method.
//
// Returns:
//   - MethodHandler: The handler function if found.
//   - bool: A boolean indicating whether a handler was found (true) or not (false).
func (r *Router) GetHandler(method string) (MethodHandler, bool) {
	handler, ok := r.handlers[method]
	return handler, ok
}
