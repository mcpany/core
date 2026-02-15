package util

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactJSON_CommentsWithStructuralChars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string // We expect valid JSON
	}{
		{
			name: "comment with closing bracket inside array",
			input: `{
				"api_key": [
					1,
					/* ] */
					2
				]
			}`,
			expected: `{"api_key":"[REDACTED]"}`,
		},
		{
			name: "comment with closing brace inside object",
			input: `{
				"secret": {
					"a": 1,
					/* } */
					"b": 2
				}
			}`,
			expected: `{"secret":"[REDACTED]"}`,
		},
		{
			name: "comment with quote inside array",
			input: `{
				"token": [
					/* " */
				]
			}`,
			expected: `{"token":"[REDACTED]"}`,
		},
		{
			name: "comment with quote and bracket inside array",
			input: `{
				"password": [
					/* " ] */
				]
			}`,
			expected: `{"password":"[REDACTED]"}`,
		},
		{
			name: "line comment with closing bracket",
			input: `{
				"secret": [
					1,
					// ]
					2
				]
			}`,
			expected: `{"secret":"[REDACTED]"}`,
		},
		{
			name: "line comment with quote",
			input: `{
				"secret": [
					// "
				]
			}`,
			expected: `{"secret":"[REDACTED]"}`,
		},
		{
			name: "nested object with comment",
			input: `{
				"secret": {
					"nested": {
						/* } */
						"a": 1
					}
				}
			}`,
			expected: `{"secret":"[REDACTED]"}`,
		},
		{
			name: "string containing comment delimiters",
			input: `{
				"secret": [
					"/* comment start",
					"// line comment"
				]
			}`,
			expected: `{"secret":"[REDACTED]"}`,
		},
		{
			name: "escaped quote in string",
			input: `{
				"secret": [
					"\""
				]
			}`,
			expected: `{"secret":"[REDACTED]"}`,
		},
		{
			name: "escaped backslash before quote",
			input: `{
				"secret": [
					"\\"
				]
			}`,
			expected: `{"secret":"[REDACTED]"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RedactJSON([]byte(tt.input))

			// Compact expected JSON for comparison
			var gotJS interface{}
			if err := json.Unmarshal(got, &gotJS); err != nil {
				t.Fatalf("Redacted output is not valid JSON: %v\nOutput: %s", err, string(got))
			}
			gotBytes, _ := json.Marshal(gotJS)

			var expectedJS interface{}
			json.Unmarshal([]byte(tt.expected), &expectedJS)
			expectedBytes, _ := json.Marshal(expectedJS)

			assert.JSONEq(t, string(expectedBytes), string(gotBytes))
		})
	}
}

func TestRedactJSON_SkipStringCoverage(t *testing.T) {
	// Test skipString with tricky escapes to bump coverage in json_utils.go
	// skipString is internal, but we can exercise it via RedactJSON

	input := `{
		"key": "value \\\" with quote"
	}`
	// Not sensitive, should be kept as is (mostly, whitespace might change if we reparsed,
	// but RedactJSON preserves non-sensitive parts exactly).

	got := RedactJSON([]byte(input))
	assert.Equal(t, input, string(got))

	inputSensitive := `{
		"secret": "value \\\" with quote"
	}`
	// Should be redacted
	expected := `{"secret":"[REDACTED]"}`

	got = RedactJSON([]byte(inputSensitive))

	var gotJS interface{}
	json.Unmarshal(got, &gotJS)
	gotBytes, _ := json.Marshal(gotJS)

	var expectedJS interface{}
	json.Unmarshal([]byte(expected), &expectedJS)
	expectedBytes, _ := json.Marshal(expectedJS)

	assert.JSONEq(t, string(expectedBytes), string(gotBytes))
}

func TestRedactJSON_UnescapeKeyCoverage(t *testing.T) {
	// Test keys with escapes to cover unescapeKeySmall
	input := `{
		"sec\u0072et": "value"
	}`
	// "sec\u0072et" decodes to "secret"

	got := RedactJSON([]byte(input))

	// The output key might still be escaped or unescaped depending on implementation?
	// RedactJSON preserves the key as is in the output, just replaces the value.
	// So expected output should have the original key.

	expectedOutput := `{
		"sec\u0072et": "[REDACTED]"
	}`

	var gotJS interface{}
	json.Unmarshal(got, &gotJS)
	gotBytes, _ := json.Marshal(gotJS)

	var expectedJS interface{}
	json.Unmarshal([]byte(expectedOutput), &expectedJS)
	expectedBytes, _ := json.Marshal(expectedJS)

	assert.JSONEq(t, string(expectedBytes), string(gotBytes))
}

func TestRedactJSON_UnescapeKeyCoverage_Complex(t *testing.T) {
    // Test with more complex escapes
    input := `{
        "a\u0070\u0069_key": "val"
    }`
    // api_key
    expectedOutput := `{
        "a\u0070\u0069_key": "[REDACTED]"
    }`
     got := RedactJSON([]byte(input))

    var gotJS interface{}
    json.Unmarshal(got, &gotJS)
    gotBytes, _ := json.Marshal(gotJS)

    var expectedJS interface{}
    json.Unmarshal([]byte(expectedOutput), &expectedJS)
    expectedBytes, _ := json.Marshal(expectedJS)

    assert.JSONEq(t, string(expectedBytes), string(gotBytes))
}

func TestRedactJSON_SkipObjectCoverage(t *testing.T) {
	// Test cases to improve skipObject coverage
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "object with line comment containing brace",
			input: `{
				"secret": {
					// }
					"a": 1
				}
			}`,
			expected: `{"secret":"[REDACTED]"}`,
		},
		{
			name: "object with string containing braces",
			input: `{
				"secret": {
					"key": "{ }"
				}
			}`,
			expected: `{"secret":"[REDACTED]"}`,
		},
		{
			name: "object with deep nesting",
			input: `{
				"secret": {
					"a": { "b": { "c": 1 } }
				}
			}`,
			expected: `{"secret":"[REDACTED]"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RedactJSON([]byte(tt.input))
			var gotJS interface{}
			if err := json.Unmarshal(got, &gotJS); err != nil {
				t.Fatalf("Redacted output is not valid JSON: %v\nOutput: %s", err, string(got))
			}
			gotBytes, _ := json.Marshal(gotJS)

			var expectedJS interface{}
			json.Unmarshal([]byte(tt.expected), &expectedJS)
			expectedBytes, _ := json.Marshal(expectedJS)

			assert.JSONEq(t, string(expectedBytes), string(gotBytes))
		})
	}
}

