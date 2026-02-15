package tokenizer

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// GenericTokenizer implements Tokenizer but is neither Simple nor Word tokenizer.
type GenericTokenizer struct{}

func (t *GenericTokenizer) CountTokens(text string) (int, error) {
	return len(text), nil
}

func TestCoverage_GenericTokenizer_Primitives(t *testing.T) {
	gt := &GenericTokenizer{}

	tests := []struct {
		name     string
		input    interface{}
		expected int
	}{
		{"int", 123, 3}, // "123" -> 3
		{"int64", int64(12345), 5},
		{"float64", 12.34, 5}, // "12.34" -> 5 (depending on formatting)
		{"bool_true", true, 4}, // "true" -> 4
		{"bool_false", false, 5}, // "false" -> 5
		{"nil", nil, 4}, // "null" -> 4
        {"string", "abc", 3},
        {"slice_string", []string{"a", "b"}, 2}, // 1+1
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

func TestCoverage_WordTokenizer_FastPath_Slices(t *testing.T) {
    // Set Factor to 2.0, so primitiveCount becomes 2.
    wt := NewWordTokenizer()
    wt.Factor = 2.0

    tests := []struct {
        name string
        input interface{}
        expected int
    }{
        {"slice_int", []int{1, 2, 3}, 6}, // 3 * 2
        {"slice_int64", []int64{1, 2}, 4}, // 2 * 2
        {"slice_float64", []float64{1.0}, 2}, // 1 * 2
        {"slice_bool", []bool{true, false}, 4}, // 2 * 2
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

func TestCoverage_SimpleTokenizer_FastPath_Slices(t *testing.T) {
    st := NewSimpleTokenizer()

    tests := []struct {
        name string
        input interface{}
        expected int
    }{
        {"slice_int", []int{123, 4567}, 1+1}, // 123->1, 4567->1
        {"slice_int64", []int64{123, 4567}, 1+1},
        {"slice_bool", []bool{true, false}, 2},
        {"slice_float64", []float64{1.1, 2.2}, 2},
        {"map_string_string", map[string]string{"a":"b"}, 2}, // "a"->1, "b"->1
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            c, err := CountTokensInValue(st, tc.input)
            assert.NoError(t, err)
            assert.Equal(t, tc.expected, c)
        })
    }
}

type CoverageStringerImpl struct{}
func (s *CoverageStringerImpl) String() string { return "stringer" }

func TestCoverage_Reflect_Stringer_Error(t *testing.T) {
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

func TestCoverage_Reflect_Fallback_Chan(t *testing.T) {
    gt := &GenericTokenizer{}
    ch := make(chan int)

    // Should fallback to fmt.Sprintf which prints pointer address usually?
    // "0xc0..."
    c, err := CountTokensInValue(gt, ch)
    assert.NoError(t, err)
    assert.Greater(t, c, 0)
}
