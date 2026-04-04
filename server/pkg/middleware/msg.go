// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MetadataSanitizationGateway implements the MSG specification for scrubbing
// malicious instructions from agent-ingested external metadata.
// MetadataSanitizationGateway scrubs sensitive metadata from tool requests and responses.
//
// Summary: Protects sensitive system metadata from leaking to downstream components.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Throws/Errors:
//   - None.
type MetadataSanitizationGateway struct {
	config *configv1.MetadataSanitizationConfig
	// Pre-compiled regular expressions for speed
	imperativeRules []*regexp.Regexp
}

// NewMetadataSanitizationGateway initializes a new gateway instance.
//
// Summary: Creates a new MetadataSanitizationGateway ready for use.
//
// Parameters:
//   - cfg (*configv1.MetadataSanitizationConfig): The configuration.
//
// Returns:
//   - *MetadataSanitizationGateway: The initialized gateway.
//
// Throws/Errors:
//   - None.
func NewMetadataSanitizationGateway(cfg *configv1.MetadataSanitizationConfig) *MetadataSanitizationGateway {
	if cfg == nil {
		cfg = configv1.MetadataSanitizationConfig_builder{}.Build()
	}

	msg := &MetadataSanitizationGateway{
		config: cfg,
	}

	// Compile semantic rules if enabled
	if cfg.GetEnabled() {
		for _, rule := range cfg.GetImperativePatterns() {
			if r, err := regexp.Compile(rule); err == nil {
				msg.imperativeRules = append(msg.imperativeRules, r)
			}
		}

		// If no rules are provided, add some default semantic boundaries
		if len(msg.imperativeRules) == 0 {
			defaultRules := []string{
				`(?i)\b(ignore previous instructions)\b`,
				`(?i)\b(system prompt)\b`,
				`(?i)\b(you are now)\b`,
				`(?i)\b(forget everything)\b`,
				`(?i)\b(execute this code)\b`,
			}
			for _, rule := range defaultRules {
				if r, err := regexp.Compile(rule); err == nil {
					msg.imperativeRules = append(msg.imperativeRules, r)
				}
			}
		}
	}

	return msg
}

// Middleware returns an MCP handler that wraps the next handler in the chain.
//
// Summary: Applies the sanitization rules to the incoming request and outgoing response.
//
// Parameters:
//   - None.
//
// Returns:
//   - mcp.Middleware: The middleware constructor.
//
// Throws/Errors:
//   - None.
func (m *MetadataSanitizationGateway) Middleware() func(mcp.MethodHandler) mcp.MethodHandler {
	if !m.config.GetEnabled() {
		return noOpMiddleware
	}

	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			res, err := next(ctx, method, req)
			if err != nil {
				return nil, err
			}

			if callRes, ok := res.(*mcp.CallToolResult); ok && callRes != nil {
				for i, contentItem := range callRes.Content {
					if textContent, ok := contentItem.(*mcp.TextContent); ok {
						sanitizedRes, err := m.sanitizeResult(textContent.Text)
						if err == nil {
							if sanitizedString, ok := sanitizedRes.(string); ok && sanitizedRes != nil {
								textContent.Text = sanitizedString
								callRes.Content[i] = textContent
							}
						} else {
							return nil, fmt.Errorf("MSG sanitization failed: %w", err)
						}
					}
				}
			}

			return res, nil
		}
	}
}

// sanitizeResult recursively traverses the result and sanitizes string values
func (m *MetadataSanitizationGateway) sanitizeResult(res any) (any, error) {
	if res == nil {
		return nil, nil
	}

	switch v := res.(type) {
	case string:
		return m.sanitizeString(v), nil
	case []byte:
		// Attempt to parse as JSON first, if not, sanitize as string
		var jsonObj any
		if err := json.Unmarshal(v, &jsonObj); err == nil {
			sanitizedObj, err := m.sanitizeResult(jsonObj)
			if err != nil {
				return nil, err
			}
			return json.Marshal(sanitizedObj)
		}
		return []byte(m.sanitizeString(string(v))), nil
	case map[string]any:
		sanitizedMap := make(map[string]any)
		for key, val := range v {
			// Sanitize the value
			sanitizedVal, err := m.sanitizeResult(val)
			if err != nil {
				return nil, err
			}
			sanitizedMap[key] = sanitizedVal
		}
		return sanitizedMap, nil
	case []any:
		sanitizedSlice := make([]any, len(v))
		for i, val := range v {
			sanitizedVal, err := m.sanitizeResult(val)
			if err != nil {
				return nil, err
			}
			sanitizedSlice[i] = sanitizedVal
		}
		return sanitizedSlice, nil
	default:
		// For other types (int, float, bool, etc.), no semantic sanitization is needed
		return res, nil
	}
}

// sanitizeString applies the imperative rules to scrub malicious instructions
func (m *MetadataSanitizationGateway) sanitizeString(input string) string {
	sanitized := input
	redactedText := "[REDACTED_BY_MSG]"

	if m.config.GetRedactionText() != "" {
		redactedText = m.config.GetRedactionText()
	}

	for _, rule := range m.imperativeRules {
		sanitized = rule.ReplaceAllString(sanitized, redactedText)
	}

	return sanitized
}
