
package middleware

import (
	"context"

	"github.com/mcpany/core/pkg/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// AuthMiddleware creates an MCP middleware for handling authentication. It is
// intended to inspect incoming requests and use the provided `AuthManager` to
// verify credentials before passing the request to the next handler.
//
// NOTE: This middleware is currently a placeholder and does not yet perform any
// authentication. It passes all requests through to the next handler without
// modification.
//
// Parameters:
//   - authManager: The authentication manager to be used for authenticating
//     requests.
//
// Returns an `mcp.Middleware` function.
func AuthMiddleware(authManager *auth.AuthManager) mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			return next(ctx, method, req)
		}
	}
}
