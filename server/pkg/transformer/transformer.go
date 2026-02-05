// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
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
//
// Returns the result.
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
			"join": joinFunc,
		}).Parse(templateStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse template: %w", err)
		}
		t.cache.Store(templateStr, tmpl)
	}

	// Use pool to reduce allocations
	// Safe to type assert because New is defined in NewTransformer
	buf := t.pool.Get().(*bytes.Buffer)
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

func joinFunc(sep string, input any) (string, error) {
	a, err := toAnySlice(input)
	if err != nil {
		return "", fmt.Errorf("join: %w", err)
	}

	var sb strings.Builder
	sepLen := len(sep)
	var totalLen int

	// First pass: try to calculate total length
	for i, v := range a {
		if i > 0 {
			totalLen += sepLen
		}

		if s, ok := v.(string); ok {
			totalLen += len(s)
		} else {
			// For non-strings, we stop estimating and just guess.
			// This avoids expensive iteration and type assertions for numbers.
			remaining := len(a) - i
			totalLen += remaining * 10
			// Add separators estimate for remaining items
			if remaining > 1 {
				totalLen += (remaining - 1) * sepLen
			}
			break
		}
	}

	sb.Grow(totalLen)

	// Scratch buffer for number conversion to avoid allocations
	var scratch [64]byte

	for i, v := range a {
		if i > 0 {
			sb.WriteString(sep)
		}
		switch val := v.(type) {
		case string:
			sb.WriteString(val)
		case int:
			sb.Write(strconv.AppendInt(scratch[:0], int64(val), 10))
		case int64:
			sb.Write(strconv.AppendInt(scratch[:0], val, 10))
		case int32:
			sb.Write(strconv.AppendInt(scratch[:0], int64(val), 10))
		case int16:
			sb.Write(strconv.AppendInt(scratch[:0], int64(val), 10))
		case int8:
			sb.Write(strconv.AppendInt(scratch[:0], int64(val), 10))
		case uint:
			sb.Write(strconv.AppendUint(scratch[:0], uint64(val), 10))
		case uint64:
			sb.Write(strconv.AppendUint(scratch[:0], val, 10))
		case uint32:
			sb.Write(strconv.AppendUint(scratch[:0], uint64(val), 10))
		case uint16:
			sb.Write(strconv.AppendUint(scratch[:0], uint64(val), 10))
		case uint8:
			sb.Write(strconv.AppendUint(scratch[:0], uint64(val), 10))
		case float64:
			// Use -1 to behave like %v / %g
			sb.Write(strconv.AppendFloat(scratch[:0], val, 'g', -1, 64))
		case float32:
			sb.Write(strconv.AppendFloat(scratch[:0], float64(val), 'g', -1, 32))
		case bool:
			sb.Write(strconv.AppendBool(scratch[:0], val))
		case fmt.Stringer:
			sb.WriteString(val.String())
		default:
			fmt.Fprint(&sb, v)
		}
	}
	return sb.String(), nil
}

func toAnySlice(input any) ([]any, error) {
	if s, ok := input.([]any); ok {
		return s, nil
	}
	val := reflect.ValueOf(input)
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return nil, fmt.Errorf("expected slice or array, got %T", input)
	}
	l := val.Len()
	a := make([]any, l)
	for i := 0; i < l; i++ {
		a[i] = val.Index(i).Interface()
	}
	return a, nil
}
