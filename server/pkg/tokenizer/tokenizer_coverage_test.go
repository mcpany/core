package tokenizer

import (
	"errors"
	"reflect"
	"testing"
)

// The simpleTokenizeInt64 unrolled loop has a default case
// which matches numbers larger than what can be typed as a literal in go sometimes,
// or at least is only reached for things with 19+ digits.
func TestSimpleTokenizeInt64Max(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected int
	}{
		{"min_int64", -9223372036854775808, 5},
		{"zero", 0, 1},
		{"digit_1", 9, 1},
		{"digit_2", 99, 1},
		{"digit_3", 999, 1},
		{"digit_4", 9999, 1},
		{"digit_5", 99999, 1},
		{"digit_6", 999999, 1},
		{"digit_7", 9999999, 1},
		{"digit_8", 99999999, 2},
		{"digit_9", 999999999, 2},
		{"digit_10", 9999999999, 2},
		{"digit_11", 99999999999, 2},
		{"digit_12", 999999999999, 3},
		{"digit_13", 9999999999999, 3},
		{"digit_14", 99999999999999, 3},
		{"digit_15", 999999999999999, 3},
		{"digit_16", 9999999999999999, 4},
		{"digit_17", 99999999999999999, 4},
		{"digit_18", 999999999999999999, 4},
		// Added coverage for negative large magnitudes (requires > 8 chars path)
		{"min_int64_2", -9000000000000000000, 5},
		{"digit_minus_8", -1000000, 2},
		{"digit_minus_9", -10000000, 2},
		{"digit_minus_10", -100000000, 2},
		{"digit_minus_11", -1000000000, 2},
		{"digit_minus_12", -10000000000, 3},
		{"digit_minus_13", -100000000000, 3},
		{"digit_minus_14", -1000000000000, 3},
		{"digit_minus_15", -10000000000000, 3},
		{"digit_minus_16", -100000000000000, 4},
		{"digit_minus_17", -1000000000000000, 4},
		{"digit_minus_18", -10000000000000000, 4},
		{"digit_minus_19", -100000000000000000, 4},
		{"digit_minus_20", -1000000000000000000, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := simpleTokenizeInt64(tt.input); got != tt.expected {
				t.Errorf("simpleTokenizeInt64() = %v, want %v", got, tt.expected)
			}
		})
	}
}

type MyStruct struct {
	Name string
}

func TestCountTokensReflectSliceCoverage3(t *testing.T) {
	st := NewSimpleTokenizer()

	tests := []struct {
		name        string
		tokenizer   recursiveTokenizer
		input       interface{}
		expectError bool
		expectCount int
	}{
		{
			name:        "nil_slice",
			tokenizer:   st,
			input:       []string(nil),
			expectError: false,
			expectCount: 0,
		},
		{
			name:        "array",
			tokenizer:   st,
			input:       [1]string{"test"},
			expectError: false,
			expectCount: 1,
		},
		{
			name:        "failing_tokenizer_generic_item",
			tokenizer:   &failingTokenizer2{},
			input:       []MyStruct{{Name: "test"}},
			expectError: true,
			expectCount: 0,
		},
		{
			name:        "failing_tokenizer_string_item",
			tokenizer:   &failingTokenizer3{},
			input:       []string{"test"},
			expectError: true,
			expectCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := reflect.ValueOf(tt.input)
			got, err := countTokensReflectSlice(tt.tokenizer, v, make(map[uintptr]bool))
			if (err != nil) != tt.expectError {
				t.Errorf("countTokensReflectSlice() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if got != tt.expectCount {
				t.Errorf("countTokensReflectSlice() = %v, want %v", got, tt.expectCount)
			}
		})
	}
}

