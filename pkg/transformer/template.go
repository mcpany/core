/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package transformer

import (
	"bytes"
	"text/template"
)

// TextTemplate provides a simple wrapper around Go's standard text/template
// for rendering strings with dynamic data.
type TextTemplate struct {
	template *template.Template
	raw      string
	startTag string
	endTag   string
}

// NewTemplate parses a template string and creates a new TextTemplate.
//
// templateString is the template content to be parsed.
// It returns a new TextTemplate or an error if the template string is invalid.
func NewTemplate(templateString, startTag, endTag string) (*TextTemplate, error) {
	tpl, err := template.New("").Delims(startTag, endTag).Parse(templateString)
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
	var buffer bytes.Buffer
	if err := t.template.Execute(&buffer, params); err != nil {
		return "", err
	}
	return buffer.String(), nil
}
