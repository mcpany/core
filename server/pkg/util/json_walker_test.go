package util

import (
	"testing"
)

func TestWalkJSONStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		visitor  func(raw []byte) ([]byte, bool)
		expected string
	}{
		{
			name:  "no changes",
			input: `{"key": "value"}`,
			visitor: func(raw []byte) ([]byte, bool) {
				return nil, false
			},
			expected: `{"key": "value"}`,
		},
		{
			name:  "replace value",
			input: `{"key": "value"}`,
			visitor: func(raw []byte) ([]byte, bool) {
				if string(raw) == `"value"` {
					return []byte(`"REDACTED"`), true
				}
				return nil, false
			},
			expected: `{"key": "REDACTED"}`,
		},
		{
			name:  "ignore keys",
			input: `{"target": "value"}`,
			visitor: func(raw []byte) ([]byte, bool) {
				if string(raw) == `"target"` {
					return []byte(`"FAIL"`), true
				}
				return nil, false
			},
			expected: `{"target": "value"}`,
		},
		{
			name:  "nested object",
			input: `{"a": {"b": "target"}}`,
			visitor: func(raw []byte) ([]byte, bool) {
				if string(raw) == `"target"` {
					return []byte(`"HIT"`), true
				}
				return nil, false
			},
			expected: `{"a": {"b": "HIT"}}`,
		},
		{
			name:  "array",
			input: `["a", "b"]`,
			visitor: func(raw []byte) ([]byte, bool) {
				return []byte(`"X"`), true
			},
			expected: `["X", "X"]`,
		},
		{
			name:  "mixed array",
			input: `["a", {"k": "v"}, "b"]`,
			visitor: func(raw []byte) ([]byte, bool) {
				if string(raw) == `"v"` {
					return []byte(`"V"`), true
				}
				return nil, false
			},
			expected: `["a", {"k": "V"}, "b"]`,
		},
		{
			name:  "escaped quotes",
			input: `{"key": "val\"ue"}`,
			visitor: func(raw []byte) ([]byte, bool) {
				if string(raw) == `"val\"ue"` {
					return []byte(`"HIT"`), true
				}
				return nil, false
			},
			expected: `{"key": "HIT"}`,
		},
		{
			name:  "key with block comment",
			input: `{"key" /* comment */ : "value"}`,
			visitor: func(raw []byte) ([]byte, bool) {
				return []byte(`"REPLACED"`), true
			},
			expected: `{"key" /* comment */ : "REPLACED"}`,
		},
		{
			name: "key with line comment",
			input: `{"key" // comment
: "value"}`,
			visitor: func(raw []byte) ([]byte, bool) {
				return []byte(`"REPLACED"`), true
			},
			expected: `{"key" // comment
: "REPLACED"}`,
		},
		{
			name: "value with comment",
			input: `{"key": "value" // comment
}`,
			visitor: func(raw []byte) ([]byte, bool) {
				return []byte(`"REPLACED"`), true
			},
			expected: `{"key": "REPLACED" // comment
}`,
		},
		{
			name:  "string in block comment",
			input: `/* "hidden" */ "visible"`,
			visitor: func(raw []byte) ([]byte, bool) {
				return []byte(`"REPLACED"`), true
			},
			expected: `/* "hidden" */ "REPLACED"`,
		},
		{
			name: "string in line comment",
			input: `// "hidden"
"visible"`,
			visitor: func(raw []byte) ([]byte, bool) {
				return []byte(`"REPLACED"`), true
			},
			expected: `// "hidden"
"REPLACED"`,
		},
		{
			name:  "string in block comment before key",
			input: `{ /* "ignore" */ "key": "value" }`,
			visitor: func(raw []byte) ([]byte, bool) {
				return []byte(`"REPLACED"`), true
			},
			expected: `{ /* "ignore" */ "key": "REPLACED" }`,
		},
		{
			name:  "slash before line comment with quote",
			input: `{"key": 1 / 2 // "commented"
, "k2": "v2"}`,
			visitor: func(raw []byte) ([]byte, bool) {
				if string(raw) == `"commented"` {
					return []byte(`"REDACTED"`), true
				}
				if string(raw) == `"v2"` {
					return []byte(`"V2"`), true
				}
				return nil, false
			},
			expected: `{"key": 1 / 2 // "commented"
, "k2": "V2"}`,
		},
		{
			name:  "slash before block comment with quote",
			input: `{"key": 1 / 2 /* "commented" */, "k2": "v2"}`,
			visitor: func(raw []byte) ([]byte, bool) {
				if string(raw) == `"commented"` {
					return []byte(`"REDACTED"`), true
				}
				if string(raw) == `"v2"` {
					return []byte(`"V2"`), true
				}
				return nil, false
			},
			expected: `{"key": 1 / 2 /* "commented" */, "k2": "V2"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WalkJSONStrings([]byte(tt.input), tt.visitor)
			if string(result) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(result))
			}
			// check zero allocation if no change
			if tt.input == tt.expected {
				if &result[0] != &([]byte(tt.input))[0] {
					// This check is tricky because []byte(str) allocates.
					// But we want to ensure WalkJSONStrings returns the input slice if no change.
					// We can check equality of pointer if we pass a slice variable.
				}
			}
		})
	}
}

func TestWalkJSONStrings_ZeroAlloc(t *testing.T) {
	input := []byte(`{"key": "value"}`)
	visitor := func(raw []byte) ([]byte, bool) {
		return nil, false
	}
	result := WalkJSONStrings(input, visitor)
	if &input[0] != &result[0] {
		t.Error("WalkJSONStrings allocated new buffer even when no changes were made")
	}
}

func TestWalkStandardJSONStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		visitor  func(raw []byte) ([]byte, bool)
		expected string
	}{
		{
			name:  "no changes",
			input: `{"key": "value"}`,
			visitor: func(raw []byte) ([]byte, bool) {
				return nil, false
			},
			expected: `{"key": "value"}`,
		},
		{
			name:  "replace value",
			input: `{"key": "value"}`,
			visitor: func(raw []byte) ([]byte, bool) {
				if string(raw) == `"value"` {
					return []byte(`"REDACTED"`), true
				}
				return nil, false
			},
			expected: `{"key": "REDACTED"}`,
		},
		{
			name:  "ignore keys",
			input: `{"target": "value"}`,
			visitor: func(raw []byte) ([]byte, bool) {
				if string(raw) == `"target"` {
					return []byte(`"FAIL"`), true
				}
				return nil, false
			},
			expected: `{"target": "value"}`,
		},
		{
			name:  "nested object",
			input: `{"a": {"b": "target"}}`,
			visitor: func(raw []byte) ([]byte, bool) {
				if string(raw) == `"target"` {
					return []byte(`"HIT"`), true
				}
				return nil, false
			},
			expected: `{"a": {"b": "HIT"}}`,
		},
		{
			name:  "array",
			input: `["a", "b"]`,
			visitor: func(raw []byte) ([]byte, bool) {
				return []byte(`"X"`), true
			},
			expected: `["X", "X"]`,
		},
		{
			name:  "mixed array",
			input: `["a", {"k": "v"}, "b"]`,
			visitor: func(raw []byte) ([]byte, bool) {
				if string(raw) == `"v"` {
					return []byte(`"V"`), true
				}
				return nil, false
			},
			expected: `["a", {"k": "V"}, "b"]`,
		},
		{
			name:  "escaped quotes",
			input: `{"key": "val\"ue"}`,
			visitor: func(raw []byte) ([]byte, bool) {
				if string(raw) == `"val\"ue"` {
					return []byte(`"HIT"`), true
				}
				return nil, false
			},
			expected: `{"key": "HIT"}`,
		},
		{
			name:  "url with slashes",
			input: `{"url": "http://example.com/path"}`,
			visitor: func(raw []byte) ([]byte, bool) {
				if string(raw) == `"http://example.com/path"` {
					return []byte(`"REDACTED"`), true
				}
				return nil, false
			},
			expected: `{"url": "REDACTED"}`,
		},
		{
			name:  "escaped slash",
			input: `{"key": "val\\/ue"}`,
			visitor: func(raw []byte) ([]byte, bool) {
				if string(raw) == `"val\\/ue"` {
					return []byte(`"HIT"`), true
				}
				return nil, false
			},
			expected: `{"key": "HIT"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WalkStandardJSONStrings([]byte(tt.input), tt.visitor)
			if string(result) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(result))
			}
		})
	}
}
