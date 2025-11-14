package template

import (
	"bytes"
	"text/template"
)

// Render renders a template with the given arguments.
func Render(tmpl string, args map[string]any) (string, error) {
	t, err := template.New("").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, args); err != nil {
		return "", err
	}
	return buf.String(), nil
}