func TestCountTokensReflectMapCoverage3(t *testing.T) {
	st := NewSimpleTokenizer()
	failTok := &failingTokenizer2{}
	failTok3 := &failingTokenizer3{}

	// Create cycle
	mCycle := make(map[string]interface{})
	vCycle := reflect.ValueOf(mCycle)
	visited := make(map[uintptr]bool)
	visited[vCycle.Pointer()] = true

	mCycle2 := make(map[int]interface{})
	vCycle2 := reflect.ValueOf(mCycle2)
	visited2 := make(map[uintptr]bool)
	visited2[vCycle2.Pointer()] = true

	tests := []struct {
		name        string
		tokenizer   recursiveTokenizer
		input       interface{}
		visited     map[uintptr]bool
		expectError bool
		expectCount int
	}{
		{
			name:        "nil_map",
			tokenizer:   st,
			input:       map[string]string(nil),
			visited:     make(map[uintptr]bool),
			expectError: false,
			expectCount: 0,
		},
		{
			name:        "map_cycle_string_key",
			tokenizer:   st,
			input:       mCycle,
			visited:     visited,
			expectError: true,
			expectCount: 0,
		},
		{
			name:        "map_cycle_generic_key",
			tokenizer:   st,
			input:       mCycle2,
			visited:     visited2,
			expectError: true,
			expectCount: 0,
		},
		{
			name:        "string_key_bool_value",
			tokenizer:   st,
			input:       map[string]bool{"test": true},
			visited:     make(map[uintptr]bool),
			expectError: false,
			expectCount: 2,
		},
		{
			name:        "generic_key_bool_value",
			tokenizer:   st,
			input:       map[int]bool{1: true},
			visited:     make(map[uintptr]bool),
			expectError: false,
			expectCount: 2,
		},
		{
			name:        "generic_value_error",
			tokenizer:   failTok,
			input:       map[string]int{"test": 1},
			visited:     make(map[uintptr]bool),
			expectError: true,
			expectCount: 0,
		},
		{
			name:        "generic_key_error",
			tokenizer:   failTok,
			input:       map[int]int{1: 1},
			visited:     make(map[uintptr]bool),
			expectError: true,
			expectCount: 0,
		},
		{
			name:        "non_string_map_key_error",
			tokenizer:   failTok,
			input:       map[bool]string{true: "test"},
			visited:     make(map[uintptr]bool),
			expectError: true,
			expectCount: 0,
		},
		{
			name:        "failing_tokenizer_string_key",
			tokenizer:   failTok3,
			input:       map[string]int{"test": 1},
			visited:     make(map[uintptr]bool),
			expectError: true,
			expectCount: 0,
		},
		{
			name:        "failing_tokenizer_string_val",
			tokenizer:   failTok3,
			input:       map[int]string{1: "test"},
			visited:     make(map[uintptr]bool),
			expectError: true,
			expectCount: 0,
		},
		{
			name:        "failing_tokenizer_bool_val",
			tokenizer:   failTok3,
			input:       map[int]bool{1: true, 2: false},
			visited:     make(map[uintptr]bool),
			expectError: true,
			expectCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := reflect.ValueOf(tt.input)
			got, err := countTokensReflectMap(tt.tokenizer, v, tt.visited)
			if (err != nil) != tt.expectError {
				t.Errorf("countTokensReflectMap() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if got != tt.expectCount {
				t.Errorf("countTokensReflectMap() = %v, want %v", got, tt.expectCount)
			}
		})
	}
}

type failingTokenizer2 struct{}

func (f *failingTokenizer2) CountTokens(text string) (int, error) {
	return 1, nil // succeed on text
}
func (f *failingTokenizer2) countRecursive(v interface{}, visited map[uintptr]bool) (int, error) {
	return 0, errors.New("fake error generic") // fail on generic recursive
}

type failingTokenizer3 struct{}

func (f *failingTokenizer3) CountTokens(text string) (int, error) {
	return 0, errors.New("fake error")
}
func (f *failingTokenizer3) countRecursive(v interface{}, visited map[uintptr]bool) (int, error) {
	return 0, errors.New("fake error")
}
