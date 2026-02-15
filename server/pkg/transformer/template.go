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
	IsJSON   bool
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
// params is a map of key-value pairs that will be available within the
// template.
// It returns the rendered string or an error if the template execution fails.
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
