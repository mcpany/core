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
				sepLen := len(sep)
				var totalLen int

				// First pass: try to calculate total length
				// Optimized for strings, but handles common types
				for i, v := range a {
					if s, ok := v.(string); ok {
						totalLen += len(s)
						if i > 0 {
							totalLen += sepLen
						}
					} else {
						// Not a string, try to estimate length for other types
						if i > 0 {
							totalLen += sepLen
						}

						// Handle current item
						l := estimateLen(v)
						if l == -1 {
							// fallback
							totalLen += (len(a) - i) * 10
							break
						}
						totalLen += l

						// Continue loop for the rest
						stop := false
						for j := i + 1; j < len(a); j++ {
							if j > 0 {
								totalLen += sepLen
							}
							l := estimateLen(a[j])
							if l == -1 {
								totalLen += (len(a) - j) * 10
								stop = true
								break
							}
							totalLen += l
						}
						if stop {
							break
						}
						// We processed all remaining items in the inner loop
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
					case fmt.Stringer:
						sb.WriteString(val.String())
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

func estimateLen(v any) int {
	switch val := v.(type) {
	case string:
		return len(val)
	case int:
		return intLen(int64(val))
	case int64:
		return intLen(val)
	case bool:
		if val {
			return 4
		}
		return 5
	default:
		return -1
	}
}

func intLen(i int64) int {
	if i >= 0 {
		return uintLen(uint64(i))
	}
	// For MinInt64 (-9223372036854775808), -i overflows and wraps back to MinInt64.
	// uint64(MinInt64) is 1<<63 (9223372036854775808).
	// uintLen handles this correctly (19 digits).
	// Result is 1 + 19 = 20. Correct.
	return 1 + uintLen(uint64(-i))
}

func uintLen(u uint64) int {
	if u < 10 {
		return 1
	}
	if u < 100 {
		return 2
	}
	if u < 1000 {
		return 3
	}
	if u < 10000 {
		return 4
	}

	// For larger numbers, use a loop
	cnt := 0
	for u >= 10000 {
		u /= 10000
		cnt += 4
	}
	// u is now < 10000
	if u < 10 {
		return cnt + 1
	}
	if u < 100 {
		return cnt + 2
	}
	if u < 1000 {
		return cnt + 3
	}
	return cnt + 4
}
