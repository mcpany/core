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
// It intercepts both incoming requests (redacting inputs) and outgoing responses (redacting outputs)
// based on the configured DLP rules (e.g., regex patterns, info types).
//
// Parameters:
//   - config: Configuration settings for the DLP engine (rules, likelihood thresholds).
//   - log: Logger instance for recording DLP actions or errors.
//
// Returns:
//   - mcp.Middleware: A middleware function that wraps the method handler.
//
// Side Effects:
//   - Modifies request arguments in-place.
//   - Modifies response content in-place.
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
