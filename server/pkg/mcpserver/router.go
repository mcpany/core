package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MethodHandler defines the signature for a function that handles an MCP method
// call. It takes a context and a request, and returns a result and an error.
type MethodHandler func(ctx context.Context, req mcp.Request) (mcp.Result, error)

// Router is responsible for mapping MCP method names to their corresponding
// handler functions. It provides a simple mechanism for registering and
// retrieving handlers.
type Router struct {
	handlers map[string]MethodHandler
}

// NewRouter creates and returns a new, empty Router.
//
// Returns:
//   - A pointer to a new Router.
func NewRouter() *Router {
	return &Router{
		handlers: make(map[string]MethodHandler),
	}
}

// Register associates a handler function with a specific MCP method name. If a
// handler for the method already exists, it will be overwritten.
//
// Parameters:
//   - method: The name of the MCP method (e.g., "tools/call").
//   - handler: The function that will handle the method call.
func (r *Router) Register(method string, handler MethodHandler) {
	r.handlers[method] = handler
}

// GetHandler retrieves the handler function for a given MCP method name.
//
// Parameters:
//   - method: The name of the MCP method.
//
// Returns:
//   - The handler function if found.
//   - A boolean indicating whether a handler was found.
func (r *Router) GetHandler(method string) (MethodHandler, bool) {
	handler, ok := r.handlers[method]
	return handler, ok
}
