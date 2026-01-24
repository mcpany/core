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
//
// config holds the configuration settings.
// log is the log.
//
// Returns the result.
func DLPMiddleware(config *configv1.DLPConfig, log *slog.Logger) mcp.Middleware {
	redactor := NewRedactor(config, log)
	if redactor == nil {
		return noOpMiddleware
	}

	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			// 1. Redact Request Inputs
			switch r := req.(type) {
			case *mcp.CallToolRequest:
				if r.Params != nil && r.Params.Arguments != nil {
					newBytes, err := redactor.RedactJSON(r.Params.Arguments)
					if err == nil {
						r.Params.Arguments = newBytes
					} else {
						log.Error("Failed to redact arguments", "error", err)
					}
				}
			case *mcp.GetPromptRequest:
				if r.Params != nil && r.Params.Arguments != nil {
					for k, v := range r.Params.Arguments {
						r.Params.Arguments[k] = redactor.RedactString(v)
					}
				}
			}

			// 2. Call Next
			result, err := next(ctx, method, req)
			if err != nil {
				return result, err
			}

			// 3. Redact Result Content
			switch r := result.(type) {
			case *mcp.CallToolResult:
				for _, content := range r.Content {
					if textContent, ok := content.(*mcp.TextContent); ok {
						redacted := redactor.RedactString(textContent.Text)
						if redacted != textContent.Text {
							textContent.Text = redacted
						}
					}
				}
			case *mcp.GetPromptResult:
				for _, msg := range r.Messages {
					if textContent, ok := msg.Content.(*mcp.TextContent); ok {
						redacted := redactor.RedactString(textContent.Text)
						if redacted != textContent.Text {
							textContent.Text = redacted
						}
					}
				}
			case *mcp.ReadResourceResult:
				for _, content := range r.Contents {
					redacted := redactor.RedactString(content.Text)
					if redacted != content.Text {
						content.Text = redacted
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
