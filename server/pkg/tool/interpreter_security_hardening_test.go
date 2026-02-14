package tool

import (
	"testing"
)

func TestInterpreterSecurityHardening(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		shouldBlock bool
	}{
		// Confirmed Bypasses
		{
			name:        "Bypass: System with hash comment",
			input:       "system # \n ('ls')",
			shouldBlock: true,
		},
		{
			name:        "Bypass: System with newline",
			input:       "system \n ('ls')",
			shouldBlock: true,
		},
		{
			name:        "Bypass: System with tab",
			input:       "system\t('ls')",
			shouldBlock: true,
		},
		{
			name:        "Bypass: Exec with double slash comment",
			input:       "exec // comment \n ('ls')",
			shouldBlock: true,
		},
		{
			name:        "Bypass: Import with block comment",
			input:       "import /* comment */ os",
			shouldBlock: true,
		},
		{
			name:        "Bypass: Python mixed quotes",
			input:       "import # \n os",
			shouldBlock: true,
		},
		{
			name:        "Bypass: Eval with space",
			input:       "eval ('ls')",
			shouldBlock: true,
		},

		// Confirmed False Positives (Should be allowed)
		{
			name:        "False Positive: Print string containing keyword",
			input:       "print \"system\"",
			shouldBlock: false,
		},
		{
			name:        "False Positive: Print single quoted keyword",
			input:       "print 'exec'",
			shouldBlock: false,
		},
		{
			name:        "Block: Print backticked keyword (treated as code)",
			input:       "print `eval`",
			shouldBlock: true,
			// Backticks are now treated as code by sanitizeInterpreterCode to prevent RCE in Perl/Ruby/PHP
			// and Template Injection in Node.js.
			// This means keywords inside backticks are detected and blocked.
		},
		{
			name:        "False Positive: Variable name containing keyword",
			input:       "my_system = 1",
			shouldBlock: false,
		},
		{
			name:        "False Positive: Variable name ending with keyword",
			input:       "filesystem = 1",
			shouldBlock: false,
		},
		{
			name:        "False Positive: Variable name starting with keyword",
			input:       "system_id = 1",
			shouldBlock: false,
		},

		// Standard Blocking
		{
			name:        "Block: Simple system call",
			input:       "system('ls')",
			shouldBlock: true,
		},
		{
			name:        "Block: Simple exec call",
			input:       "exec 'ls'",
			shouldBlock: true,
		},
		{
			name:        "Block: Import os",
			input:       "import os",
			shouldBlock: true,
		},
		{
			name:        "Block: Popen",
			input:       "popen('ls')",
			shouldBlock: true,
		},
		{
			name:        "Block: Spawn",
			input:       "spawn('ls')",
			shouldBlock: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := checkInterpreterFunctionCalls(tc.input)
			if tc.shouldBlock && err == nil {
				t.Errorf("Expected block for input: %q, but got nil", tc.input)
			}
			if !tc.shouldBlock && err != nil {
				t.Errorf("Expected allow for input: %q, but got error: %v", tc.input, err)
			}
		})
	}
}
