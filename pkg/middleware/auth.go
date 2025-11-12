/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
import (
	"fmt"
	"net/http"
	"strings"
)

// Returns an `mcp.Middleware` function.
func AuthMiddleware(authManager *auth.AuthManager) mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			// Extract serviceID from the method. Assuming the format is "service.method".
			parts := strings.SplitN(method, ".", 2)
			if len(parts) < 2 {
				// If the method format is invalid, we might want to return an error
				// or let it pass, depending on the desired behavior for malformed requests.
				// For now, we'll let it pass.
				return next(ctx, method, req)
			}
			serviceID := parts[0]

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
			_, err := authManager.Authenticate(ctx, serviceID, httpReq)
			if err != nil {
				return nil, fmt.Errorf("unauthorized: %w", err)
			}

			// If authentication is successful, proceed to the next handler.
			return next(ctx, method, req)
		}
	}
}
