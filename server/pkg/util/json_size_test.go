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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EstimateJSONSize(tt.v)
			b, _ := json.Marshal(tt.v)
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
