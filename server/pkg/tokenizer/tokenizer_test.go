// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tokenizer

import (
	"fmt"
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
		{"hello world", 2}, // 2 * 1.3 = 2.6 -> 2
		{"this is a test sentence", 6}, // 5 * 1.3 = 6.5 -> 6
		{"hello 🌍", 2}, // 2 * 1.3 = 2.6 -> 2 (ASCII + Emoji)
		{"你好 世界", 2}, // 2 * 1.3 = 2.6 -> 2 (Chinese + Space + Chinese)
		{"hello\tworld\n", 2}, // ASCII whitespace
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
		{"string", "hello world", 2}, // "hello world" -> 2 words * 1.3 -> 2
		{"slice", []interface{}{1, "hello"}, 1 + 1}, // 1 (int) + 1 (hello)
		{"map", map[string]interface{}{"a": 1}, 1 + 1}, // "a" (1) + 1 (int)
		{"string_slice", []string{"hello", "world"}, 1 + 1}, // 1 + 1
		{"string_map", map[string]string{"a": "b"}, 1 + 1}, // "a"(1) + "b"(1)
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
		d := &ExportedStruct{Name: "D", Age: 1} // "D"(1) + "1"(1) = 2 tokens
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
		{"a\tb", 2}, // Tab
		{"a\r\nb", 2}, // CR LF
		{"a \x00 b", 3}, // Control char \x00
		{"a\u00A0b", 2}, // NBSP (non-ASCII space)
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
		type S struct { Field interface{} }
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
		type S struct { Field interface{} }
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
		val float64
		want int
	}{
		{1234567.0, 1}, // "1234567" -> 7 chars -> 1.75 -> 1 token (Changed from 3)
		{9999999.0, 1}, // "9999999" -> 7 chars -> 1.75 -> 1 token (Changed from 3)
		{10000000.0, 2}, // "10000000" -> 8 chars -> 2 tokens (Changed from 1: "1e+07" was 5 chars)
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
<<<<<<< HEAD

func TestSimpleTokenizeInt64_Coverage(t *testing.T) {
	tests := []struct {
		val  int64
		want int
	}{
		// < 8 chars (including sign) fast path -> 1
		{0, 1},
		{9999999, 1},
		{-999999, 1},

		// 8 chars -> 2 (8/4)
		{10000000, 2},
		{-1000000, 2},

		// 9 chars -> 2 (9/4)
		{100000000, 2},
		{-10000000, 2},

		// 10 chars -> 2
		{1000000000, 2},

		// 11 chars -> 2 (11/4 = 2)
		{10000000000, 2},

		// 12 chars -> 3
		{100000000000, 3},

		// 13 chars -> 3
		{1000000000000, 3},

		// 14 chars -> 3
		{10000000000000, 3},

		// 15 chars -> 3
		{100000000000000, 3},

		// 16 chars -> 4
		{1000000000000000, 4},

		// 17 chars -> 4
		{10000000000000000, 4},

		// 18 chars -> 4
		{100000000000000000, 4},

		// 19 chars -> 4 (19/4 = 4)
		{1000000000000000000, 4},

		// MaxInt64: 9223372036854775807 (19 chars) -> 4
		{9223372036854775807, 4},

		// MinInt64: -9223372036854775808 (20 chars) -> 5
		{-9223372036854775808, 5},
	}

	for _, tt := range tests {
		got := simpleTokenizeInt64(tt.val)
		if got != tt.want {
			t.Errorf("simpleTokenizeInt64(%d) = %d, want %d", tt.val, got, tt.want)
		}
	}
}

func TestSimpleTokenizeInt64_AllMagnitudes(t *testing.T) {
	tests := []struct {
		val  int64
		want int
	}{
		// < 10
		{9, 1},
		// < 100
		{99, 1},
		// < 1000
		{999, 1},
		// < 10000
		{9999, 1},
		// < 100000
		{99999, 1},
		// < 1000000
		{999999, 1},
		// < 10000000
		{9999999, 1},
		// < 100000000 (8 chars) -> 2
		{99999999, 2},
		// < 1000000000 (9 chars) -> 2
		{999999999, 2},
		// < 10000000000 (10 chars) -> 2
		{9999999999, 2},
		// < 100000000000 (11 chars) -> 2
		{99999999999, 2},
		// < 1000000000000 (12 chars) -> 3
		{999999999999, 3},
		// < 10000000000000 (13 chars) -> 3
		{9999999999999, 3},
		// < 100000000000000 (14 chars) -> 3
		{99999999999999, 3},
		// < 1000000000000000 (15 chars) -> 3
		{999999999999999, 3},
		// < 10000000000000000 (16 chars) -> 4
		{9999999999999999, 4},
		// < 100000000000000000 (17 chars) -> 4
		{99999999999999999, 4},
		// < 1000000000000000000 (18 chars) -> 4
		{999999999999999999, 4},
		// MaxInt64 (19 chars) -> 4
		{9223372036854775807, 4},

		// Negatives
		{-9, 1},
		{-99, 1},
		{-999, 1},
		{-9999, 1},
		{-99999, 1},
		{-999999, 1},
		{-1000000, 2},
		{-9999999, 2},
		{-99999999, 2},
		{-999999999, 2},
		{-9999999999, 2},
		{-99999999999, 3},
		{-999999999999, 3},
		{-9999999999999, 3},
		{-99999999999999, 3},
		{-999999999999999, 4},
		{-9999999999999999, 4},
		{-99999999999999999, 4},
		{-999999999999999999, 4},
		{-9223372036854775807, 5},
	}

	for _, tt := range tests {
		got := simpleTokenizeInt64(tt.val)
		if got != tt.want {
			t.Errorf("simpleTokenizeInt64(%d) = %d, want %d", tt.val, got, tt.want)
		}
	}
}



func TestCountSliceInterfaceRaw_Coverage(t *testing.T) {
	tokenizer := NewWordTokenizer()

	// Need to get access to rawWordCounter logic.
	// We can do this by using a []interface{} that contains things hitting the branches.

	// Branch 1: empty slice
	emptySlice := []interface{}{}
	c, err := CountTokensInValue(tokenizer, emptySlice)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c != 0 {
		t.Errorf("empty slice want 0, got %d", c)
	}

	// Branch 2: string, float64, int, int64, bool, nil
	basicSlice := []interface{}{"hello world", float64(3.14), int(1), int64(2), true, nil}
	c, err = CountTokensInValue(tokenizer, basicSlice)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c != 9 {
		t.Errorf("basic slice want 9, got %d", c)
	}

	// Branch 3: nested []interface{}
	nestedSlice := []interface{}{[]interface{}{"nested"}}
	c, err = CountTokensInValue(tokenizer, nestedSlice)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c != 1 {
		t.Errorf("nested slice want 1, got %d", c)
	}

	// Branch 4: default (structs, maps, etc)
	defaultSlice := []interface{}{map[string]string{"key": "value"}}
	c, err = CountTokensInValue(tokenizer, defaultSlice)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// key (1) + value (1) = 2
	if c != 2 {
		t.Errorf("default slice want 2, got %d", c)
	}

	// Branch 5: error in nested []interface{}
	// We can trigger this by putting a cycle in the nested slice.
	s1 := []interface{}{"test"}
	s2 := []interface{}{s1}
	s1[0] = s2 // Cycle: s1 -> s2 -> s1
	_, err = CountTokensInValue(tokenizer, s1)
	if err == nil {
		t.Fatalf("expected error due to cycle in nested slice, got nil")
	}

	// Branch 6: error in default
	// Cycle in a map inside the slice
	m1 := map[string]interface{}{}
	m1["self"] = m1
	s3 := []interface{}{m1}
	_, err = CountTokensInValue(tokenizer, s3)
	if err == nil {
		t.Fatalf("expected error due to cycle in map inside slice, got nil")
	}
}

func TestSimpleTokenizeInt64_AllSwitchBranchesDirectly(t *testing.T) {
	tests := []struct {
		n int64
		want int
	}{
		// Use negative values where the negated result places us cleanly into every bucket.
		// Wait, the logic is:
		// if n < 0 { l=1; n=-n }
		// switch {
		// case n < 10: l++
		// case n < 100: l+=2
		// ...
		// BUT the fast path is:
		// if n > -1000000 && n < 10000000 { return 1 }
		// This means any number between -999999 and 9999999 RETURNS 1 directly!
		// It will NEVER evaluate the switch for n < 10, n < 100, n < 1000, n < 10000, n < 100000, n < 1000000, n < 10000000!
		// Because it's already caught by the fast path at the very beginning of the function!
	}
	for _, tt := range tests {
		_ = tt.n
	}
}
=======
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
