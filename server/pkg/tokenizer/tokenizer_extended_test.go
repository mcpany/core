package tokenizer_test

import (
	"math"
	"testing"

	"github.com/mcpany/core/server/pkg/tokenizer"
	"github.com/stretchr/testify/assert"
)

// MockTokenizer is a tokenizer that is neither SimpleTokenizer nor WordTokenizer.
type MockTokenizer struct{}

func (m *MockTokenizer) CountTokens(text string) (int, error) {
	return len(text), nil
}

func TestCountTokensInValue_GenericFallback(t *testing.T) {
	tok := &MockTokenizer{}

	tests := []struct {
		name     string
		input    interface{}
		expected int
	}{
		{"string", "hello", 5},
		{"int", 123, 3},
		{"int64", int64(1234), 4},
		{"float", 3.14, 4},
		{"bool_true", true, 4},
		{"bool_false", false, 5},
		{"nil", nil, 4}, // "null"
		{"slice", []interface{}{"a", "bb"}, 1 + 2},
		{"map", map[string]interface{}{"a": "b"}, 1 + 1},
		{"struct", struct{ A int }{1}, 1}, // "1" (values only, consistent with map content)
		{"ptr_struct", &struct{ A int }{1}, 1},
		{"ptr_int", func() *int { i := 123; return &i }(), 3}, // "123" -> 3
		{"ptr_nil", (*int)(nil), 4},                           // "null"
		{"nested_ptr", func() **int { i := 123; p := &i; return &p }(), 3},
		{"struct_unexported", struct{ a int }{1}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tokenizer.CountTokensInValue(tok, tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestWordTokenizer_Types(t *testing.T) {
	tok := tokenizer.NewWordTokenizer()
	// WordTokenizer factor is 1.3. Primitive strings are short.
	// "123" -> 1 word * 1.3 = 1.3 -> 1 token.

	tests := []struct {
		name     string
		input    interface{}
		expected int
	}{
		{"int64", int64(999), 1},
		{"float64", 123.456, 1},
		{"bool_false", false, 1}, // "false" -> 1 word -> 1.3 -> 1
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tokenizer.CountTokensInValue(tok, tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestSimpleTokenizeInt_EdgeCases(t *testing.T) {
	tok := tokenizer.NewSimpleTokenizer()

	tests := []struct {
		name     string
		input    interface{}
		expected int
	}{
		{"zero", 0, 1},
		{"zero_int64", int64(0), 1},
		{"negative_small", -5, 1}, // "-5" -> 2 chars -> 1 token
		{"negative_large", -12345, 1}, // "-12345" -> 6 chars -> 1.5 -> 1 token? No, 6/4 = 1.
		{"min_int", math.MinInt, 5}, // "-9223372036854775808" -> 20 chars -> 5 tokens
		{"max_int", math.MaxInt, 4}, // "9223372036854775807" -> 19 chars -> 4.75 -> 4 tokens
		{"min_int64", int64(math.MinInt64), 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tokenizer.CountTokensInValue(tok, tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, got, "input: %v", tt.input)
		})
	}
}
