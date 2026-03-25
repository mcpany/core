// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MethodHandler defines the signature for a function that handles an MCP method call.
//
// Summary: Handler function signature for MCP methods.
//
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - req (mcp.Request): The request object.
//
// Returns:
//   - mcp.Result: The result of the operation.
// NewRouter creates and returns a new, empty Router.
//
// Summary: Creates a new Router instance.
//
// Register associates a handler function with a specific MCP method name.
//
// Summary: Registers a handler for an MCP method.
// GetHandler retrieves the handler function for a given MCP method name.
//
// Summary: Retrieves a handler for an MCP method.
//
// Parameters:
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//   - method (string): The name of the MCP method.
//
// Returns:
//   - MethodHandler: The handler function if found.
//   - bool: A boolean indicating whether a handler was found (true) or not (false).
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
//   - None.
// Side Effects:
//   - None.
// Errors:
//
// GetHandler retrieves the handler function for a given MCP method name.
// Summary: Registers a handler for an MCP method.
//
// Register associates a handler function with a specific MCP method name.
//
// Summary: Creates a new Router instance.
//
// NewRouter creates and returns a new, empty Router.
//   - mcp.Result: The result of the operation.
// Returns:
//
//   - req (mcp.Request): The request object.
//   - ctx (context.Context): The context for the request.
// Parameters:
//
// Summary: Handler function signature for MCP methods.
//
// MethodHandler defines the signature for a function that handles an MCP method call.
// Side Effects:
//   - None.
// Errors:
//   - triggers relevant error states on failure.
func (r *Router) GetHandler(method string) (MethodHandler, bool) {
	handler, ok := r.handlers[method]
	return handler, ok
}
//   - None.
// Side Effects:
//   - None.
// Errors:
// NewRouter creates and returns a new, empty Router.
//   - mcp.Result: The result of the operation.
// Returns:
//
//   - req (mcp.Request): The request object.
//   - ctx (context.Context): The context for the request.
// Parameters:
//
// Summary: Handler function signature for MCP methods.
//
// MethodHandler defines the signature for a function that handles an MCP method call.
