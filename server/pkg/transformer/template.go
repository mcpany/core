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
// Summary: High-performance template engine using fasttemplate.
type TextTemplate struct {
	template *fasttemplate.Template
	raw      string
	startTag string
	endTag   string
	IsJSON   bool
}

// NewTemplate provides newtemplate functionality.
//
// Summary: NewTemplate.
//
// Parameters.
//   - templateString: The parameter.
//   - startTag: The parameter.
//   - endTag: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
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

// Render provides render functionality.
//
// Summary: Render.
//
// Parameters.
//   - params: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
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
