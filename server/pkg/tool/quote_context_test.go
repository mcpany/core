// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnalyzeQuoteContext(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		placeholder string
		wantLevel   int
	}{
		{
			name:        "Empty template",
			template:    "",
			placeholder: "{{val}}",
			wantLevel:   0,
		},
		{
			name:        "Unquoted simple",
			template:    "echo {{val}}",
			placeholder: "{{val}}",
			wantLevel:   0,
		},
		{
			name:        "Double quoted simple",
			template:    `echo "{{val}}"`,
			placeholder: "{{val}}",
			wantLevel:   1,
		},
		{
			name:        "Single quoted simple",
			template:    `echo '{{val}}'`,
			placeholder: "{{val}}",
			wantLevel:   2,
		},
		{
			name:        "Backticked simple",
			template:    "echo `{{val}}`",
			placeholder: "{{val}}",
			wantLevel:   3,
		},
		{
			name:        "Mixed quotes - double inside single",
			template:    `echo '"{{val}}"'`,
			placeholder: "{{val}}",
			wantLevel:   2,
		},
		{
			name:        "Mixed quotes - single inside double",
			template:    `echo "'{{val}}'"`,
			placeholder: "{{val}}",
			wantLevel:   1,
		},
		{
			name:        "Escaped double quote inside double quotes",
			template:    `echo "foo \" {{val}} \" bar"`,
			placeholder: "{{val}}",
			wantLevel:   1,
		},
		{
			name:        "Escaped single quote inside single quotes",
			// Note: single quotes in shell usually don't support backslash escaping like this,
			// but some interpreters or contexts might.
			// analyzeQuoteContext treats backslash as escape char generally unless in single quotes?
			// Let's check implementation:
			// if char == '\\' && !inSingle { escaped = true; continue }
			// So in single quotes, backslash is NOT an escape char.
			// 'foo \' is 'foo \' (literal backslash and end quote?) No.
			// 'foo \' bar' -> the single quote closes.
			template:    `echo 'foo \' {{val}}`,
			placeholder: "{{val}}",
			// 'foo \ ' -> closed at second '
			// {{val}} is outside.
			wantLevel:   0,
		},
		{
			name:        "Escaped double quote outside",
			template:    `echo \"{{val}}\"`,
			placeholder: "{{val}}",
			wantLevel:   0,
		},
		{
			name:        "Backslash before placeholder",
			template:    `echo \{{val}}`,
			placeholder: "{{val}}",
			wantLevel:   0,
		},
		{
			name:        "Multiple placeholders - take minimum security (unquoted)",
			template:    `echo "{{val}}" && echo {{val}}`,
			placeholder: "{{val}}",
			wantLevel:   0,
		},
		{
			name:        "Multiple placeholders - take minimum security (double)",
			template:    `echo '{{val}}' && echo "{{val}}"`,
			placeholder: "{{val}}",
			wantLevel:   1,
		},
		{
			name:        "Partial quote - unclosed double",
			template:    `echo "{{val}}`,
			placeholder: "{{val}}",
			wantLevel:   1,
		},
		{
			name:        "Partial quote - unclosed single",
			template:    `echo '{{val}}`,
			placeholder: "{{val}}",
			wantLevel:   2,
		},
		{
			name:        "Complex nesting",
			template:    `sh -c "echo '{{val}}'"`,
			placeholder: "{{val}}",
			wantLevel:   1, // Inside double quotes (outermost significant context for interpretation)
		},
		{
			name:        "Complex nesting reversed",
			template:    `sh -c 'echo "{{val}}"'`,
			placeholder: "{{val}}",
			wantLevel:   2, // Inside single quotes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzeQuoteContext(tt.template, tt.placeholder)
			assert.Equal(t, tt.wantLevel, got)
		})
	}
}

