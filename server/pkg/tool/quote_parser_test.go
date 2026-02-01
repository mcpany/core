// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnalyzeQuoteContext_ShellInjectionBypass(t *testing.T) {
	// Vulnerability: analyzeQuoteContext incorrectly handles backslash escaping inside single quotes.
	// In sh/bash, single quotes are strong quotes and backslash is a literal.
	//
	// Template: echo 'foo\' {{val}}'
	//
	// Parser (Vulnerable):
	// ' -> inSingle=true
	// \ -> escaped=true
	// ' -> skipped (escaped) -> inSingle STAYS true
	// {{val}} -> seen as Single Quoted (Level 2)
	//
	// Shell (Real):
	// 'foo\' -> string containing "foo\"
	// {{val}} -> UNQUOTED
	// ' -> starts new string
	//
	// If the parser thinks it is Level 2, it allows strict injection characters like $ and ().
	// If the shell sees it as Unquoted, it executes them.

	template := "echo 'foo\\' {{val}}'"
	placeholder := "{{val}}"

	level := analyzeQuoteContext(template, placeholder)

	// We expect Level 0 (Unquoted) because the first single quote block ends at the second single quote.
	// The backslash is just a character inside the first block.
	assert.Equal(t, 0, level, "Expected Unquoted (0) context for %q", template)
}

func TestAnalyzeQuoteContext_StandardCases(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		placeholder string
		wantLevel   int
	}{
		{
			name:        "Unquoted simple",
			template:    "echo {{val}}",
			placeholder: "{{val}}",
			wantLevel:   0,
		},
		{
			name:        "Double quoted",
			template:    "echo \"{{val}}\"",
			placeholder: "{{val}}",
			wantLevel:   1,
		},
		{
			name:        "Single quoted",
			template:    "echo '{{val}}'",
			placeholder: "{{val}}",
			wantLevel:   2,
		},
		{
			name:        "Backtick quoted",
			template:    "echo `{{val}}`",
			placeholder: "{{val}}",
			wantLevel:   3,
		},
		{
			name:        "Escaped double quote",
			template:    "echo \"foo\\\" {{val}}\"",
			placeholder: "{{val}}",
			wantLevel:   1,
		},
		{
			name:        "Escaped backslash then quote",
			template:    "echo \"foo\\\\\" {{val}}",
			placeholder: "{{val}}",
			wantLevel:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzeQuoteContext(tt.template, tt.placeholder)
			assert.Equal(t, tt.wantLevel, got)
		})
	}
}
