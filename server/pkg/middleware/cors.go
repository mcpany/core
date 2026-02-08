// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// CORSMiddleware creates an MCP middleware for handling Cross-Origin Resource Sharing (CORS).
//
// Summary: A placeholder middleware for MCP-level CORS handling. HTTP CORS is handled by HTTPCORSMiddleware.
//
// Returns:
//   - mcp.Middleware: The CORS middleware function.
func CORSMiddleware() mcp.Middleware {
	// Note: This is a placeholder. Real CORS is handled at the HTTP transport layer.
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			return next(ctx, method, req)
		}
	}
}
