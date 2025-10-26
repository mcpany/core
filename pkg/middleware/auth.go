/*
 * Copyright 2025 Author(s) of MCP-XY
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
	"net/http"
	"strings"

	"github.com/mcpxy/core/pkg/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"golang.org/x/oauth2"
)

// AuthMiddleware creates an MCP middleware for handling authentication.
func AuthMiddleware(authManager *auth.Manager) mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			if !authManager.IsEnabled() {
				return next(ctx, method, req)
			}

			transport, ok := mcp.TransportFromContext(ctx)
			if !ok {
				return nil, &mcp.Error{Code: mcp.ErrorCodeInternalError, Message: "transport not found in context"}
			}

			httpTransport, ok := transport.(mcp.HTTPTransport)
			if !ok {
				return nil, &mcp.Error{Code: mcp.ErrorCodeInternalError, Message: "transport is not HTTP"}
			}

			authHeader := httpTransport.RequestHeader().Get("Authorization")
			if authHeader == "" {
				return nil, &mcp.Error{Code: mcp.ErrorCodeInvalidRequest, Message: "missing Authorization header"}
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return nil, &mcp.Error{Code: mcp.ErrorCodeInvalidRequest, Message: "invalid Authorization header"}
			}
			token := parts[1]

			idToken, err := authManager.VerifyToken(ctx, token)
			if err != nil {
				return nil, &mcp.Error{Code: mcp.ErrorCodeInvalidRequest, Message: "invalid token"}
			}

			// Add the token and user info to the context
			ctx = context.WithValue(ctx, "id_token", idToken)
			ctx = context.WithValue(ctx, "access_token", token)

			// Add an authenticated http client to the context
			ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
			httpClient := oauth2.NewClient(ctx, ts)
			ctx = context.WithValue(ctx, "http_client", httpClient)

			return next(ctx, method, req)
		}
	}
}
