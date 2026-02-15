package tokenizer_test

import (
	"testing"

	"github.com/mcpany/core/server/pkg/tokenizer"
)

func TestCountTokensInValue_Types(t *testing.T) {
	tok := tokenizer.NewSimpleTokenizer()

	tests := []struct {
		name     string
		input    interface{}
		expected int // "true" -> 4/4 = 1, "123" -> 3/4 -> 1
	}{
		{"int", 12345, 1},        // "12345" len 5
		{"int64", int64(123), 1}, // "123" len 3
		{"float", 3.14, 1},       // "3.14" len 4
		{"float_sci", 1e20, 1},   // "1e+20" len 5 -> 1. If 'f', it would be 21 chars -> 5 tokens
		{"bool_true", true, 1},   // "true" len 4
		{"bool_false", false, 1}, // "false" len 5 -> 1
		{"nil", nil, 1},          // "null" len 4
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tokenizer.CountTokensInValue(tok, tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("CountTokensInValue(%v) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}
