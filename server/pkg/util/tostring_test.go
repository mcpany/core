package util

import (
	"encoding/json"
	"math"
	"testing"
)

func TestToString_FloatFormatting(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "Small integer float",
			input:    float64(123),
			expected: "123",
		},
		{
			name:     "Large integer float (1 million)",
			input:    float64(1000000),
			expected: "1000000",
		},
		{
			name:     "Large integer float (1.23 billion)",
			input:    float64(1234567890),
			expected: "1234567890",
		},
		{
			name:     "Float with decimal",
			input:    float64(123.45),
			expected: "123.45",
		},
		{
			name:     "Negative integer float",
			input:    float64(-1000000),
			expected: "-1000000",
		},
		{
			name:     "Zero float",
			input:    float64(0),
			expected: "0",
		},
		{
			name:     "Very large float (exceeds int64)",
			input:    math.MaxFloat64,
			expected: "1.7976931348623157e+308", // Expect scientific notation
		},
		{
			name:     "Float32 integer",
			input:    float32(123456),
			expected: "123456",
		},
		{
			name:     "Float32 decimal",
			input:    float32(123.45),
			expected: "123.45",
		},
		{
			name:     "Very large float32 (exceeds int64)",
			input:    float32(math.MaxFloat32),
			expected: "3.4028235e+38", // Approx
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToString(tt.input); got != tt.expected {
				t.Errorf("ToString(%v) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestReplaceURLPath_FloatIntegration(t *testing.T) {
	params := map[string]interface{}{
		"id": float64(1000000),
	}
	path := "/users/{{id}}"
	result := ReplaceURLPath(path, params, nil)
	expected := "/users/1000000"
	if result != expected {
		t.Errorf("ReplaceURLPath(%q, %v) = %q, want %q", path, params, result, expected)
	}
}

func TestToString_OtherTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"String", "hello", "hello"},
		{"Bool true", true, "true"},
		{"Bool false", false, "false"},
		{"Int", 123, "123"},
		{"Int8", int8(127), "127"},
		{"Int64", int64(1234567890123), "1234567890123"},
		{"Uint", uint(123), "123"},
		{"JSON Number", json.Number("123.456"), "123.456"},
		{"Nil", nil, "<nil>"}, // fmt.Sprintf("%v", nil) -> "<nil>"
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToString(tt.input); got != tt.expected {
				t.Errorf("ToString(%v) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
