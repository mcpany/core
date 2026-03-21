// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tokenizer

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestSimpleTokenizer(t *testing.T) {
	tokenizer := NewSimpleTokenizer()

	tests := []struct {
		input string
		want  int
	}{
		{"", 0},
		{"a", 1},
		{"abcd", 1},
		{"abcdefgh", 2},
		{"hello world", 2}, // 11 chars / 4 = 2.75 -> 2
	}

	for _, tt := range tests {
		got, _ := tokenizer.CountTokens(tt.input)
		if got != tt.want {
			t.Errorf("CountTokens(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestWordTokenizer(t *testing.T) {
	tokenizer := NewWordTokenizer()

	tests := []struct {
		input string
		want  int // approximate
	}{
		{"", 0},
		{"hello", 1},
		{"hello world", 2},             // 2 * 1.3 = 2.6 -> 2
		{"this is a test sentence", 6}, // 5 * 1.3 = 6.5 -> 6
		{"hello 🌍", 2},                 // 2 * 1.3 = 2.6 -> 2 (ASCII + Emoji)
		{"你好 世界", 2},                   // 2 * 1.3 = 2.6 -> 2 (Chinese + Space + Chinese)
		{"hello\tworld\n", 2},          // ASCII whitespace
	}

	for _, tt := range tests {
		got, _ := tokenizer.CountTokens(tt.input)
		if got != tt.want {
			t.Errorf("CountTokens(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestCountTokensInValue(t *testing.T) {
	tokenizer := NewSimpleTokenizer()

	val := map[string]interface{}{
		"key": "abcdefgh", // key "key" (1) + val "abcdefgh" (2) = 3
		"list": []interface{}{ // key "list" (1) + val list...
			"abcd", // 1
			"1234", // 1
		},
	}
	// Total: 3 (key:value) + 1 (list key) + 1 (abcd) + 1 (1234) = 6

	got, _ := CountTokensInValue(tokenizer, val)
	if got != 6 {
		t.Errorf("CountTokensInValue = %d, want 6", got)
	}

	// Test specific types for SimpleTokenizer (optimization paths)
	stringSlice := []string{"abcd", "efgh"} // 1 + 1 = 2
	gotSlice, _ := CountTokensInValue(tokenizer, stringSlice)
	if gotSlice != 2 {
		t.Errorf("CountTokensInValue([]string) = %d, want 2", gotSlice)
	}

	stringMap := map[string]string{"key": "val"} // "key"(1) + "val"(1) = 2
	gotMap, _ := CountTokensInValue(tokenizer, stringMap)
	if gotMap != 2 {
		t.Errorf("CountTokensInValue(map[string]string) = %d, want 2", gotMap)
	}
}

func TestCountTokensInValue_Word(t *testing.T) {
	tokenizer := NewWordTokenizer()

	tests := []struct {
		name     string
		input    interface{}
		expected int // int(1.3) = 1 for primitives
	}{
		{"int", 12345, 1},
		{"bool", true, 1},
		{"nil", nil, 1},
		{"string", "hello world", 2},                        // "hello world" -> 2 words * 1.3 -> 2
		{"slice", []interface{}{1, "hello"}, 1 + 1},         // 1 (int) + 1 (hello)
		{"map", map[string]interface{}{"a": 1}, 1 + 1},      // "a" (1) + 1 (int)
		{"string_slice", []string{"hello", "world"}, 1 + 1}, // 1 + 1
		{"string_map", map[string]string{"a": "b"}, 1 + 1},  // "a"(1) + "b"(1)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CountTokensInValue(tokenizer, tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("CountTokensInValue(%v) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func BenchmarkWordTokenizer(b *testing.B) {
	t := NewWordTokenizer()
	text := strings.Repeat("This is a sample sentence to test tokenization. ", 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = t.CountTokens(text)
	}
}

// ----------------------------------------------------------------------------
// NEW TESTS FOR COVERAGE
// ----------------------------------------------------------------------------

type ExportedStruct struct {
	Name string
	Age  int
}

type UnexportedStruct struct {
	name string
	Age  int
}

type StringerImpl struct {
	msg string
}

func (s StringerImpl) String() string {
	return s.msg
}

type RecursiveNode struct {
	Next *RecursiveNode
}

func TestCountTokensInValue_Coverage(t *testing.T) {
	tokenizer := NewSimpleTokenizer() // 4 chars per token

	t.Run("Struct Exported", func(t *testing.T) {
		s := ExportedStruct{Name: "abcd", Age: 1234}
		// "abcd" -> 1 token. "1234" -> 1 token. Total 2.
		got, err := CountTokensInValue(tokenizer, s)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if got != 2 {
			t.Errorf("Struct Exported: got %d, want 2", got)
		}
	})

	t.Run("Struct Unexported", func(t *testing.T) {
		s := UnexportedStruct{name: "abcd", Age: 1234}
		// "abcd" unexported -> ignored. "1234" exported -> 1 token. Total 1.
		got, err := CountTokensInValue(tokenizer, s)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if got != 1 {
			t.Errorf("Struct Unexported: got %d, want 1", got)
		}
	})

	t.Run("Slice of Strings", func(t *testing.T) {
		s := []string{"abcd", "efgh"}
		// 1 + 1 = 2
		got, err := CountTokensInValue(tokenizer, s)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if got != 2 {
			t.Errorf("Slice of Strings: got %d, want 2", got)
		}
	})

	t.Run("Array", func(t *testing.T) {
		arr := [2]int{1234, 5678}
		// 1 + 1 = 2
		got, err := CountTokensInValue(tokenizer, arr)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if got != 2 {
			t.Errorf("Array: got %d, want 2", got)
		}
	})

	t.Run("Pointer", func(t *testing.T) {
		i := 1234
		p := &i
		// 1
		got, err := CountTokensInValue(tokenizer, p)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if got != 1 {
			t.Errorf("Pointer: got %d, want 1", got)
		}
	})

	t.Run("Nil Pointer", func(t *testing.T) {
		var p *int
		// "null" -> 1 token
		got, err := CountTokensInValue(tokenizer, p)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if got != 1 {
			t.Errorf("Nil Pointer: got %d, want 1", got)
		}
	})

	t.Run("Error", func(t *testing.T) {
		errVal := fmt.Errorf("error msg")
		// "error msg" (9 chars) / 4 = 2.25 -> 2
		got, err := CountTokensInValue(tokenizer, errVal)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if got != 2 {
			t.Errorf("Error: got %d, want 2", got)
		}
	})

	t.Run("Stringer", func(t *testing.T) {
		s := StringerImpl{msg: "abcd"}
		// "abcd" -> 1
		got, err := CountTokensInValue(tokenizer, s)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if got != 1 {
			t.Errorf("Stringer: got %d, want 1", got)
		}
	})

	t.Run("Cycle Detection", func(t *testing.T) {
		node := &RecursiveNode{}
		node.Next = node
		_, err := CountTokensInValue(tokenizer, node)
		if err == nil {
			t.Error("Expected error due to cycle, got nil")
		} else if !strings.Contains(err.Error(), "cycle detected") {
			t.Errorf("Expected 'cycle detected' error, got: %v", err)
		}
	})

	t.Run("DAG Shared Reference", func(t *testing.T) {
		// A -> B
		// A -> C
		// B -> D
		// C -> D
		// Should count D twice (expanded).
		d := &ExportedStruct{Name: "D", Age: 1}         // "D"(1) + "1"(1) = 2 tokens
		b := &struct{ Child *ExportedStruct }{Child: d} // 2 tokens
		c := &struct{ Child *ExportedStruct }{Child: d} // 2 tokens
		a := &struct{ Left, Right interface{} }{Left: b, Right: c}
		// Left: 2. Right: 2. Total 4.
		got, err := CountTokensInValue(tokenizer, a)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if got != 4 {
			t.Errorf("DAG: got %d, want 4", got)
		}
	})
}

func TestWordTokenizer_Struct(t *testing.T) {
	tokenizer := NewWordTokenizer()
	// Test falling back to reflect (struct)
	s := ExportedStruct{Name: "hello", Age: 123}
	// "hello" -> 1 token (Word)
	// 123 -> 1 token (Word primitive)
	// Total 2.
	got, err := CountTokensInValue(tokenizer, s)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if got != 2 {
		t.Errorf("WordTokenizer Struct: got %d, want 2", got)
	}
}

func TestWordTokenizer_Branches(t *testing.T) {
	tokenizer := NewWordTokenizer()

	tests := []struct {
		input string
		want  int
	}{
		{"  hello  ", 1}, // Leading/trailing whitespace
		{"a\tb", 2},      // Tab
		{"a\r\nb", 2},    // CR LF
		{"a \x00 b", 3},  // Control char \x00
		{"a\u00A0b", 2},  // NBSP (non-ASCII space)
	}

	for _, tt := range tests {
		got, _ := tokenizer.CountTokens(tt.input)
		if got != tt.want {
			t.Errorf("CountTokens(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestErrorPropagation(t *testing.T) {
	node := &RecursiveNode{}
	node.Next = node

	// Using SimpleTokenizer
	t.Run("SimpleTokenizer", func(t *testing.T) {
		tokenizer := NewSimpleTokenizer()

		// Map with cycle
		m := map[string]interface{}{"key": node}
		if _, err := CountTokensInValue(tokenizer, m); err == nil {
			t.Error("Expected error from map with cycle")
		}

		// Slice with cycle
		s := []interface{}{node}
		if _, err := CountTokensInValue(tokenizer, s); err == nil {
			t.Error("Expected error from slice with cycle")
		}

		// Struct with cycle (field)
		type S struct{ Field interface{} }
		st := S{Field: node}
		if _, err := CountTokensInValue(tokenizer, st); err == nil {
			t.Error("Expected error from struct with cycle")
		}
	})

	// Using WordTokenizer
	t.Run("WordTokenizer", func(t *testing.T) {
		tokenizer := NewWordTokenizer()

		// Map with cycle
		m := map[string]interface{}{"key": node}
		if _, err := CountTokensInValue(tokenizer, m); err == nil {
			t.Error("Expected error from map with cycle")
		}

		// Slice with cycle
		s := []interface{}{node}
		if _, err := CountTokensInValue(tokenizer, s); err == nil {
			t.Error("Expected error from slice with cycle")
		}

		// Struct with cycle (field)
		type S struct{ Field interface{} }
		st := S{Field: node}
		if _, err := CountTokensInValue(tokenizer, st); err == nil {
			t.Error("Expected error from struct with cycle")
		}
	})

	// Error from map key (unlikely as keys are usually simple, but if key is string...)
	t.Run("Map Key Cycle", func(t *testing.T) {
		tokenizer := NewSimpleTokenizer()
		m := map[*RecursiveNode]string{node: "val"}
		if _, err := CountTokensInValue(tokenizer, m); err == nil {
			t.Error("Expected error from map key cycle")
		}
	})

	// Reflect-based cases for Slice and Map
	t.Run("Reflect Slice Cycle", func(t *testing.T) {
		tokenizer := NewSimpleTokenizer()
		s := []*RecursiveNode{node}
		if _, err := CountTokensInValue(tokenizer, s); err == nil {
			t.Error("Expected error from reflect slice cycle")
		}
	})

	t.Run("Reflect Map Cycle", func(t *testing.T) {
		tokenizer := NewSimpleTokenizer()
		m := map[int]*RecursiveNode{1: node}
		if _, err := CountTokensInValue(tokenizer, m); err == nil {
			t.Error("Expected error from reflect map cycle")
		}
	})
}

func TestFloatConsistency(t *testing.T) {
	tokenizer := NewSimpleTokenizer()

	// These numbers are integers but represented as floats.
	// We expect the token count to match the standard JSON string representation,
	// which avoids scientific notation for these ranges (unlike strconv %v).
	tests := []struct {
		val  float64
		want int
	}{
		{1234567.0, 1},   // "1234567" -> 7 chars -> 1.75 -> 1 token (Changed from 3)
		{9999999.0, 1},   // "9999999" -> 7 chars -> 1.75 -> 1 token (Changed from 3)
		{10000000.0, 2},  // "10000000" -> 8 chars -> 2 tokens (Changed from 1: "1e+07" was 5 chars)
		{123456789.0, 2}, // "123456789" -> 9 chars -> 2.25 -> 2 tokens (Changed from 3)
	}

	for _, tt := range tests {
		got, err := CountTokensInValue(tokenizer, tt.val)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if got != tt.want {
			t.Errorf("CountTokensInValue(%f) = %d, want %d", tt.val, got, tt.want)
		}
	}
}

func TestCountTokensInValueSimpleFast(t *testing.T) {
	st := NewSimpleTokenizer()

	tests := []struct {
		name      string
		input     interface{}
		wantCount int
		wantBool  bool
	}{
		{"string", "hello world", 2, true},
		{"int", int(42), 1, true},
		{"int64", int64(1234567890), 2, true},
		{"bool_true", true, 1, true},
		{"bool_false", false, 1, true},
		{"nil", nil, 1, true},
		{"float64_int", float64(42), 1, true},
		{"float64_frac", float64(3.14159), 1, true},
		{"slice_string", []string{"hello", "world"}, 2, true},
		{"slice_int", []int{1, 2, 3}, 3, true},
		{"slice_int64", []int64{1000, 2000}, 2, true},
		{"slice_bool", []bool{true, false, true}, 3, true},
		{"slice_float64", []float64{1.1, 2.0, 3.14159}, 3, true},
		{"map_string_string", map[string]string{"key1": "value1", "key2": "value2"}, 4, true},
		{"map_string_int", map[string]int{"k1": 1, "k2": 2}, 4, true},
		{"map_string_int64", map[string]int64{"k1": 1000, "k2": 2000}, 4, true},
		{"map_string_float64", map[string]float64{"k1": 1.1, "k2": 2.0}, 4, true},
		{"map_string_bool", map[string]bool{"k1": true, "k2": false}, 4, true},
		{"byte_empty", []byte{}, 0, true},
		{"byte_nonempty", []byte("hello world"), 2, true}, // len(11)/4=2
		{"unhandled", struct{}{}, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCount, gotBool, err := countTokensInValueSimpleFast(st, tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotBool != tt.wantBool {
				t.Errorf("want bool %v, got %v", tt.wantBool, gotBool)
			}
			if gotCount != tt.wantCount {
				t.Errorf("want count %d, got %d", tt.wantCount, gotCount)
			}
		})
	}
}

// ----------------------------------------------------------------------------
// NEW TESTS FOR COVERAGE
// ----------------------------------------------------------------------------

func TestCountWordsInValueFast(t *testing.T) {
	tests := []struct {
		name      string
		input     interface{}
		wantCount int
		wantBool  bool
	}{
		{"string", "hello world", 2, true},
		{"int", int(42), 1, true},
		{"int64", int64(42), 1, true},
		{"float64", float64(3.14), 1, true},
		{"bool", true, 1, true},
		{"nil", nil, 1, true},
		{"slice_string", []string{"hello", "world"}, 2, true},
		{"slice_int", []int{1, 2, 3}, 3, true},
		{"slice_int64", []int64{1, 2, 3}, 3, true},
		{"slice_float64", []float64{1.1, 2.2}, 2, true},
		{"slice_bool", []bool{true, false}, 2, true},
		{"map_string_string", map[string]string{"k1": "v1", "k2": "v2"}, 4, true}, // k1(1)+v1(1)+k2(1)+v2(1)
		{"map_string_int", map[string]int{"k1": 1, "k2": 2}, 4, true},
		{"map_string_int64", map[string]int64{"k1": 1, "k2": 2}, 4, true},
		{"map_string_float64", map[string]float64{"k1": 1.1, "k2": 2.2}, 4, true},
		{"map_string_bool", map[string]bool{"k1": true, "k2": false}, 4, true},
		{"byte_slice", []byte("hello world"), 2, true},
		{"unhandled", struct{}{}, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCount, gotBool := countWordsInValueFast(tt.input)
			if gotBool != tt.wantBool {
				t.Errorf("want bool %v, got %v", tt.wantBool, gotBool)
			}
			if gotCount != tt.wantCount {
				t.Errorf("want count %d, got %d", tt.wantCount, gotCount)
			}
		})
	}
}

func TestSimpleTokenizeInt64(t *testing.T) {
	tests := []struct {
		name  string
		input int64
		want  int
	}{
		{"fast_path_positive", 100, 1},
		{"fast_path_negative", -100, 1},
		{"exactly_fast_path_max", 9999999, 1},  // 7 digits
		{"1e7", 10000000, 2},                   // 8 digits -> len=8 -> 2 tokens
		{"1e8", 100000000, 2},                  // 9 digits -> 2 tokens
		{"1e9", 1000000000, 2},                 // 10 digits -> 2 tokens
		{"1e10", 10000000000, 2},               // 11 digits -> 2 tokens
		{"1e11", 100000000000, 3},              // 12 digits -> 3 tokens
		{"1e12", 1000000000000, 3},             // 13 digits -> 3 tokens
		{"1e13", 10000000000000, 3},            // 14 digits -> 3 tokens
		{"1e14", 100000000000000, 3},           // 15 digits -> 3 tokens
		{"1e15", 1000000000000000, 4},          // 16 digits -> 4 tokens
		{"1e16", 10000000000000000, 4},         // 17 digits -> 4 tokens
		{"1e17", 100000000000000000, 4},        // 18 digits -> 4 tokens
		{"1e18", 1000000000000000000, 4},       // 19 digits -> 4 tokens
		{"max_int64", 9223372036854775807, 4},  // 19 digits -> 4 tokens
		{"min_int64", -9223372036854775808, 5}, // 20 digits -> 5 tokens
		{"negative_large", -10000000, 2},       // 9 digits (incl sign) -> 2 tokens
		{"edge_1", 9, 1},
		{"edge_2", 99, 1},
		{"edge_3", 999, 1},
		{"edge_4", 9999, 1},
		{"edge_5", 99999, 1},
		{"edge_6", 999999, 1},
		{"edge_7", 9999999, 1},
		{"edge_11", 99999999999, 2},
		{"edge_12", 999999999999, 3},
		{"edge_13", 9999999999999, 3},
		{"edge_14", 99999999999999, 3},
		{"edge_15", 999999999999999, 3},
		{"edge_16", 9999999999999999, 4},
		{"edge_17", 99999999999999999, 4},
		{"edge_18", 999999999999999999, 4},
		{"edge_n_100", -99, 1},
		{"edge_n_1000", -999, 1},
		{"edge_n_10000", -9999, 1},
		{"edge_n_100000", -99999, 1},
		{"edge_n_1000000", -999999, 1},
		{"edge_n_10000000", -9999999, 2},
		{"edge_n_100000000", -99999999, 2},
		{"edge_n_1000000000", -999999999, 2},
		{"edge_n_10000000000", -9999999999, 2},
		{"edge_n_100000000000", -99999999999, 3},
		{"edge_n_1000000000000", -999999999999, 3},
		{"edge_n_10000000000000", -9999999999999, 3},
		{"edge_n_100000000000000", -99999999999999, 3},
		{"edge_n_1000000000000000", -999999999999999, 4},
		{"edge_n_10000000000000000", -9999999999999999, 4},
		{"edge_n_100000000000000000", -99999999999999999, 4},
		{"edge_n_1000000000000000000", -999999999999999999, 4},
		{"zero", 0, 1},
		{"single_positive", 5, 1},
		{"single_negative", -5, 1},
		{"exactly_fast_path_max_plus_one", 10000000, 2},   // len 8 -> /4 = 2
		{"exactly_fast_path_min_minus_one", -10000001, 2}, // len 9 -> /4 = 2
		{"max_default", 9223372036854775807, 4},
		{"negative_max_default", -9223372036854775807, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := simpleTokenizeInt64(tt.input)
			if got != tt.want {
				t.Errorf("simpleTokenizeInt64(%d) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestCountSliceInterfaceRaw(t *testing.T) {
	r := &rawWordCounter{}

	tests := []struct {
		name      string
		input     []interface{}
		wantCount int
	}{
		{"empty", []interface{}{}, 0},
		{"mixed", []interface{}{"hello world", 42, int64(42), 3.14, true, nil}, 2 + 1 + 1 + 1 + 1 + 1},
		{"nested", []interface{}{"hello", []interface{}{"world", 42}}, 1 + 1 + 1},
		{"unhandled", []interface{}{struct{}{}}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			visited := make(map[uintptr]bool)
			gotCount, err := countSliceInterfaceRaw(r, tt.input, visited)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotCount != tt.wantCount {
				t.Errorf("want count %d, got %d", tt.wantCount, gotCount)
			}
		})
	}
}

func TestCountTokensReflect(t *testing.T) {
	st := NewSimpleTokenizer()

	tests := []struct {
		name      string
		input     interface{}
		wantCount int
	}{
		{"struct", ExportedStruct{Name: "hello", Age: 42}, 2},
		{"slice", []int{1, 2, 3}, 3},
		{"map", map[string]int{"k1": 1}, 2},
		{"stringer", StringerImpl{msg: "hello world"}, 2},
		{"error", fmt.Errorf("error msg"), 2}, // len=9 -> /4 -> 2
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			visited := make(map[uintptr]bool)
			gotCount, err := countTokensReflect(st, tt.input, visited)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotCount != tt.wantCount {
				t.Errorf("want count %d, got %d", tt.wantCount, gotCount)
			}
		})
	}
}

func TestCountSliceInterfaceSimple(t *testing.T) {
	st := NewSimpleTokenizer()

	tests := []struct {
		name      string
		input     []interface{}
		wantCount int
	}{
		{"empty", []interface{}{}, 0},
		{"mixed_primitive", []interface{}{
			"hello world",   // len=11 -> 2
			float64(42.0),   // int path -> 42 -> 1
			float64(3.1415), // float path -> "3.1415" (6 chars) -> 1
			int(100),        // 1
			int64(200),      // 1
			true,            // 1
			nil,             // 1
		}, 8},
		{"nested_slice", []interface{}{
			"test",                  // 1
			[]interface{}{"nested"}, // 1
		}, 2},
		{"nested_map", []interface{}{
			map[string]interface{}{"k1": "v1"}, // k(1)+v(1)=2
		}, 2},
		{"unhandled", []interface{}{struct{}{}}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			visited := make(map[uintptr]bool)
			gotCount, err := countSliceInterfaceSimple(st, tt.input, visited)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotCount != tt.wantCount {
				t.Errorf("want count %d, got %d", tt.wantCount, gotCount)
			}
		})
	}
}

func TestCountTokensReflectHelpers(t *testing.T) {
	st := NewSimpleTokenizer()

	t.Run("Struct", func(t *testing.T) {
		type TestStruct struct {
			StrField  string
			BoolField bool
			IntField  int
		}

		val := reflect.ValueOf(TestStruct{
			StrField:  "hello",
			BoolField: true,
			IntField:  42,
		})

		count, err := countTokensReflectStruct(st, val, make(map[uintptr]bool))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if count != 3 {
			t.Errorf("want count 3, got %d", count)
		}
	})

	t.Run("Slice", func(t *testing.T) {
		val := reflect.ValueOf([]interface{}{
			"hello", // 1
			true,    // "true"->1
			42,      // 1
		})

		count, err := countTokensReflectSlice(st, val, make(map[uintptr]bool))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if count != 3 {
			t.Errorf("want count 3, got %d", count)
		}
	})

	t.Run("Map", func(t *testing.T) {
		val := reflect.ValueOf(map[interface{}]interface{}{
			"k1": "v1", // string -> string: 1+1=2
			"k2": true, // string -> bool: 1+1=2
			42:   "v3", // int -> string: 1+1=2
			43:   44,   // int -> int: 1+1=2
		})

		count, err := countTokensReflectMap(st, val, make(map[uintptr]bool))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if count != 8 { // 4 pairs of 2 tokens = 8
			t.Errorf("want count 8, got %d", count)
		}
	})

	t.Run("Map Cycle", func(t *testing.T) {
		m := make(map[string]interface{})
		m["self"] = m
		val := reflect.ValueOf(m)
		visited := make(map[uintptr]bool)
		visited[val.Pointer()] = true

		_, err := countTokensReflectMap(st, val, visited)
		if err == nil {
			t.Errorf("expected cycle error")
		}
	})

	t.Run("Slice Cycle", func(t *testing.T) {
		s := make([]interface{}, 1)
		s[0] = s
		val := reflect.ValueOf(s)
		visited := make(map[uintptr]bool)
		visited[val.Pointer()] = true

		_, err := countTokensReflectSlice(st, val, visited)
		if err == nil {
			t.Errorf("expected cycle error")
		}
	})
}

func TestCountTokensReflectSliceCoverage(t *testing.T) {
	st := NewSimpleTokenizer()

	val := reflect.ValueOf([]string{"hello", "world"})
	count, err := countTokensReflectSlice(st, val, make(map[uintptr]bool))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("want count 2, got %d", count)
	}
}

func TestCountTokensReflectMapCoverage(t *testing.T) {
	st := NewSimpleTokenizer()
	visited := make(map[uintptr]bool)

	// Valid map string key -> string val
	m1 := map[string]string{"k": "v"}
	c, err := countTokensReflectMap(st, reflect.ValueOf(m1), visited)
	if err != nil || c != 2 {
		t.Errorf("expected 2, nil, got %v, %v", c, err)
	}

	// Valid map bool key -> bool val (key is not string, falls into else branch)
	m2 := map[bool]bool{true: false}
	c, err = countTokensReflectMap(st, reflect.ValueOf(m2), visited)
	if err != nil || c != 2 {
		t.Errorf("expected 2, nil, got %v, %v", c, err)
	}

	// Valid map interface key -> bool val (value fallback branch)
	m3 := map[int]int{1: 2}
	c, err = countTokensReflectMap(st, reflect.ValueOf(m3), visited)
	if err != nil || c != 2 {
		t.Errorf("expected 2, nil, got %v, %v", c, err)
	}

	// Error path tests
	et := errTokenizer{}

	// Error on key count
	c, err = countTokensReflectMap(et, reflect.ValueOf(m1), visited)
	if err == nil {
		t.Errorf("expected error on key count")
	}

	c, err = countTokensReflectMap(et, reflect.ValueOf(m2), visited)
	if err == nil {
		t.Errorf("expected error on non-string key count")
	}
}

type errTokenizer struct{}

func (e errTokenizer) CountTokens(s string) (int, error) {
	return 0, fmt.Errorf("tokenizer error")
}

func (e errTokenizer) countRecursive(v interface{}, visited map[uintptr]bool) (int, error) {
	return 0, fmt.Errorf("tokenizer error")
}

type errValTokenizer struct{}

func (e errValTokenizer) CountTokens(s string) (int, error) {
	return 1, nil
}

func (e errValTokenizer) countRecursive(v interface{}, visited map[uintptr]bool) (int, error) {
	return 0, fmt.Errorf("val error")
}

func TestCountTokensReflectMapValueErrors(t *testing.T) {
	et := errValTokenizer{}
	visited := make(map[uintptr]bool)

	// String key (succeeds), int value (fails)
	m := map[string]int{"k": 1}
	_, err := countTokensReflectMap(et, reflect.ValueOf(m), visited)
	if err == nil {
		t.Errorf("expected error on value count")
	}
}
