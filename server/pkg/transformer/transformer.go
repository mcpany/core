// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"text/template"
)

// Transformer provides functionality to transform a map of data into a
// structured string using a Go template. It supports multiple output formats
// specified by the template, such as JSON, XML, or plain text.
type Transformer struct {
	cache sync.Map
	pool  sync.Pool
}

// NewTransformer creates and returns a new instance of Transformer.
func NewTransformer() *Transformer {
	return &Transformer{
		pool: sync.Pool{
			New: func() any {
				return new(bytes.Buffer)
			},
		},
	}
}

// Transform takes a map of data and a Go template string and returns a byte
// slice containing the transformed output.
//
// templateStr is the Go template to be executed.
// data is the data to be used in the template.
// It returns the transformed data as a byte slice or an error if the
// transformation fails.
func (t *Transformer) Transform(templateStr string, data any) ([]byte, error) {
	var tmpl *template.Template
	var err error

	// Check cache to avoid expensive template parsing
	if v, ok := t.cache.Load(templateStr); ok {
		tmpl = v.(*template.Template)
	} else {
		tmpl, err = template.New("transformer").Funcs(template.FuncMap{
			"json": func(v any) (string, error) {
				b, err := json.Marshal(v)
				if err != nil {
					return "", err
				}
				return string(b), nil
			},
			"join": func(sep string, a []any) string {
				var sb strings.Builder
				// Heuristic: estimate 5 chars per item plus separator
				if n := len(a); n > 0 {
					sb.Grow(n * (5 + len(sep)))
				}
				for i, v := range a {
					if i > 0 {
						sb.WriteString(sep)
					}
					switch val := v.(type) {
					case string:
						sb.WriteString(val)
					case int:
						sb.WriteString(strconv.Itoa(val))
					case int64:
						sb.WriteString(strconv.FormatInt(val, 10))
					case float64:
						sb.WriteString(strconv.FormatFloat(val, 'g', -1, 64))
					default:
						fmt.Fprint(&sb, v)
					}
				}
				return sb.String()
			},
		}).Parse(templateStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse template: %w", err)
		}
		t.cache.Store(templateStr, tmpl)
	}

	// Use pool to reduce allocations
	bufPtr := t.pool.Get()
	var buf *bytes.Buffer
	if bufPtr == nil {
		buf = new(bytes.Buffer)
	} else {
		buf = bufPtr.(*bytes.Buffer)
	}
	defer func() {
		buf.Reset()
		t.pool.Put(buf)
	}()

	if err := tmpl.Execute(buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	// We must copy the bytes because the buffer is returned to the pool
	// and subsequent uses would overwrite the data.
	out := make([]byte, buf.Len())
	copy(out, buf.Bytes())

	return out, nil
}
