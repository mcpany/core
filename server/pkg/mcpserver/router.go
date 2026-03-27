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
// Parameters.
//   - ctx (context.Context): The context for the request.
//   - req (mcp.Request): The request object.
//
// Returns.
//   - mcp.Result: The result of the operation.
//   - error: An error if the operation fails.
type MethodHandler func(ctx context.Context, req mcp.Request) (mcp.Result, error)

// Router is responsible for mapping MCP method names to their corresponding handler functions.
//
// Summary. Routes MCP requests to registered handlers.
//
// Side Effects.
//   - Stores handlers in an internal map.
type Router struct {
	handlers map[string]MethodHandler
}

// NewRouter provides newrouter functionality.
//
// Summary: NewRouter.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func NewRouter() *Router {
	return &Router{
		handlers: make(map[string]MethodHandler),
	}
}

// Register provides register functionality.
//
// Summary: Register.
//
// Parameters.
//   - method: The parameter.
//   - handler: The parameter.
//
// Returns.
//   - None.
func (r *Router) Register(method string, handler MethodHandler) {
	r.handlers[method] = handler
}

// GetHandler provides gethandler functionality.
//
// Summary: GetHandler.
//
// Parameters.
//   - method: The parameter.
//   - bool: The parameter.
//
// Returns.
//   - None.
func (r *Router) GetHandler(method string) (MethodHandler, bool) {
	handler, ok := r.handlers[method]
	return handler, ok
}
