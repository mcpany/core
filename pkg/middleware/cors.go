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

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// CORSMiddleware creates an MCP middleware for handling Cross-Origin Resource
// Sharing (CORS). It is intended to add the necessary CORS headers to outgoing
// responses to allow browsers to securely make cross-origin requests.
//
// NOTE: This middleware is currently a placeholder and does not add any CORS
// headers. It passes all requests through to the next handler without
// modification.
func CORSMiddleware() mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			return next(ctx, method, req)
		}
	}
}
