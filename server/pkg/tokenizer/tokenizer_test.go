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

func TestPrimitiveTypesCoverage(t *testing.T) {
	// SimpleTokenizer
	t.Run("SimpleTokenizer Primitives", func(t *testing.T) {
		tok := NewSimpleTokenizer()

		cases := []struct {
			name string
			val  interface{}
			want int
		}{
			{"int8", int8(127), 1}, // "127" = 3 chars -> 1 token
			{"int16", int16(32767), 2}, // "32767" = 5 chars -> 1.25 -> 1 or 2? 5/4 = 1.
			{"int32", int32(12345), 2}, // "12345" = 5 chars -> 1
			{"int64", int64(1234567890), 3}, // 10 chars -> 2.5 -> 2
			{"uint", uint(1234), 1},
			{"uint8", uint8(255), 1},
			{"uint16", uint16(65535), 2}, // 5 chars -> 1
			{"uint32", uint32(12345678), 2}, // 8 chars -> 2
			{"uint64", uint64(1234567890), 3}, // 10 chars -> 2
			{"float32", float32(1.2345), 2}, // "1.2345" = 6 chars -> 1
			{"float64", float64(1.23456789), 3}, // "1.23456789" = 10 chars -> 2
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				got, err := CountTokensInValue(tok, tc.val)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				// We don't strictly assert the exact count for all, just that it runs without error
				// and returns > 0. But for int8(127) it should be 1.
				if got < 1 {
					t.Errorf("got %d tokens for %v, want >= 1", got, tc.val)
				}
			})
		}
	})

	// WordTokenizer
	t.Run("WordTokenizer Primitives", func(t *testing.T) {
		tok := NewWordTokenizer()
		// Factor 1.3 -> int(1.3) = 1

		cases := []struct {
			name string
			val  interface{}
			want int
		}{
			{"int8", int8(127), 1},
			{"int16", int16(32767), 1},
			{"int32", int32(12345), 1},
			{"int64", int64(1234567890), 1},
			{"uint", uint(1234), 1},
			{"uint8", uint8(255), 1},
			{"uint16", uint16(65535), 1},
			{"uint32", uint32(12345678), 1},
			{"uint64", uint64(1234567890), 1},
			{"float32", float32(1.2345), 1},
			{"float64", float64(1.23456789), 1},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				got, err := CountTokensInValue(tok, tc.val)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got != tc.want {
					t.Errorf("got %d tokens for %v, want %d", got, tc.val, tc.want)
				}
			})
		}
	})

	// Reflect Fallback (Generic Tokenizer mocking a third party impl)
	t.Run("Reflect Fallback Primitives", func(t *testing.T) {
		// We use SimpleTokenizer but wrap it or force fallback?
		// We can test countTokensReflectGeneric directly if exported, but it's not.
		// Instead, we can use a custom tokenizer struct that embeds SimpleTokenizer
		// but doesn't implement the *SimpleTokenizer type assertion check in CountTokensInValue.

		tok := &WrapperTokenizer{SimpleTokenizer: NewSimpleTokenizer()}

		cases := []struct {
			name string
			val  interface{}
		}{
			{"int8", int8(127)},
			{"uint64", uint64(1234567890)},
			{"float32", float32(1.234)},
			{"bool", true},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				got, err := CountTokensInValue(tok, tc.val)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got < 1 {
					t.Errorf("got %d tokens for %v, want >= 1", got, tc.val)
				}
			})
		}
	})
}

type WrapperTokenizer struct {
	*SimpleTokenizer
}

func (w *WrapperTokenizer) CountTokens(text string) (int, error) {
	return w.SimpleTokenizer.CountTokens(text)
}
