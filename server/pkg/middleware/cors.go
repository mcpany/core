package middleware

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// CORSMiddleware creates an MCP middleware for handling Cross-Origin Resource
// Sharing (CORS). It is intended to add the necessary CORS headers to outgoing
// responses, allowing web browsers to securely make cross-origin requests to
// the MCP server.
//
// NOTE: This middleware is currently a placeholder and does not yet add any
// CORS headers. It passes all requests and responses through without
// modification.
//
// Returns an `mcp.Middleware` function.
func CORSMiddleware() mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			return next(ctx, method, req)
		}
	}
}
