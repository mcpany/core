// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// CORSMiddleware creates a placeholder CORS middleware for MCP.
//
// Summary: Returns a no-op MCP middleware, as CORS is handled at the HTTP layer.
//
// Returns:
//   - mcp.Middleware: A pass-through MCP middleware.
func CORSMiddleware() mcp.Middleware {
	// Log a warning once when the middleware is created to inform the user.
	// This helps avoid confusion if they expect this middleware to handle HTTP CORS.
	logging.GetLogger().Warn("CORSMiddleware (MCP) is a placeholder and does not handle HTTP CORS. Ensure HTTP CORS is configured via the HTTP server middleware (HTTPCORSMiddleware).")

	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			return next(ctx, method, req)
		}
	}
}
