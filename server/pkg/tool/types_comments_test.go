// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripInterpreterComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		language string
		expected string
	}{
		{
			name:     "No comments",
			input:    "print('hello')",
			language: "python",
			expected: "print('hello')",
		},
		{
			name:     "Python hash comment",
			input:    "print('hello') # comment",
			language: "python",
			expected: "print('hello') ",
		},
		{
			name:     "Python hash comment inside string (single quotes)",
			input:    "s = '# not a comment'",
			language: "python",
			expected: "s = '# not a comment'",
		},
		{
			name:     "Python hash comment inside string (double quotes)",
			input:    `s = "# not a comment"`,
			language: "python",
			expected: `s = "# not a comment"`,
		},
		{
			name:     "Escaped quote in string",
			input:    `s = "\"# still not a comment"`,
			language: "python",
			expected: `s = "\"# still not a comment"`,
		},
		{
			name:     "Escaped backslash before quote",
			input:    `s = "\\"; # comment`,
			language: "python",
			expected: `s = "\\"; `,
		},
		{
			name:     "JavaScript line comment",
			input:    "console.log('hello'); // comment",
			language: "node",
			expected: "console.log('hello'); ",
		},
		{
			name:     "JavaScript block comment",
			input:    "console.log('hello'); /* comment */ var x = 1;",
			language: "node",
			expected: "console.log('hello');  var x = 1;",
		},
		{
			name:     "JavaScript comment inside string",
			input:    `s = "// not a comment";`,
			language: "node",
			expected: `s = "// not a comment";`,
		},
		{
			name:     "JavaScript comment inside template literal",
			input:    "s = `// not a comment`;",
			language: "node",
			expected: "s = `// not a comment`;",
		},
		{
			name:     "Mixed quotes",
			input:    `s = '"# not a comment"';`,
			language: "python",
			expected: `s = '"# not a comment"';`,
		},
		{
			name:     "PHP mixed comments",
			input:    "$x = 1; # hash comment\n$y = 2; // line comment\n/* block comment */",
			language: "php",
			expected: "$x = 1; \n$y = 2; \n",
		},
        {
            name:     "Edge case: escaped backslash inside string followed by quote",
            input:    `s = "\\"; # comment`,
            language: "python",
            expected: `s = "\\"; `,
        },
        {
            name:     "Edge case: quote inside comment",
            input:    `x = 1 # comment with "quote"`,
            language: "python",
            expected: `x = 1 `,
        },
        {
            name:     "Edge case: comment marker at start of string",
            input:    `s = "# start"`,
            language: "python",
            expected: `s = "# start"`,
        },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripInterpreterComments(tt.input, tt.language)
			assert.Equal(t, tt.expected, result)
		})
	}
}
