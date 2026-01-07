// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"log/slog"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// DLPMiddleware creates a middleware that redacts PII from request arguments and result content.
func DLPMiddleware(config *configv1.DLPConfig, log *slog.Logger) mcp.Middleware {
	redactor := NewRedactor(config, log)
	if redactor == nil {
		return noOpMiddleware
	}

	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			// 1. Redact Tool Call Arguments
			if callReq, ok := req.(*mcp.CallToolRequest); ok && callReq.Params != nil && callReq.Params.Arguments != nil {
				newBytes, err := redactor.RedactJSON(callReq.Params.Arguments)
				if err == nil {
					callReq.Params.Arguments = newBytes
				} else {
					log.Error("Failed to redact arguments", "error", err)
				}
			}

			// 2. Call Next
			result, err := next(ctx, method, req)
			if err != nil {
				return result, err
			}

			// 3. Redact Result Content
			if callResult, ok := result.(*mcp.CallToolResult); ok {
				for _, content := range callResult.Content {
					if textContent, ok := content.(*mcp.TextContent); ok {
						redacted := redactor.RedactString(textContent.Text)
						if redacted != textContent.Text {
							textContent.Text = redacted
						}
					}
				}
			}

			return result, nil
		}
	}
}

func noOpMiddleware(next mcp.MethodHandler) mcp.MethodHandler {
	return next
}
