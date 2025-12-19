// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package middleware provides HTTP middleware for the application.
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
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

// AuthenticationMiddleware is a tool middleware that enforces authentication.
type AuthenticationMiddleware struct {
	config *configv1.AuthenticationConfig
}

// NewAuthenticationMiddleware creates a new AuthenticationMiddleware.
func NewAuthenticationMiddleware(config *configv1.AuthenticationConfig) *AuthenticationMiddleware {
	return &AuthenticationMiddleware{config: config}
}

// Execute enforces authentication before proceeding to the next handler.
func (m *AuthenticationMiddleware) Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error) {
	httpReq, ok := ctx.Value("http.request").(*http.Request)
	if !ok {
		// If we are not in an HTTP context (e.g. stdio), authentication strategy might differ.
		// For now, fail if auth is required but no request present.
		return nil, fmt.Errorf("authentication middleware: http request not found in context")
	}

	if err := auth.ValidateAuthentication(ctx, m.config, httpReq); err != nil {
		return nil, fmt.Errorf("unauthorized: %w", err)
	}
	return next(ctx, req)
}
