// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MethodHandler defines the signature for a function that handles an MCP method call.
//
// Summary: A function type that processes an MCP request and returns a result.
//
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - req (mcp.Request): The request object containing parameters.
//
// Returns:
//   - mcp.Result: The result of the operation.
//   - error: An error if the operation fails.
type MethodHandler func(ctx context.Context, req mcp.Request) (mcp.Result, error)

// Router is responsible for mapping MCP method names to their corresponding handler functions.
//
// Summary: Manages the dispatching of MCP requests to registered handlers.
//
// Fields:
//   - handlers: A map where keys are method names (e.g., "tools/call") and values are handler functions.
//
// Side Effects:
//   - Stores handlers in an internal map.
type Router struct {
	handlers map[string]MethodHandler
}

// NewRouter creates and returns a new, empty Router.
//
// Summary: Initializes a new Router instance with an empty handler map.
//
// Returns:
//   - *Router: A pointer to a new, initialized Router.
//
// Side Effects:
//   - Allocates memory for the handlers map.
func NewRouter() *Router {
	return &Router{
		handlers: make(map[string]MethodHandler),
	}
}

// Register associates a handler function with a specific MCP method name.
//
// Summary: Adds a new route to the router.
//
// Parameters:
//   - method (string): The name of the MCP method (e.g., "tools/call").
//   - handler (MethodHandler): The function that will handle the method call.
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
// Summary: Lookups a handler by method name.
//
// Parameters:
//   - method (string): The name of the MCP method.
//
// Returns:
//   - MethodHandler: The handler function if found.
//   - bool: A boolean indicating whether a handler was found (true) or not (false).
func (r *Router) GetHandler(method string) (MethodHandler, bool) {
	handler, ok := r.handlers[method]
	return handler, ok
}
