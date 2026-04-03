// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MethodHandler represents the public MethodHandler entity.
//
// Summary: Defines the structured data model representing a handler.
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

// Router represents the public Router entity.
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
type Router struct {
	handlers map[string]MethodHandler
}

// NewRouter serves as a public interface for interacting with NewRouter.
//
// Summary: Constructs and returns an initialized router ready for consumption.
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
func NewRouter() *Router {
	return &Router{
		handlers: make(map[string]MethodHandler),
	}
}

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
func (r *Router) Register(method string, handler MethodHandler) {
	r.handlers[method] = handler
}

// GetHandler serves as a public interface for interacting with GetHandler.
//
// Summary: Fetches and returns the underlying handler from the system state.
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
func (r *Router) GetHandler(method string) (MethodHandler, bool) {
	handler, ok := r.handlers[method]
	return handler, ok
}
