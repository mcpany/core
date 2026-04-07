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
			name:     "hash comment python",
			input:    "x = 1 # this is a comment\ny = 2",
			language: "python",
			expected: "x = 1 \ny = 2",
		},
		{
			name:     "slash comment go",
			input:    "x := 1 // this is a comment\ny := 2",
			language: "go",
			expected: "x := 1 \ny := 2",
		},
		{
			name:     "block comment nodejs",
			input:    "let x = 1; /* this is a comment */ let y = 2;",
			language: "node",
			expected: "let x = 1;  let y = 2;",
		},
		{
			name:     "hash inside quotes python",
			input:    "x = '1 # not a comment'\ny = 2",
			language: "python",
			expected: "x = '1 # not a comment'\ny = 2",
		},
		{
			name:     "slash inside double quotes go",
			input:    "x := \"1 // not a comment\"\ny := 2",
			language: "go",
			expected: "x := \"1 // not a comment\"\ny := 2",
		},
		{
			name:     "block comment inside backticks js",
			input:    "let x = `1 /* not a comment */`\nlet y = 2",
			language: "node",
			expected: "let x = `1 /* not a comment */`\nlet y = 2",
		},
		{
			name:     "escaped quote inside string",
			input:    "x = 'it\\'s # not a comment'\ny = 2",
			language: "python",
			expected: "x = 'it\\'s # not a comment'\ny = 2",
		},
		{
			name:     "backslash before comment python",
			input:    "x = 1 \\# this is a comment",
			language: "python",
			expected: "x = 1 \\# this is a comment",
		},
		{
			name:     "multiline block comment",
			input:    "x = 1; /*\nmulti\nline\n*/ y = 2;",
			language: "java",
			expected: "x = 1;  y = 2;",
		},
		{
			name:     "language default strict mode",
			input:    "x = 1 # hash\ny = 2 // slash\nz = 3 /* block */",
			language: "unknown",
			expected: "x = 1 \ny = 2 \nz = 3 ",
		},
		{
			name:     "php all types",
			input:    "$x = 1; # hash\n$y = 2; // slash\n$z = 3; /* block */",
			language: "php",
			expected: "$x = 1; \n$y = 2; \n$z = 3; ",
		},
		{
			name:     "node no hash support",
			input:    "x = 1 # not a comment in node\n// slash\n/* block */",
			language: "node",
			expected: "x = 1 # not a comment in node\n\n",
		},
		{
			name:     "bash no slash support",
			input:    "x=1 // not a comment in bash\n# hash\ny=2",
			language: "bash",
			expected: "x=1 // not a comment in bash\ny=2",
		},
		{
			name:     "block comment unclosed",
			input:    "x = 1; /* unclosed",
			language: "c",
			expected: "x = 1; ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := stripInterpreterComments(tt.input, tt.language)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestCheckUnquotedKeywords(t *testing.T) {
	keywords := []string{"rm", "exec", "eval"}
	tests := []struct {
		name     string
		input    string
		expectErr bool
	}{
		{
			name:      "no keywords",
			input:     "echo 'hello world'",
			expectErr: false,
		},
		{
			name:      "keyword unquoted",
			input:     "rm -rf /",
			expectErr: true,
		},
		{
			name:      "keyword in single quotes",
			input:     "echo 'rm -rf /'",
			expectErr: false,
		},
		{
			name:      "keyword in double quotes",
			input:     "echo \"rm -rf /\"",
			expectErr: false,
		},
		{
			name:      "keyword in backticks",
			input:     "echo `rm -rf /`",
			expectErr: false,
		},
		{
			name:      "keyword with escaped quote",
			input:     "echo \\\"rm\\\"",
			expectErr: true,
		},
		{
			name:      "keyword as substring",
			input:     "format=1", // 'rm' is in 'format', but it should only match whole words
			expectErr: false,
		},
		{
			name:      "keyword after space",
			input:     "  rm -rf /",
			expectErr: true,
		},
		{
			name:      "keyword after equals",
			input:     "CMD=rm",
			expectErr: true,
		},
		{
			name:      "keyword after pipe",
			input:     "cat file | rm",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkUnquotedKeywords(tt.input, keywords)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
