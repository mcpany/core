// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Summary: MethodHandler represents a data structure.
//
// Parameters:
//   - None
//
// Returns:
//   - None
//
// Errors:
//   - None
//
// Side Effects:
//   - None
type MethodHandler func(ctx context.Context, req mcp.Request) (mcp.Result, error)

// Router is responsible for mapping MCP method names to their corresponding handler functions.
//
// Summary: Routes MCP requests to registered handlers.
//
// Side Effects:
//   - Stores handlers in an internal map.
// Summary: Router represents a data structure.
//
// Parameters:
//   - None
//
// Returns:
//   - None
//
// Errors:
//   - None
//
// Side Effects:
//   - None
type Router struct {
	handlers map[string]MethodHandler
}

// NewRouter creates and returns a new, empty Router.
//
// Summary: Creates a new Router instance.
//
// Parameters:
//   - None.
//
// Returns:
//   - *Router: A new, initialized Router.
// Summary: NewRouter executes the operation.
//
// Parameters:
//   - None
//
// Returns:
//   - *Router {
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func NewRouter() *Router {
	return &Router{
		handlers: make(map[string]MethodHandler),
	}
}

// Register associates a handler function with a specific MCP method name.
//
// Summary: Registers a handler for an MCP method.
//
// Parameters:
//   - method (string): The method name.
//   - handler (MethodHandler): The handler function.
//
// Returns:
//   - None.
//
// Summary: Register executes the operation.
//
// Parameters:
//   - method string: Input parameter.
//   - handler MethodHandler: Input parameter.
//
// Returns:
//   - {
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func (r *Router) Register(method string, handler MethodHandler) {
	r.handlers[method] = handler
}

// GetHandler retrieves the handler function for a given MCP method name.
//
// Summary: Retrieves a handler for an MCP method.
//
// Parameters:
//   - method (string): The name of the MCP method.
//
// Returns:
//   - MethodHandler: The handler function if found.
//   - bool: A boolean indicating whether a handler was found (true) or not (false).
//
// Side Effects:
// Summary: GetHandler executes the operation.
//
// Parameters:
//   - method string: Input parameter.
//
// Returns:
//   - (MethodHandler, bool): Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func (r *Router) GetHandler(method string) (MethodHandler, bool) {
	handler, ok := r.handlers[method]
	return handler, ok
}
