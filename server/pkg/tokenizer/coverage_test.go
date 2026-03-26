// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tokenizer

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// GenericTokenizer implements Tokenizer but is neither Simple nor Word tokenizer.
// GenericTokenizer implements Tokenizer but is neither Simple nor Word tokenizer.
// Summary: GenericTokenizer

// CountTokens ...
// Summary: CountTokens
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return len(text), nil
}

// TestCoverage_GenericTokenizer_Primitives ...
// Summary: TestCoverage_GenericTokenizer_Primitives
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	gt := &GenericTokenizer{}

	tests := []struct {
		name     string
		input    interface{}
		expected int
	}{
		{"int", 123, 3}, // "123" -> 3
		{"int64", int64(12345), 5},
		{"float64", 12.34, 5},    // "12.34" -> 5 (depending on formatting)
		{"bool_true", true, 4},   // "true" -> 4
		{"bool_false", false, 5}, // "false" -> 5
		{"nil", nil, 4},          // "null" -> 4
		{"string", "abc", 3},
		{"slice_string", []string{"a", "b"}, 2},               // 1+1
		{"map_string_string", map[string]string{"a": "b"}, 2}, // "a"(1) + "b"(1)
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, err := CountTokensInValue(gt, tc.input)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, c)
		})
	}
}

// TestCoverage_WordTokenizer_FastPath_Slices ...
// Summary: TestCoverage_WordTokenizer_FastPath_Slices
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Set Factor to 2.0, so primitiveCount becomes 2.
	wt := NewWordTokenizer()
	wt.Factor = 2.0

	tests := []struct {
		name     string
		input    interface{}
		expected int
	}{
		{"slice_int", []int{1, 2, 3}, 6},                      // 3 * 2
		{"slice_int64", []int64{1, 2}, 4},                     // 2 * 2
		{"slice_float64", []float64{1.0}, 2},                  // 1 * 2
		{"slice_bool", []bool{true, false}, 4},                // 2 * 2
		{"map_string_string", map[string]string{"a": "b"}, 4}, // "a"(1*2) + "b"(1*2) = 4
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, err := CountTokensInValue(wt, tc.input)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, c)
		})
	}
}

// TestCoverage_SimpleTokenizer_FastPath_Slices ...
// Summary: TestCoverage_SimpleTokenizer_FastPath_Slices
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	st := NewSimpleTokenizer()

	tests := []struct {
		name     string
		input    interface{}
		expected int
	}{
		{"slice_int", []int{123, 4567}, 1 + 1}, // 123->1, 4567->1
		{"slice_int64", []int64{123, 4567}, 1 + 1},
		{"slice_bool", []bool{true, false}, 2},
		{"slice_float64", []float64{1.1, 2.2}, 2},
		{"map_string_string", map[string]string{"a": "b"}, 2}, // "a"->1, "b"->1
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, err := CountTokensInValue(st, tc.input)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, c)
		})
	}
}

// CoverageStringerImpl ...
// Summary: CoverageStringerImpl

// String ...
// Summary: String
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.

// TestCoverage_Reflect_Stringer_Error ...
// Summary: TestCoverage_Reflect_Stringer_Error
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	gt := &GenericTokenizer{}

	// Test Stringer
	c, err := CountTokensInValue(gt, &CoverageStringerImpl{})
	assert.NoError(t, err)
	assert.Equal(t, 8, c) // "stringer"

	// Test Error
	errVal := fmt.Errorf("error")
	c, err = CountTokensInValue(gt, errVal)
	assert.NoError(t, err)
	assert.Equal(t, 5, c) // "error"
}

// TestCoverage_Reflect_Fallback_Chan ...
// Summary: TestCoverage_Reflect_Fallback_Chan
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	gt := &GenericTokenizer{}
	ch := make(chan int)

	// Should fallback to fmt.Sprintf which prints pointer address usually?
	// "0xc0..."
	c, err := CountTokensInValue(gt, ch)
	assert.NoError(t, err)
	assert.Greater(t, c, 0)
}
