// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
)

// TextTemplate provides a wrapper around Go's standard text/template
// for rendering strings with dynamic data.
type TextTemplate struct {
	template *template.Template
	raw      string
}

// NewTemplate parses a template string and creates a new TextTemplate.
//
// templateString is the template content to be parsed.
// startTag and endTag are currently ignored as text/template uses {{ }} by default,
// but kept for API compatibility.
// It returns a new TextTemplate or an error if the template string is invalid.
func NewTemplate(templateString, startTag, endTag string) (*TextTemplate, error) {
	// Create a new template with helper functions
	tpl := template.New("template").Option("missingkey=error").Funcs(template.FuncMap{
		"json": func(v any) (string, error) {
			b, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			return string(b), nil
		},
		"join": func(sep string, s []string) string {
			return strings.Join(s, sep)
		},
	})

	// Parse the template
	// Note: text/template uses {{ }} by default. Custom delimiters are supported via Delims(),
	// but we only support if startTag/endTag are provided and non-empty.
	if startTag != "" && endTag != "" {
		tpl = tpl.Delims(startTag, endTag)
	}

	parsedTpl, err := tpl.Parse(templateString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return &TextTemplate{
		template: parsedTpl,
		raw:      templateString,
	}, nil
}

// Render executes the template with the provided parameters and returns the
// resulting string.
//
// params is a map of key-value pairs that will be available within the
// template.
// It returns the rendered string or an error if the template execution fails.
func (t *TextTemplate) Render(params map[string]any) (string, error) {
	var buf bytes.Buffer
	if err := t.template.Execute(&buf, params); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	return buf.String(), nil
}
