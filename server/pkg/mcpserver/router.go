// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MethodHandler defines the signature for a function that handles an MCP method call.
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
// Side Effects:
//   - Stores handlers in an internal map.
type Router struct {
	handlers map[string]MethodHandler
}

// NewRouter creates and returns a new, empty Router.
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
// Parameters:
//   - method: string. The name of the MCP method (e.g., "tools/call").
//   - handler: MethodHandler. The function that will handle the method call.
//
// Side Effects:
//   - If a handler for the method already exists, it will be overwritten.
func (r *Router) Register(method string, handler MethodHandler) {
	r.handlers[method] = handler
}

// GetHandler retrieves the handler function for a given MCP method name.
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
