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
		{"int", 123},
		{"bool", true},
		{"float", 123.456},
		{"slice", []int{1, 2, 3}},
		{"map", map[string]int{"a": 1, "b": 2}},
		{"struct", struct {
			A int    `json:"a"`
			B string `json:"b,omitempty"`
		}{1, "test"}},
		{"nested", map[string]interface{}{
			"a": []int{1, 2},
			"b": map[string]string{"c": "d"},
		}},
		{"int8", int8(12)},
		{"int16", int16(123)},
		{"int32", int32(1234)},
		{"int64", int64(12345)},
		{"uint", uint(123)},
		{"uint8", uint8(12)},
		{"uint16", uint16(123)},
		{"uint32", uint32(1234)},
		{"uint64", uint64(12345)},
		{"uint zero", uint(0)},
		{"float32", float32(123.456)},
		{"bool false", false},
		{"empty slice", []interface{}{}},
		{"slice interface", []interface{}{1, "two", true}},
		{"empty map", map[string]interface{}{}},
		{"nil map", map[string]interface{}(nil)},
		{"nil slice", []interface{}(nil)},
		{"byte slice", []byte("test")},
		{"empty byte slice", []byte("")},
		{"string slice", []string{"a", "b", "c"}},
		{"struct with omitempty empty", struct {
			A int    `json:"a,omitempty"`
			B string `json:"b,omitempty"`
			C bool   `json:"c,omitempty"`
			D []int  `json:"d,omitempty"`
			E map[string]int `json:"e,omitempty"`
			F *int   `json:"f,omitempty"`
			G float64 `json:"g,omitempty"`
			H uint   `json:"h,omitempty"`
			I interface{} `json:"i,omitempty"`
			J struct{} `json:"-"`
			unexported int
		}{}},
		{"struct reflect map empty", struct {
			A map[string]interface{}
		}{A: map[string]interface{}{}}},
		{"struct reflect map string interface", struct {
			A map[string]interface{}
		}{A: map[string]interface{}{"a": 1}}},
		{"struct reflect array interface", struct {
			A [2]interface{}
		}{A: [2]interface{}{1, 2}}},
		{"struct reflect interface nil", struct {
			A interface{}
		}{A: nil}},
		{"struct reflect ptr value", struct {
			A *int
		}{A: func() *int { i := 5; return &i }()}},
		{"struct reflect map int key", struct {
			A map[int]string
		}{A: map[int]string{1: "a"}}},
		{"struct reflect ptr nil", struct {
			A *int
		}{A: nil}},
		{"struct reflect interface string", struct {
			A interface{}
		}{A: "hello"}},
		{"uint small", uint(5)},
		{"uint mid", uint(50)},
		{"int small", int(5)},
		{"int mid", int(50)},
		{"int big", int(500)},
		{"empty struct", struct{}{}},
		{"complex type", complex(1, 2)},
		{"float32 zero", float32(0)},
		{"float64 zero", float64(0)},
		{"reflect.Value string", func() interface{} {
			v := "test"
			return &v
		}()},
		{"isEmptyValue array empty", [0]int{}},
		{"isEmptyValue slice empty", make([]int, 0)},
		{"isEmptyValue uint8 zero", uint8(0)},
		{"isEmptyValue float32 zero obj", struct{ A float32 `json:",omitempty"` }{}},
		{"struct cycle slice ptr", func() interface{} {
			type Node struct { Next *Node; Value int }
			n1 := &Node{Value: 1}
			n2 := &Node{Value: 2, Next: n1}
			n1.Next = n2
			return n1
		}()},
		{"struct map cycle ptr", func() interface{} {
			m := make(map[string]interface{})
			m["self"] = m
			return m
		}()},
		{"slice cycle interface", func() interface{} {
			s := make([]interface{}, 1)
			s[0] = &s
			return s
		}()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EstimateJSONSize(tt.v)
			b, err := json.Marshal(tt.v)

			// if json.Marshal fails (e.g. unsupported type like cyclic structure), we just assert it didn't crash and returns some size >= 0
			if err != nil {
				if got < 0 {
					t.Errorf("EstimateJSONSize() = %v on cyclic structure, want >= 0", got)
				}
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
