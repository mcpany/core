// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tokenizer

import (
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
