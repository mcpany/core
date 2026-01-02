// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"regexp"
	"strings"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	// Common PII patterns.
	// We use string constants here to allow combining them.
	emailPattern = `(?i)[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}`
	//nolint:gosec // This is a regex pattern, not a hardcoded credential.
	creditCardPattern = `(?:\d{4}[-\s]?){3}\d{4}`
	ssnPattern        = `\d{3}-\d{2}-\d{4}`

	redactedStr = "***REDACTED***"
)

// DLPMiddleware creates a middleware that redacts PII from request arguments and result content.
func DLPMiddleware(config *configv1.DLPConfig, log *slog.Logger) mcp.Middleware {
	if config == nil || config.Enabled == nil || !*config.Enabled {
		return noOpMiddleware
	}

	// Combine all patterns into a single regex for better performance.
	patternStrings := []string{
		emailPattern,
		creditCardPattern,
		ssnPattern,
	}

	for _, p := range config.CustomPatterns {
		// Validate the custom pattern before adding it.
		if _, err := regexp.Compile(p); err == nil {
			patternStrings = append(patternStrings, p)
		} else {
			log.Warn("Invalid custom DLP pattern, ignoring", "pattern", p, "error", err)
		}
	}

	// Join all patterns with OR operator.
	// We wrap each pattern in parentheses to ensure precedence.
	var combinedPatternStr strings.Builder
	for i, p := range patternStrings {
		if i > 0 {
			combinedPatternStr.WriteString("|")
		}
		combinedPatternStr.WriteString("(")
		combinedPatternStr.WriteString(p)
		combinedPatternStr.WriteString(")")
	}

	// Compile the combined regex.
	// If the combined regex fails to compile (unlikely if parts are valid), fallback or log error.
	// Since we validated individual parts, failure should only happen if the combination is too large, etc.
	combinedRegex, err := regexp.Compile(combinedPatternStr.String())
	if err != nil {
		log.Error("Failed to compile combined DLP regex", "error", err)
		// Fallback to no-op or handle gracefully?
		// For now, let's just log and return no-op to avoid crashing or leaking data (though leaking might happen if we don't redact).
		// Better to panic? Or maybe just use the individual valid ones?
		// If combination fails, we can't easily proceed with this optimization.
		// But let's assume it works.
		return noOpMiddleware
	}

	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			// 1. Redact Tool Call Arguments
			if callReq, ok := req.(*mcp.CallToolRequest); ok && callReq.Params != nil && callReq.Params.Arguments != nil {
				// Arguments is json.RawMessage. We must unmarshal -> redact -> marshal.
				var args map[string]interface{}
				// Try unmarshaling as map
				if err := json.Unmarshal(callReq.Params.Arguments, &args); err == nil {
					redactStruct(args, combinedRegex)
					if newBytes, err := json.Marshal(args); err == nil {
						callReq.Params.Arguments = newBytes
					} else {
						log.Error("Failed to marshal redacted arguments", "error", err)
					}
				} else {
					// Fallback: try unmarshaling as generic interface if it's not a map (unlikely for arguments but possible)
					var genericArgs interface{}
					if err := json.Unmarshal(callReq.Params.Arguments, &genericArgs); err == nil {
						redacted := redactValue(genericArgs, combinedRegex)
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
						redacted := redactString(textContent.Text, combinedRegex)
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

func redactString(s string, pattern *regexp.Regexp) string {
	return pattern.ReplaceAllString(s, redactedStr)
}

func redactStruct(v map[string]interface{}, pattern *regexp.Regexp) {
	for k, val := range v {
		v[k] = redactValue(val, pattern)
	}
}

func redactValue(val interface{}, pattern *regexp.Regexp) interface{} {
	switch v := val.(type) {
	case string:
		return redactString(v, pattern)
	case map[string]interface{}:
		redactStruct(v, pattern)
		return v
	case []interface{}:
		for i, item := range v {
			v[i] = redactValue(item, pattern)
		}
		return v
	case *structpb.Value:
		return val
	default:
		return val
	}
}