func TestRedactJSON_UnescapeKeyMoreCoverage(t *testing.T) {
	// Test other escapes in keys
	// "se\ncret" -> secret? No.
	// We want to hit cases in unescapeKeySmall
	// \b \f \n \r \t

	// Create a key that looks like "secret" but with escapes that resolve to "secret"?
	// "secret" contains only normal chars.
	// But we want to test that unescapeKeySmall handles them.
	// If the key is sensitive AFTER unescaping, it redacts.
	// "secret" is sensitive.
	// "s\u0065cret" -> secret.

	// What if we use a key that is NOT sensitive but has escapes?
	// It should just NOT be redacted.
	input2 := `{
		"public\tdata": "val"
	}`
	got := RedactJSON([]byte(input2))
	assert.Contains(t, string(got), `"val"`)

	// Test failure cases for unescape (e.g. truncated) - hard to trigger via RedactJSON as input must be valid JSON to be parsed?
	// But RedactJSON works on byte slice, scanning.
	// If we provide malformed JSON key: `"key\` -> unescape might fail.
}

func TestRedactJSON_CommentWithQuoteAfterNumber(t *testing.T) {
	// This test case demonstrates a fix for a bug where a comment immediately following a number
	// causes the scanner to misinterpret the comment content as a string.
	// We verify that "password" is redacted.
	input := "{\"count\": 123// \" } \n, \"password\": \"leak\"}"
	got := RedactJSON([]byte(input))
	assert.Contains(t, string(got), "\"[REDACTED]\"")
	assert.NotContains(t, string(got), "\"leak\"")
}
