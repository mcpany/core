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

// TextTemplate provides a simple wrapper around Go's standard text/template for rendering strings with dynamic data. Summary: High-performance template engine using fasttemplate.
//
// Summary: TextTemplate provides a simple wrapper around Go's standard text/template for rendering strings with dynamic data. Summary: High-performance template engine using fasttemplate.
//
// Fields:
//   - Contains the configuration and state properties required for TextTemplate functionality.
type TextTemplate struct {
	template *fasttemplate.Template
	raw      string
	startTag string
	endTag   string
	IsJSON   bool
}

// NewTemplate parses a template string and creates a new TextTemplate.
//
// Summary: Initializes a new TextTemplate.
//
// Parameters:
//   - templateString: string. The template source.
//   - startTag: string. The start delimiter (e.g. "{{").
//   - endTag: string. The end delimiter (e.g. "}}").
//
// Returns:
//   - *TextTemplate: The parsed template.
//   - error: An error if parsing fails.
//
// Side Effects:
//   - Auto-detects if the template output is likely JSON to enable automatic escaping.
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
// Summary: Renders the template with data.
//
// Parameters:
//   - params: map[string]any. The data map for variable substitution.
//
// Returns:
//   - string: The rendered output.
//   - error: An error if a key is missing or rendering fails.
//
// Errors:
//   - Returns error if a required tag is missing in params.
//
// Side Effects:
//   - Automatically escapes strings if the template is detected as JSON.
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
