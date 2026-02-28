// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"encoding/json"
	"testing"
)

func TestEstimateJSONSize(t *testing.T) {
	tests := []struct {
		name string
		v    interface{}
	}{
		{"nil", nil},
		{"string", "hello"},
		{"string_empty", ""},
		{"int", 123},
		{"int8", int8(12)},
		{"int16", int16(-1234)},
		{"int32", int32(1234567)},
		{"int64", int64(-1234567890)},
		{"uint", uint(123)},
		{"uint8", uint8(12)},
		{"uint16", uint16(1234)},
		{"uint32", uint32(1234567)},
		{"uint64", uint64(1234567890)},
		{"uint64_zero", uint64(0)},
		{"uint64_small", uint64(5)},
		{"uint64_medium", uint64(50)},
		{"bool_true", true},
		{"bool_false", false},
		{"float32", float32(123.45)},
		{"float64", float64(123.456)},
		{"byte_slice", []byte("hello world")},
		{"byte_slice_empty", []byte("")},
		{"slice_int", []int{1, 2, 3}},
		{"slice_interface", []interface{}{1, "two", 3.0, true}},
		{"slice_string", []string{"a", "b", "c"}},
		{"slice_empty", []string{}},
		{"slice_nil", []string(nil)},
		{"array", [3]int{1, 2, 3}},
		{"map", map[string]int{"a": 1, "b": 2}},
		{"map_empty", map[string]int{}},
		{"map_nil", map[string]int(nil)},
		{"map_string_string", map[string]string{"k1": "v1", "k2": "v2"}},
		{"map_interface", map[string]interface{}{"k1": 1, "k2": "v2"}},
		{"map_int_key", map[int]string{1: "a", 2: "b"}},
		{"struct", struct {
			A int    `json:"a"`
			B string `json:"b,omitempty"`
		}{1, "test"}},
		{"struct_omitempty_empty", struct {
			A int    `json:"a"`
			B string `json:"b,omitempty"`
		}{1, ""}},
		{"struct_omitempty_zero", struct {
			A int `json:"a"`
			B int `json:"b,omitempty"`
		}{1, 0}},
		{"struct_omitempty_float_zero", struct {
			A float64 `json:"a,omitempty"`
		}{0.0}},
		{"struct_omitempty_bool_false", struct {
			A bool `json:"a,omitempty"`
		}{false}},
		{"struct_omitempty_uint_zero", struct {
			A uint `json:"a,omitempty"`
		}{0}},
		{"struct_omitempty_ptr_nil", struct {
			A *int `json:"a,omitempty"`
		}{nil}},
		{"struct_omitempty_map_empty", struct {
			A map[string]int `json:"a,omitempty"`
		}{map[string]int{}}},
		{"struct_omitempty_slice_empty", struct {
			A []int `json:"a,omitempty"`
		}{[]int{}}},
		{"struct_omitempty_array_empty", struct {
			A [0]int `json:"a,omitempty"`
		}{[0]int{}}},
		{"struct_omitempty_interface_nil", struct {
			A interface{} `json:"a,omitempty"`
		}{nil}},
		{"struct_omitempty_interface_empty_string", struct {
			A interface{} `json:"a,omitempty"`
		}{""}},
		{"struct_unexported", struct {
			A int `json:"a"`
			b int
		}{1, 2}},
		{"struct_no_tag", struct {
			A int
			B string
		}{1, "test"}},
		{"pointer", func() *int { i := 42; return &i }()},
		{"pointer_nil", (*int)(nil)},
		{"interface_nil", func() interface{} { var i interface{} = (*int)(nil); return i }()},
		{"nested", map[string]interface{}{
			"a": []int{1, 2},
			"b": map[string]string{"c": "d"},
		}},
		{"struct_cycle", func() interface{} {
			type Node struct {
				Next *Node
				Value int
			}
			n := &Node{Value: 1}
			n.Next = n
			return n
		}()},
		{"unsupported_type", complex(1, 2)},
		{"cycle_map", func() interface{} {
			m := make(map[string]interface{})
			m["self"] = m
			return m
		}()},
		{"cycle_ptr", func() interface{} {
			type Node struct {
				Next *Node
			}
			n := &Node{}
			n.Next = n
			return n
		}()},
		{"cycle_reflect_slice", func() interface{} {
			type Node struct {
				Items []interface{}
			}
			n := Node{}
			n.Items = append(n.Items, &n)
			return n
		}()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EstimateJSONSize(tt.v)
			b, err := json.Marshal(tt.v)
			if err != nil {
				// We still want to test cycles are not panicking or hanging.
				// In this case `b` is empty and `want` is zero.
				// For valid JSON, we compare the sizes.
				return
			}
			want := len(b)

			// Allow some margin of error for whitespace or float formatting differences
			// But for simple types it should be exact or very close
			diff := got - want
			if diff < 0 { diff = -diff }

			// We accept up to 10% difference or 5 bytes (for small objects)
			if diff > 5 && float64(diff)/float64(want) > 0.1 {
				t.Errorf("EstimateJSONSize() = %v, want %v (json: %s)", got, want, string(b))
			}
		})
	}
}

func BenchmarkEstimateJSONSize(b *testing.B) {
	v := map[string]interface{}{
		"id": "12345",
		"data": make([]string, 1000),
		"items": make([]map[string]int, 100),
		"nested": map[string]interface{}{
			"name": "test",
			"value": 123.456,
		},
	}
	for i := 0; i < 1000; i++ {
		v["data"].([]string)[i] = "some string data"
	}
	for i := 0; i < 100; i++ {
		v["items"].([]map[string]int)[i] = map[string]int{"a": 1, "b": 2}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EstimateJSONSize(v)
	}
}

func BenchmarkJSONMarshal(b *testing.B) {
	v := map[string]interface{}{
		"id": "12345",
		"data": make([]string, 1000),
		"items": make([]map[string]int, 100),
		"nested": map[string]interface{}{
			"name": "test",
			"value": 123.456,
		},
	}
	for i := 0; i < 1000; i++ {
		v["data"].([]string)[i] = "some string data"
	}
	for i := 0; i < 100; i++ {
		v["items"].([]map[string]int)[i] = map[string]int{"a": 1, "b": 2}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(v)
	}
}
