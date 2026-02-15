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
//
// Summary: A template engine wrapper for text transformation.
type TextTemplate struct {
	template *fasttemplate.Template
	raw      string
	startTag string
	endTag   string
	// IsJSON indicates if the template is detected as a JSON structure, enabling JSON-specific escaping.
	IsJSON bool
}

// NewTemplate parses a template string and creates a new TextTemplate.
//
// Summary: Creates a new template instance.
//
// Parameters:
//   - templateString: The raw template string to be parsed.
//   - startTag: The delimiter for the start of a variable (e.g., "{{").
//   - endTag: The delimiter for the end of a variable (e.g., "}}").
//
// Returns:
//   - *TextTemplate: The parsed template instance.
//   - error: An error if the template string is invalid.
func NewTemplate(templateString, startTag, endTag string) (*TextTemplate, error) {
	tpl, err := fasttemplate.NewTemplate(templateString, startTag, endTag)
	if err != nil {
		return nil, err
	}

	trimmed := strings.TrimSpace(templateString)
	// Heuristic detection of JSON:
	// 1. Must start with { and end with } (Object) OR start with [ and end with ] (Array)
	// 2. Must NOT start with the startTag (to avoid misidentifying "{{ var }}" as JSON)
	isObject := strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")
	isArray := strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")

	isJSON := (isObject || isArray) && !strings.HasPrefix(trimmed, startTag)

	return &TextTemplate{
		template: tpl,
		raw:      templateString,
		startTag: startTag,
		endTag:   endTag,
		IsJSON:   isJSON,
	}, nil
}

// Render executes the template with the provided parameters and returns the
// resulting string.
//
// Summary: Renders the template with the given parameters.
//
// Parameters:
//   - params: A map of key-value pairs to substitute into the template.
//
// Returns:
//   - string: The rendered string.
//   - error: An error if a parameter is missing or rendering fails.
func (t *TextTemplate) Render(params map[string]any) (string, error) {
	return t.template.ExecuteFuncStringWithErr(func(w io.Writer, tag string) (int, error) {
		val, ok := params[tag]
		if !ok {
			return 0, fmt.Errorf("missing key: %s", tag)
		}

		if t.IsJSON {
			if s, ok := val.(string); ok {
				return io.WriteString(w, escapeJSONString(s))
			}
			b, err := json.Marshal(val)
			if err != nil {
				return fmt.Fprintf(w, "%v", val)
			}
			return w.Write(b)
		}

		if s, ok := val.(string); ok {
			return io.WriteString(w, s)
		}
		return fmt.Fprintf(w, "%v", val)
	})
}

func escapeJSONString(s string) string {
	b, _ := json.Marshal(s)
	if len(b) >= 2 {
		return string(b[1 : len(b)-1])
	}
	return s
}
