// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/valyala/fasttemplate"
)

// TextTemplate provides a simple wrapper around Go's standard text/template
// for rendering strings with dynamic data.
type TextTemplate struct {
	template *fasttemplate.Template
	raw      string
	startTag string
	endTag   string
}

// NewTemplate parses a template string and creates a new TextTemplate.
//
// templateString is the template content to be parsed.
// It returns a new TextTemplate or an error if the template string is invalid.
func NewTemplate(templateString, startTag, endTag string) (*TextTemplate, error) {
	tpl, err := fasttemplate.NewTemplate(templateString, startTag, endTag)
	if err != nil {
		return nil, err
	}
	return &TextTemplate{
		template: tpl,
		raw:      templateString,
		startTag: startTag,
		endTag:   endTag,
	}, nil
}

// Render executes the template with the provided parameters and returns the
// resulting string.
//
// params is a map of key-value pairs that will be available within the
// template.
// It returns the rendered string or an error if the template execution fails.
func (t *TextTemplate) Render(params map[string]any) (string, error) {
	return t.template.ExecuteFuncStringWithErr(func(w io.Writer, tag string) (int, error) {
		val, ok := params[tag]
		if !ok {
			return 0, fmt.Errorf("missing key: %s", tag)
		}
		if s, ok := val.(string); ok {
			return io.WriteString(w, s)
		}
		return fmt.Fprintf(w, "%v", val)
	})
}

// RenderSafe executes the template with context-aware escaping.
// If the template appears to be JSON, it will enforce JSON string escaping
// for values substituted inside JSON strings.
func (t *TextTemplate) RenderSafe(params map[string]any) (string, error) {
	trimmed := strings.TrimSpace(t.raw)
	// Check if it looks like a JSON object or array
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		return t.renderJSON(params)
	}
	return t.Render(params)
}

func (t *TextTemplate) renderJSON(params map[string]any) (string, error) {
	var sb strings.Builder
	input := t.raw

	inString := false
	isEscaped := false

	startLen := len(t.startTag)
	endLen := len(t.endTag)

	for i := 0; i < len(input); {
		if strings.HasPrefix(input[i:], t.startTag) {
			// Found tag start
			end := strings.Index(input[i+startLen:], t.endTag)
			if end == -1 {
				// No end tag, just write current char and continue
				sb.WriteByte(input[i])
				i++
				continue
			}

			// end is relative to i+startLen
			tagName := input[i+startLen : i+startLen+end]
			tagName = strings.TrimSpace(tagName)

			val, ok := params[tagName]
			if !ok {
				return "", fmt.Errorf("missing key: %s", tagName)
			}

			// Convert value to string (handling non-string types same as Render)
			var valStr string
			if s, ok := val.(string); ok {
				valStr = s
			} else {
				valStr = fmt.Sprintf("%v", val)
			}

			if inString {
				// We are inside a JSON string. We must escape the value.
				// json.Marshal returns "escaped_string" (with quotes).
				marshaled, err := json.Marshal(valStr)
				if err != nil {
					return "", fmt.Errorf("failed to marshal value for key %s: %w", tagName, err)
				}
				s := string(marshaled)
				// Trim outer quotes to inject into the existing string
				if len(s) >= 2 {
					s = s[1 : len(s)-1]
				}
				sb.WriteString(s)
			} else {
				// Outside string, just write raw
				sb.WriteString(valStr)
			}

			i += startLen + end + endLen
			continue
		}

		char := input[i]
		sb.WriteByte(char)

		if isEscaped {
			isEscaped = false
		} else {
			if char == '\\' {
				isEscaped = true
			} else if char == '"' {
				inString = !inString
			}
		}
		i++
	}

	return sb.String(), nil
}
