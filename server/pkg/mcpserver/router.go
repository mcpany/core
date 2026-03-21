// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Summary: MethodHandler defines the signature for a function that handles an MCP method call. Handler function signature for MCP methods.
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
type MethodHandler func(ctx context.Context, req mcp.Request) (mcp.Result, error)

// Summary: Router is responsible for mapping MCP method names to their corresponding handler functions. Routes MCP requests to registered handlers.
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
type Router struct {
	handlers map[string]MethodHandler
}

// Summary: NewRouter creates and returns a new, empty Router. Creates a new Router instance.
//
// Parameters:
//   - None.
//
// Returns:
//   - *Router: The resulting *Router.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewRouter() *Router {
	return &Router{
		handlers: make(map[string]MethodHandler),
	}
}

// Summary: Register associates a handler function with a specific MCP method name. Registers a handler for an MCP method.
//
// Parameters:
//   - method (string): The method parameter.
//   - handler (MethodHandler): The handler parameter.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (r *Router) Register(method string, handler MethodHandler) {
	r.handlers[method] = handler
}

// Summary: GetHandler retrieves the handler function for a given MCP method name. Retrieves a handler for an MCP method.
//
// Parameters:
//   - method (string): The method parameter.
//
// Returns:
//   - MethodHandler: The resulting MethodHandler.
//   - bool: The resulting bool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (r *Router) GetHandler(method string) (MethodHandler, bool) {
	handler, ok := r.handlers[method]
	return handler, ok
}
