package tokenizer_test

import (
	"testing"

	"github.com/mcpany/core/server/pkg/tokenizer"
	"github.com/stretchr/testify/assert"
)

// Structs for testing reflection
type TestStruct struct {
	Name string
	Val  int
}

type NestedStruct struct {
	Child *TestStruct
}

func TestSimpleTokenizer_ReflectionCoverage(t *testing.T) {
	tok := tokenizer.NewSimpleTokenizer()

	tests := []struct {
		name     string
		input    interface{}
		expected int
	}{
		// Pointers
		{"ptr_nil", (*int)(nil), 1}, // "null" -> 1
		{"ptr_val", func() *int { i := 123; return &i }(), 1}, // "123" -> 3/4 = 0 -> 1

		// Structs
		{"struct", TestStruct{Name: "hello", Val: 123}, 2}, // "hello"(2) + "123"(1) = 3? No: "hello" is 5 chars -> 1. 123 is 3 chars -> 1. Total 2.

		// Slices/Arrays via reflection
		{"slice_struct", []TestStruct{{Name: "a", Val: 1}}, 2},
		{"array_struct", [1]TestStruct{{Name: "a", Val: 1}}, 2},

		// Maps via reflection
		{"map_int_string", map[int]string{1: "a"}, 2}, // "1"(1) + "a"(1) = 2
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tokenizer.CountTokensInValue(tok, tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestWordTokenizer_ReflectionCoverage(t *testing.T) {
	tok := tokenizer.NewWordTokenizer()

	tests := []struct {
		name     string
		input    interface{}
		expected int
	}{
		// Pointers
		{"ptr_nil", (*int)(nil), 1}, // "null" -> 1
		{"ptr_val", func() *int { i := 123; return &i }(), 1}, // 123 -> 1

		// Structs
		{"struct", TestStruct{Name: "hello", Val: 123}, 2}, // "hello"(1) + 123(1) = 2
		{"struct_nested", NestedStruct{Child: &TestStruct{Name: "a", Val: 1}}, 2}, // "a"(1) + 1(1) = 2

		// Slices/Arrays via reflection (when not caught by type switch)
		// Note: []string and []interface{} are caught by type switch.
		// We need a slice of a struct or custom type to hit reflect path.
		{"slice_struct", []TestStruct{{Name: "a", Val: 1}}, 2},
		{"array_struct", [1]TestStruct{{Name: "a", Val: 1}}, 2},

		// Maps via reflection
		// map[string]interface{} and map[string]string are caught.
		// map[int]string needs reflection.
		{"map_int_string", map[int]string{1: "a"}, 2}, // 1(1) + "a"(1) = 2

		// Cycle detection
		// "cycle" test needs manual construction below
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tokenizer.CountTokensInValue(tok, tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}

	// Cycle detection test
	t.Run("cycle_ptr", func(t *testing.T) {
		type Node struct {
			Next *Node
		}
		n := &Node{}
		n.Next = n
		_, err := tokenizer.CountTokensInValue(tok, n)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cycle detected")
	})
}

// MockTokenizer for generic reflection path
type MockReflectTokenizer struct{}

func (m *MockReflectTokenizer) CountTokens(text string) (int, error) {
	return 1, nil
}

func TestGenericTokenizer_ReflectionCoverage(t *testing.T) {
	tok := &MockReflectTokenizer{}

	tests := []struct {
		name     string
		input    interface{}
		expected int
	}{
		// Pointers
		{"ptr_nil", (*int)(nil), 1}, // "null" -> 1
		{"ptr_val", func() *int { i := 123; return &i }(), 1}, // 123 -> "123" -> 1

		// Structs
		{"struct", TestStruct{Name: "hello", Val: 123}, 2}, // "hello"(1) + "123"(1) = 2

		// Slices/Arrays via reflection
		{"slice_struct", []TestStruct{{Name: "a", Val: 1}}, 2},
		{"array_struct", [1]TestStruct{{Name: "a", Val: 1}}, 2},

		// Maps via reflection
		{"map_int_string", map[int]string{1: "a"}, 2}, // "1"(1) + "a"(1) = 2
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tokenizer.CountTokensInValue(tok, tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}

	// Cycle detection test
	t.Run("cycle_ptr", func(t *testing.T) {
		type Node struct {
			Next *Node
		}
		n := &Node{}
		n.Next = n
		_, err := tokenizer.CountTokensInValue(tok, n)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cycle detected")
	})
}
