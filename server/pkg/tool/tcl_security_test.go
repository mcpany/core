// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"
)

func TestTclInjection(t *testing.T) {
	// Tcl (tclsh) allows "exec ls" without parentheses.
	// Currently checkInterpreterFunctionCalls treats tclsh as non-strict,
	// so "exec" is only blocked if followed by '(', '=', ':'.

	tests := []struct {
		name     string
		val      string
		command  string
		expected bool // true if error expected (blocked), false otherwise (allowed)
	}{
		{
			name:     "Direct exec call",
			val:      "exec ls",
			command:  "tclsh",
			expected: true,
		},
		{
			name:     "Eval exec call",
			val:      "eval exec ls",
			command:  "tclsh",
			expected: true,
		},
		{
			name:     "Open pipe",
			val:      "open \"|ls\"",
			command:  "tclsh",
			expected: true,
		},
		{
			name:     "Spawn process",
			val:      "spawn ls",
			command:  "expect", // expect is also tcl-based
			expected: true,
		},
		{
			name:     "Valid input",
			val:      "puts hello",
			command:  "tclsh",
			expected: false,
		},
		{
			name:     "Safe use of exec as part of word",
			val:      "my_exec_status",
			command:  "tclsh",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := "{{script}}"
			placeholder := "{{script}}"
			// Use real isShell logic to match LocalCommandTool behavior
			isShellVal := isShell(tt.command)

			err := checkForShellInjection(tt.val, template, placeholder, tt.command, isShellVal)
			if tt.expected {
				if err == nil {
					t.Errorf("Expected error (blocked) for input %q, got nil", tt.val)
				}
			} else {
				if err != nil {
					t.Errorf("Expected nil (allowed) for input %q, got error: %v", tt.val, err)
				}
			}
		})
	}
}
