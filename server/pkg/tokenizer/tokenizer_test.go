// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tokenizer

import "testing"

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
		"key": "abcdefgh", // 2
		"list": []interface{}{
			"abcd", // 1
			"1234", // 1
		},
	}
	// Total: 2 + 1 + 1 = 4

	got, _ := CountTokensInValue(tokenizer, val)
	if got != 4 {
		t.Errorf("CountTokensInValue = %d, want 4", got)
	}
}
