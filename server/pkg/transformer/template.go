// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/valyala/fasttemplate"
)

// TemplateFormat defines the output format for escaping.
type TemplateFormat int

const (
	TemplateFormatUnspecified TemplateFormat = 0
	TemplateFormatJSON        TemplateFormat = 1
	TemplateFormatXML         TemplateFormat = 2
	TemplateFormatYAML        TemplateFormat = 3
)

// TextTemplate provides a simple wrapper around Go's standard text/template
// for rendering strings with dynamic data.
type TextTemplate struct {
	template *fasttemplate.Template
	raw      string
	startTag string
	endTag   string
	Format   TemplateFormat
}

// NewTemplate parses a template string and creates a new TextTemplate.
//
// templateString is the template content to be parsed.
// format is optional. If provided, it sets the escaping format.
// It returns a new TextTemplate or an error if the template string is invalid.
func NewTemplate(templateString, startTag, endTag string, format ...TemplateFormat) (*TextTemplate, error) {
	tpl, err := fasttemplate.NewTemplate(templateString, startTag, endTag)
	if err != nil {
		return nil, err
	}

	resolvedFormat := TemplateFormatUnspecified
	if len(format) > 0 {
		resolvedFormat = format[0]
	}

	if resolvedFormat == TemplateFormatUnspecified {
		trimmed := strings.TrimSpace(templateString)
		// Heuristic detection of JSON:
		// 1. Must start with { and end with } (Object) OR start with [ and end with ] (Array)
		// 2. Must NOT start with the startTag (to avoid misidentifying "{{ var }}" as JSON)
		isObject := strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")
		isArray := strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")

		if (isObject || isArray) && !strings.HasPrefix(trimmed, startTag) {
			resolvedFormat = TemplateFormatJSON
		}
	}

	return &TextTemplate{
		template: tpl,
		raw:      templateString,
		startTag: startTag,
		endTag:   endTag,
		Format:   resolvedFormat,
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

		switch t.Format {
		case TemplateFormatJSON, TemplateFormatYAML:
			// JSON and YAML (double-quoted) share similar escaping requirements for strings.
			// We assume the user has quoted the placeholder in the template: "{{val}}".
			if s, ok := val.(string); ok {
				return io.WriteString(w, escapeJSONString(s))
			}
			b, err := json.Marshal(val)
			if err != nil {
				return fmt.Fprintf(w, "%v", val)
			}
			return w.Write(b)

		case TemplateFormatXML:
			// XML escaping.
			var s string
			if strVal, ok := val.(string); ok {
				s = strVal
			} else {
				// Convert non-string to string representation before escaping
				// For objects/arrays, this might not be ideal (it prints Go struct repr),
				// but XML serialization of arbitrary maps is complex.
				// We fallback to fmt.Sprint.
				s = fmt.Sprintf("%v", val)
			}

			cw := &countingWriter{w: w}
			if err := xml.EscapeText(cw, []byte(s)); err != nil {
				return 0, err
			}
			return cw.n, nil

		default:
			// No escaping (Text)
			if s, ok := val.(string); ok {
				return io.WriteString(w, s)
			}
			return fmt.Fprintf(w, "%v", val)
		}
	})
}

func escapeJSONString(s string) string {
	b, _ := json.Marshal(s)
	if len(b) >= 2 {
		return string(b[1 : len(b)-1])
	}
	return s
}

type countingWriter struct {
	w io.Writer
	n int
}

func (cw *countingWriter) Write(p []byte) (int, error) {
	n, err := cw.w.Write(p)
	cw.n += n
	return n, err
}
