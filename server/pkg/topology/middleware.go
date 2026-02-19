// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package topology

import (
	"context"
	"time"

	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/consts"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Middleware returns a middleware function to track session activity.
//
// next is the next.
//
// Returns the result.
func (m *Manager) Middleware(next mcp.MethodHandler) mcp.MethodHandler {
	return func(
		ctx context.Context,
		method string,
		req mcp.Request,
	) (mcp.Result, error) {
		// Identify Client Session
		// In standard MCP, the JSON-RPC connection usually corresponds to a session.
		// `req.GetSession()` from SDK likely returns the connection/session object.
		// However, we want a persistent ID if possible.
		// If using `mcp.StdioTransport`, there might be one session.
		// If using `mcp.SSEServerTransport`, session ID exists.

		sessionID := "unknown"
		meta := make(map[string]interface{})

		// Extract session ID
		// The SDK Request object has GetSession() but it returns an interface.
		// We'll rely on our auth context or custom logic if available.
		// For now, let's use a "default" session if not found, or try to cast.

		// Fallback: Check Auth Context (API Key / User ID)
		if uid, ok := auth.UserFromContext(ctx); ok {
			sessionID = "user-" + uid
			meta["type"] = "authenticated_user"
		} else {
			// Check for generic remote address
			if ip, ok := ctx.Value(consts.ContextKeyRemoteAddr).(string); ok {
				sessionID = "ip-" + ip
				meta["type"] = "anonymous_ip"
			}
		}

		start := time.Now()
		res, err := next(ctx, method, req)
		duration := time.Since(start)

		isError := err != nil
		// We could check specific result types for application-level errors (e.g. CallToolResult.IsError)
		// but standardizing on error return is safer without knowing the exact SDK interface.

		serviceID := ""
		if method == "tools/call" {
			if callReq, ok := req.(*mcp.CallToolRequest); ok {
				toolName := callReq.Params.Name
				if t, found := m.toolManager.GetTool(toolName); found {
					serviceID = t.Tool().GetServiceId()
				}
			}
		}

		m.RecordActivity(sessionID, meta, duration, isError, serviceID)

		return res, err
	}
}
