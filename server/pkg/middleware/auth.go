// Package middleware provides HTTP middleware for the application.

package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/consts"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// AuthMiddleware creates an MCP middleware for handling authentication. It is
// intended to inspect incoming requests and use the provided `AuthManager` to
// verify credentials before passing the request to the next handler.
//
// Parameters:
//   - authManager: The authentication manager to be used for authenticating
//     requests.
//
// Returns an `mcp.Middleware` function.
func AuthMiddleware(authManager *auth.Manager) mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			var serviceID string

			// Special handling for tool calls
			if method == consts.MethodToolsCall {
				if r, ok := req.(*mcp.CallToolRequest); ok {
					// We expect tool names to be prefixed with the service ID (e.g. "service.tool")
					parts := strings.SplitN(r.Params.Name, ".", 2)
					if len(parts) >= 2 {
						serviceID = parts[0]
					}
				}
			}

			// Fallback to method-based extraction if serviceID not yet found
			if serviceID == "" {
				// Extract serviceID from the method. Assuming the format is "service.method".
				parts := strings.SplitN(method, ".", 2)
				if len(parts) >= 2 {
					serviceID = parts[0]
				}
			}

			if serviceID == "" {
				// If we couldn't determine a service ID, we assume it's a global method or
				// a malformed request that doesn't map to a protected service.
				// We let it pass, relying on downstream handlers to validate if necessary.
				return next(ctx, method, req)
			}

			// If no authenticator is registered for this service, let the request through.
			if _, ok := authManager.GetAuthenticator(serviceID); !ok {
				return next(ctx, method, req)
			}

			// Extract the http.Request from the context.
			// The key "http.request" is not formally defined in the MCP spec, but it's a
			// de-facto standard used by the reference implementation.
			httpReq, ok := ctx.Value("http.request").(*http.Request)
			if !ok {
				// If the http.Request is not in the context, we cannot perform authentication.
				// This should ideally not happen in a server environment.
				return nil, fmt.Errorf("unauthorized: http.Request not found in context")
			}

			// Authenticate the request.
			newCtx, err := authManager.Authenticate(ctx, serviceID, httpReq)
			if err != nil {
				return nil, fmt.Errorf("unauthorized: %w", err)
			}

			// If authentication is successful, proceed to the next handler.
			return next(newCtx, method, req)
		}
	}
}
