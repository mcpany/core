// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"regexp"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	// Common PII patterns
	emailRegex      = regexp.MustCompile(`(?i)[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}`)
	creditCardRegex = regexp.MustCompile(`(?:\d{4}[-\s]?){3}\d{4}`)
	ssnRegex        = regexp.MustCompile(`\d{3}-\d{2}-\d{4}`)

	redactedStr = "***REDACTED***"
)

// DLPMiddleware creates a middleware that redacts PII from request arguments and result content.
func DLPMiddleware(config *configv1.DLPConfig, log *slog.Logger) mcp.Middleware {
	if config == nil || config.Enabled == nil || !*config.Enabled {
		return noOpMiddleware
	}

	patterns := []*regexp.Regexp{
		emailRegex,
		creditCardRegex,
		ssnRegex,
	}

	for _, p := range config.CustomPatterns {
		if r, err := regexp.Compile(p); err == nil {
			patterns = append(patterns, r)
		} else {
			log.Warn("Invalid custom DLP pattern, ignoring", "pattern", p, "error", err)
		}
	}

	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			// 1. Redact Tool Call Arguments
			if callReq, ok := req.(*mcp.CallToolRequest); ok && callReq.Params != nil && callReq.Params.Arguments != nil {
				// Arguments is json.RawMessage. We must unmarshal -> redact -> marshal.
				var args map[string]interface{}
				// Try unmarshaling as map
				if err := json.Unmarshal(callReq.Params.Arguments, &args); err == nil {
					redactStruct(args, patterns)
					if newBytes, err := json.Marshal(args); err == nil {
						callReq.Params.Arguments = newBytes
					} else {
						log.Error("Failed to marshal redacted arguments", "error", err)
					}
				} else {
					// Fallback: try unmarshaling as generic interface if it's not a map (unlikely for arguments but possible)
					var genericArgs interface{}
					if err := json.Unmarshal(callReq.Params.Arguments, &genericArgs); err == nil {
						redacted := redactValue(genericArgs, patterns)
						if newBytes, err := json.Marshal(redacted); err == nil {
							callReq.Params.Arguments = newBytes
						}
					}
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
						redacted := redactString(textContent.Text, patterns)
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

func redactString(s string, patterns []*regexp.Regexp) string {
	res := s
	for _, p := range patterns {
		res = p.ReplaceAllString(res, redactedStr)
	}
	return res
}

func redactStruct(v map[string]interface{}, patterns []*regexp.Regexp) {
	for k, val := range v {
		v[k] = redactValue(val, patterns)
	}
}

func redactValue(val interface{}, patterns []*regexp.Regexp) interface{} {
	switch v := val.(type) {
	case string:
		return redactString(v, patterns)
	case map[string]interface{}:
		redactStruct(v, patterns)
		return v
	case []interface{}:
		for i, item := range v {
			v[i] = redactValue(item, patterns)
		}
		return v
	case *structpb.Value:
		return val
	default:
		return val
	}
}
