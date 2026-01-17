package util

import (
	"testing"
)

func TestToString_NoScientificNotation(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "large int as float",
			input:    1000000000.0,
			expected: "1000000000",
		},
		{
			name:     "float with decimals",
			input:    123.456,
			expected: "123.456",
		},
		{
			name:     "large float with decimals",
			input:    12345678.123,
			expected: "12345678.123",
		},
		{
			name:     "float32 large",
			input:    float32(1000000.0),
			expected: "1000000",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ToString(tc.input)
			if got != tc.expected {
				t.Errorf("ToString(%v) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}
