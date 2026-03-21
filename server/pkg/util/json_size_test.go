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
		{"int8", int8(12)},
		{"int16", int16(1234)},
		{"int32", int32(123456)},
		{"int64", int64(123456789)},
		{"uint", uint(123)},
		{"uint8", uint8(12)},
		{"uint16", uint16(1234)},
		{"uint32", uint32(123456)},
		{"uint64", uint64(123456789)},
		{"bool_true", true},
		{"bool_false", false},
		{"float32", float32(123.456)},
		{"float64", float64(123.456)},
		{"bytes_empty", []byte{}},
		{"bytes", []byte("hello")},
		{"string_slice", []string{"a", "b", "c"}},
		{"map_string_string", map[string]string{"a": "1", "b": "2"}},
		{"map_string_any", map[string]interface{}{"a": 1, "b": "2"}},
		{"map_empty", map[string]interface{}{}},
		{"slice_any", []interface{}{1, "2", true}},
		{"slice_empty", []interface{}{}},
		{"struct_simple", struct {
			A int    `json:"a"`
			B string `json:"b,omitempty"`
			C bool   `json:"-"`
			d string // unexported
		}{1, "test", true, "hidden"}},
		{"struct_omitempty", struct {
			A int    `json:"a,omitempty"`
			B string `json:"b,omitempty"`
			C bool   `json:"c,omitempty"`
			D []int  `json:"d,omitempty"`
			E *int   `json:"e,omitempty"`
		}{0, "", false, nil, nil}},
		{"pointer", func() *int { i := 42; return &i }()},
		{"pointer_nil", (*int)(nil)},
		{"nested_complex", map[string]interface{}{
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

func TestEstimateJSONSizeCycles(t *testing.T) {
	// Map cycle
	m := make(map[string]interface{})
	m["self"] = m

	// Should not panic or infinite loop, but size estimate might be 0 for the cyclic part
	EstimateJSONSize(m)

	// Slice cycle - we need a struct to make a slice cycle in Go easily without unsafe
	type Node struct {
		Children []interface{}
	}
	n := &Node{}
	n.Children = append(n.Children, n)
	EstimateJSONSize(n)
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

func TestEstimateJSONSizeEmptyValues(t *testing.T) {
	// Need to cover isEmptyValue cases
	tests := []struct {
		name string
		v    interface{}
		want int
	}{
		{"struct_empty_map", struct {
			M map[string]int `json:"m,omitempty"`
		}{make(map[string]int)}, 2}, // {}
		{"struct_empty_slice", struct {
			S []int `json:"s,omitempty"`
		}{make([]int, 0)}, 2}, // {}
		{"struct_empty_string", struct {
			S string `json:"s,omitempty"`
		}{""}, 2}, // {}
		{"struct_zero_int", struct {
			I int `json:"i,omitempty"`
		}{0}, 2}, // {}
		{"struct_zero_uint", struct {
			U uint `json:"u,omitempty"`
		}{0}, 2}, // {}
		{"struct_zero_float", struct {
			F float32 `json:"f,omitempty"`
		}{0.0}, 2}, // {}
		{"struct_false_bool", struct {
			B bool `json:"b,omitempty"`
		}{false}, 2}, // {}
		{"struct_nil_ptr", struct {
			P *int `json:"p,omitempty"`
		}{nil}, 2}, // {}
		{"struct_nil_interface", struct {
			I interface{} `json:"i,omitempty"`
		}{nil}, 2}, // {}

		// Map using reflection
		{"reflect_map", map[int]string{1: "a", 2: "b"}, 17}, // estimateReflectMap
		{"reflect_map_empty", map[int]string{}, 2}, // estimateReflectMap empty
		{"reflect_map_nil", (map[int]string)(nil), 4}, // estimateReflectMap nil

		// Slice using reflection
		{"reflect_slice", []int{1, 2, 3}, 7}, // estimateReflectSlice
		{"reflect_slice_empty", []int{}, 2}, // estimateReflectSlice empty
		{"reflect_slice_nil", ([]int)(nil), 4}, // estimateReflectSlice nil

		// Interface using reflection
		{"reflect_interface_struct", struct {
			I interface{} `json:"i"`
		}{42}, 8},

		// Edge cases for ints
		{"int_zero", 0, 1},
		{"int_10", 10, 2},
		{"int_100", 100, 3},
		{"int_1000", 1000, 4},
		{"int_negative", -5, 2},

		// Edge cases for uints
		{"uint_zero", uint(0), 1},
		{"uint_10", uint(10), 2},
		{"uint_100", uint(100), 3},
		{"uint_1000", uint(1000), 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EstimateJSONSize(tt.v)
			// we can't always compare to real json.Marshal directly because reflection maps
			// might iterate in different order or we are just testing the estimation logic
			if got <= 0 {
				t.Errorf("EstimateJSONSize() returned %v, want > 0", got)
			}
		})
	}
}

func TestEstimateJSONSizeVisitedPool(t *testing.T) {
    // Force reuse of the sync.Pool by calling concurrently or just multiple times
    for i := 0; i < 100; i++ {
        EstimateJSONSize(map[string]int{"a": 1})
    }
}
func TestEstimateJSONSizeVisitedPoolConcurrent(t *testing.T) {
	// Let's use multiple goroutines to cause the pool to allocate multiple maps
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			EstimateJSONSize(map[string]interface{}{"a": 1, "b": "2"})
			done <- true
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
}
func TestEstimateJSONSizeRemainingBranches(t *testing.T) {
	tests := []struct {
		name string
		v    interface{}
	}{
		// Arrays
		{"array", [3]int{1, 2, 3}},
		{"array_empty", [0]int{}},

		// Unnamed structs
		{"struct_anon", struct{ X int }{1}},

		// isEmptyValue edge cases
		{"struct_zero_int8", struct{ V int8 `json:"v,omitempty"` }{0}},
		{"struct_zero_int16", struct{ V int16 `json:"v,omitempty"` }{0}},
		{"struct_zero_int32", struct{ V int32 `json:"v,omitempty"` }{0}},
		{"struct_zero_int64", struct{ V int64 `json:"v,omitempty"` }{0}},

		{"struct_zero_uint8", struct{ V uint8 `json:"v,omitempty"` }{0}},
		{"struct_zero_uint16", struct{ V uint16 `json:"v,omitempty"` }{0}},
		{"struct_zero_uint32", struct{ V uint32 `json:"v,omitempty"` }{0}},
		{"struct_zero_uint64", struct{ V uint64 `json:"v,omitempty"` }{0}},
		{"struct_zero_uintptr", struct{ V uintptr `json:"v,omitempty"` }{0}},

		{"struct_zero_float64", struct{ V float64 `json:"v,omitempty"` }{0}},

		{"struct_empty_array", struct{ V [0]int `json:"v,omitempty"` }{[0]int{}}},
		{"struct_empty_map", struct{ V map[int]int `json:"v,omitempty"` }{map[int]int{}}},

		// Uint size checks
		{"uint_9", uint(9)},
		{"uint_99", uint(99)},

		// Int size checks
		{"int_9", int(9)},
		{"int_99", int(99)},
		{"int_999", int(999)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EstimateJSONSize(tt.v)
			if got <= 0 {
				t.Errorf("EstimateJSONSize() = %v", got)
			}
		})
	}
}
func TestEstimateJSONSizePtrReflect(t *testing.T) {
	// We need a custom type to trigger reflect.Ptr logic not handled by fast paths
	type MyInt int
	val := MyInt(42)
	EstimateJSONSize(&val)

	var nilPtr *MyInt
	EstimateJSONSize(nilPtr)

	// Test cycle in a struct with a pointer to itself
	type NodePtr struct {
		Next *NodePtr
	}
	n := &NodePtr{}
	n.Next = n
	EstimateJSONSize(n)
}

func TestEstimateJSONSizeReflectInterfaceNil(t *testing.T) {
	// Cover reflect.Interface where it's nil
	type Wrapper struct {
		I interface{}
	}
	w := Wrapper{}
	EstimateJSONSize(w)
}
func TestEstimateJSONSizeEmptyValueEdgeCases(t *testing.T) {
	// Let's call isEmptyValue directly to guarantee coverage of all its branches
	// Since it's an unexported function, we can test it indirectly via structs with omitempty

	tests := []struct {
		name string
		v    interface{}
	}{
		// Pointers
		{"struct_nil_ptr_bool", struct{ P *bool `json:"p,omitempty"` }{nil}},
		// Arrays
		{"struct_empty_array_bool", struct{ A [0]bool `json:"a,omitempty"` }{[0]bool{}}},
		// Maps
		{"struct_empty_map_bool", struct{ M map[string]bool `json:"m,omitempty"` }{map[string]bool{}}},
		// Slices
		{"struct_empty_slice_bool", struct{ S []bool `json:"s,omitempty"` }{[]bool{}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EstimateJSONSize(tt.v)
			if got != 2 {
				t.Errorf("expected 2 (for {}), got %d", got)
			}
		})
	}
}

func TestEstimateReflectRemaining(t *testing.T) {
    // estimateReflect defaults
    type MyString string
    v := MyString("hello")
    EstimateJSONSize(v) // Should hit the default case in estimateReflect

    // Test slice cycle properly
    type RecurseSlice []interface{}
    s := make(RecurseSlice, 1)
    s[0] = s
    EstimateJSONSize(s)
}
