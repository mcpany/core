// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// DebugMiddleware returns a middleware function that logs the full request and
// response of each MCP method call. This is useful for debugging and
// understanding the flow of data through the server.
func DebugMiddleware() mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			log := logging.GetLogger()
			debugEnabled := log.Enabled(ctx, slog.LevelDebug)

			if debugEnabled {
				reqBytes, err := json.Marshal(req)
				if err != nil {
					log.Error("Failed to marshal request for debugging", "error", err)
				} else {
					log.Debug("MCP Request", "method", method, "request", string(reqBytes))
				}
			}

			result, err := next(ctx, method, req)
			if err != nil {
				log.Error("MCP method failed", "method", method, "error", err)
				return nil, err
			}

			if debugEnabled {
				resBytes, err := json.Marshal(result)
				if err != nil {
					log.Error("Failed to marshal response for debugging", "error", err)
				} else {
					log.Debug("MCP Response", "method", method, "response", string(resBytes))
				}
			}

			return result, nil
		}
	}
}
