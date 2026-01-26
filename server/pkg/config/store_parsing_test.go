package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnquoteCSV(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"foo"`, "foo"},
		{`"foo""bar"`, `foo"bar`},
		{`foo`, "foo"},
		{`"foo`, `"foo`},
		{`foo"`, `foo"`},
		{`""`, ""},
		{`""""`, `"`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, unquoteCSV(tt.input))
		})
	}
}

func TestSplitByCommaIgnoringBraces(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{`a,b,c`, []string{"a", "b", "c"}},
		{`a, b, c`, []string{"a", "b", "c"}},
		{`{"a": 1}, {"b": 2}`, []string{`{"a": 1}`, `{"b": 2}`}},
		{`[1,2], [3,4]`, []string{`[1,2]`, `[3,4]`}},
		{`"a,b", "c,d"`, []string{`"a,b"`, `"c,d"`}},
		{`a, "b,c", d`, []string{"a", `"b,c"`, "d"}},
		{`a, {"key": "val,ue"}, b`, []string{"a", `{"key": "val,ue"}`, "b"}},
		{`a, [1, 2, 3], b`, []string{"a", `[1, 2, 3]`, "b"}},
		{`a, "quoted", b`, []string{"a", `"quoted"`, "b"}},
		{`escaped\,comma`, []string{`escaped\,comma`}}, // splitByCommaIgnoringBraces handles CSV quotes, not backslash escape for comma itself unless inside quotes? No, the code checks for escape.
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, splitByCommaIgnoringBraces(tt.input))
		})
	}
}

func TestSplitByCommaIgnoringBraces_Escape(t *testing.T) {
	// The implementation has escape logic:
	/*
		if r == '\\' {
			escape = true
			current.WriteRune(r)
			continue
		}
	*/
	// It writes the backslash and the next char.
	// So `\,` -> `\,` in the output part.
	// But it prevents splitting.

	input := `a\,b,c`
	expected := []string{`a\,b`, "c"}
	assert.Equal(t, expected, splitByCommaIgnoringBraces(input))
}
