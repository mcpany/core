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
	// ⚡ BOLT: Optimized joinFunc to avoid allocation of []any for common slice types
	// Randomized Selection from Top 5 High-Impact Targets

	val := reflect.ValueOf(input)
	kind := val.Kind()
	if kind != reflect.Slice && kind != reflect.Array {
		return "", fmt.Errorf("expected slice or array, got %T", input)
	}

	n := val.Len()
	if n == 0 {
		return "", nil
	}

	var sb strings.Builder
	// Heuristic for capacity: assume average 10 bytes per item + separator
	// This avoids expensive pre-calculation via reflection
	sb.Grow(n * (10 + len(sep)))

	// Fast path for []string to avoid reflection overhead per item
	if s, ok := input.([]string); ok {
		for i, v := range s {
			if i > 0 {
				sb.WriteString(sep)
			}
			sb.WriteString(v)
		}
		return sb.String(), nil
	}

	// Scratch buffer for number conversion to avoid allocations
	var scratch [64]byte

	for i := 0; i < n; i++ {
		if i > 0 {
			sb.WriteString(sep)
		}

		v := val.Index(i)
		// Handle interface elements (e.g. []any)
		if v.Kind() == reflect.Interface {
			if v.IsNil() {
				sb.WriteString("<nil>")
				continue
			}
			v = v.Elem()
		}

		// Check if the type is a standard Go type (empty PkgPath).
		// If it's a named type (e.g. type MyInt int), it might implement fmt.Stringer,
		// so we should fall back to fmt package logic to be safe.
		if v.Type().PkgPath() == "" {
			switch v.Kind() {
			case reflect.String:
				sb.WriteString(v.String())
				continue
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				sb.Write(strconv.AppendInt(scratch[:0], v.Int(), 10))
				continue
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				sb.Write(strconv.AppendUint(scratch[:0], v.Uint(), 10))
				continue
			case reflect.Float32:
				sb.Write(strconv.AppendFloat(scratch[:0], v.Float(), 'g', -1, 32))
				continue
			case reflect.Float64:
				sb.Write(strconv.AppendFloat(scratch[:0], v.Float(), 'g', -1, 64))
				continue
			case reflect.Bool:
				sb.Write(strconv.AppendBool(scratch[:0], v.Bool()))
				continue
			}
		}

		// Fallback for named types or complex types (structs, maps, etc.)
		// Also handles fmt.Stringer implementation correctly.
		if v.IsValid() && v.CanInterface() {
			if stringer, ok := v.Interface().(fmt.Stringer); ok {
				sb.WriteString(stringer.String())
				continue
			}
			fmt.Fprint(&sb, v.Interface())
		} else {
			fmt.Fprint(&sb, v)
		}
	}
	return sb.String(), nil
}