func TestCheckForShellInjection_Comprehensive(t *testing.T) {
	tests := []struct {
		name        string
		val         string
		template    string
		placeholder string
		command     string
		wantErr     bool
		errContains string
	}{
		// Unquoted Context (Level 0)
		{
			name:        "Unquoted safe",
			val:         "safe_value",
			template:    "echo {{val}}",
			placeholder: "{{val}}",
			command:     "bash",
			wantErr:     false,
		},
		{
			name:        "Unquoted injection - semicolon",
			val:         "val; rm -rf /",
			template:    "echo {{val}}",
			placeholder: "{{val}}",
			command:     "bash",
			wantErr:     true,
			errContains: ";",
		},
		{
			name:        "Unquoted injection - space (new argument)",
			val:         "val argument2",
			template:    "echo {{val}}",
			placeholder: "{{val}}",
			command:     "bash",
			wantErr:     true,
			errContains: " ",
		},
		{
			name:        "Unquoted injection - pipe",
			val:         "val | ls",
			template:    "echo {{val}}",
			placeholder: "{{val}}",
			command:     "bash",
			wantErr:     true,
			errContains: "shell injection detected", // Found space first
		},
		{
			name:        "Unquoted injection - backtick",
			val:         "val `ls`",
			template:    "echo {{val}}",
			placeholder: "{{val}}",
			command:     "bash",
			wantErr:     true,
			errContains: "shell injection detected", // Found space first
		},

		// Double Quoted Context (Level 1)
		{
			name:        "Double quoted safe",
			val:         "safe value with spaces",
			template:    `echo "{{val}}"`,
			placeholder: "{{val}}",
			command:     "bash",
			wantErr:     false,
		},
		{
			name:        "Double quoted injection - quote breakout",
			val:         `val" ; rm -rf / ; "`,
			template:    `echo "{{val}}"`,
			placeholder: "{{val}}",
			command:     "bash",
			wantErr:     true,
			errContains: `"`,
		},
		{
			name:        "Double quoted injection - variable expansion",
			val:         "$HOME",
			template:    `echo "{{val}}"`,
			placeholder: "{{val}}",
			command:     "bash",
			wantErr:     true,
			errContains: "$",
		},
		{
			name:        "Double quoted injection - backtick",
			val:         "`ls`",
			template:    `echo "{{val}}"`,
			placeholder: "{{val}}",
			command:     "bash",
			wantErr:     true,
			errContains: "`",
		},
		{
			name:        "Double quoted injection - backslash",
			val:         `val\`,
			template:    `echo "{{val}}"`,
			placeholder: "{{val}}",
			command:     "bash",
			wantErr:     true,
			errContains: `\`,
		},

		// Single Quoted Context (Level 2)
		{
			name:        "Single quoted safe",
			val:         `safe value with $ and \ and "`,
			template:    `echo '{{val}}'`,
			placeholder: "{{val}}",
			command:     "bash",
			wantErr:     false,
		},
		{
			name:        "Single quoted injection - quote breakout",
			val:         `val' ; rm -rf / ; '`,
			template:    `echo '{{val}}'`,
			placeholder: "{{val}}",
			command:     "bash",
			wantErr:     true,
			errContains: "value contains single quote",
		},

		// Backticked Context (Level 3)
		{
			name:        "Backticked injection - quote breakout",
			val:         "val` ; rm -rf / ; `",
			template:    "echo `{{val}}`",
			placeholder: "{{val}}",
			command:     "bash",
			wantErr:     true,
			errContains: "`",
		},

		// Python Injection
		{
			name:        "Python safe",
			val:         "print('hello')",
			template:    `python -c "{{val}}"`,
			placeholder: "{{val}}",
			command:     "python",
			wantErr:     false,
		},
		{
			name:        "Python f-string injection",
			val:         "{__import__('os').system('ls')}",
			template:    `python -c "f'{{val}}'"`,
			placeholder: "{{val}}",
			command:     "python",
			wantErr:     true,
			errContains: "python f-string injection",
		},

		// Ruby Injection
		{
			name:        "Ruby interpolation",
			val:         "#{system('ls')}",
			template:    `ruby -e "{{val}}"`,
			placeholder: "{{val}}",
			command:     "ruby",
			wantErr:     true,
			errContains: "ruby interpolation",
		},
		{
			name:        "Ruby open pipe",
			val:         "|ls",
			template:    `ruby -e 'open("{{val}}")'`,
			placeholder: "{{val}}",
			command:     "ruby",
			wantErr:     true,
			errContains: "ruby open injection",
		},

		// Perl/PHP Injection
		{
			name:        "Perl qx injection",
			val:         "qx/ls/",
			template:    `perl -e '{{val}}'`,
			placeholder: "{{val}}",
			command:     "perl",
			wantErr:     true,
			errContains: "perl qx execution",
		},

		// Dangerous Environment Variables (checked in checkUnquotedInjection if unquoted)
		// But verify via checkEnvInjection logic indirectly if needed?
		// checkEnvInjection is separate.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkForShellInjection(tt.val, tt.template, tt.placeholder, tt.command)
			if tt.wantErr {
				if assert.Error(t, err) {
					if tt.errContains != "" {
						assert.Contains(t, err.Error(), tt.errContains)
					}
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckInterpreterFunctionCalls(t *testing.T) {
	tests := []struct {
		val         string
		hasError    bool
		description string
	}{
		{"system('ls')", true, "system call"},
		{"exec ('ls')", true, "exec call with space"},
		{"popen\t('ls')", true, "popen call with tab"},
		{"eval('ls')", true, "eval call"},
		{"system", false, "system word safe"},
		{"system_call", false, "system_call safe"},
		// mysystem('ls') triggers the check because it contains "system(".
		// The current implementation strips spaces and does not check word boundaries.
		// This is a known false positive.
		{"mysystem('ls')", true, "mysystem safe (known false positive)"},
		{"__import__('os')", true, "__import__ check"},
		{"require 'fs'", true, "require call (ruby/node style)"},
		{"import 'os'", true, "import call"},
		{"spawn('ls')", true, "spawn call"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			err := checkInterpreterFunctionCalls(tt.val)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckForArgumentInjection_Comprehensive(t *testing.T) {
	tests := []struct {
		val      string
		wantErr  bool
	}{
		{"safe", false},
		{"-unsafe", true},
		{"--unsafe", true},
		{"-123", false}, // Negative number allowed
		{"-1.23", false}, // Negative float allowed
		{"-", true}, // Just dash
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Input: %s", tt.val), func(t *testing.T) {
			err := checkForArgumentInjection(tt.val)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
