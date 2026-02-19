// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"
)

func TestInterpreterFunctionCalls_BypassAttempts(t *testing.T) {
	cases := []struct {
		name        string
		val         string
		language    string
		shouldBlock bool
	}{
		// Obfuscation Bypasses (Should be BLOCKED)
		{"Perl Comment Bypass", "system # \n ('ls')", "perl", true},
		{"Python Comment Bypass", "system # \n ('ls')", "python", true},
		{"Line Continuation Bypass", "system \\\n ('ls')", "python", true},
		{"Slash Bypass", "system \\ ('ls')", "python", true},
		{"C-Style Line Comment Bypass (Node)", "system // \n ('ls')", "node", true},
		{"C-Style Block Comment Bypass (Node)", "system /* comment */ ('ls')", "node", true},
		{"Multi-line Block Comment Bypass", "system /* \n comment \n */ ('ls')", "node", true},
		{"Mixed Comments (Node)", "system /* block */ // line \n ('ls')", "node", true},

		// String Literal Awareness (Should be BLOCKED because system call is REAL)
		// This ensures we don't accidentally strip code that looks like a comment but is inside a string,
		// which would hide the subsequent malicious code if we were just stripping blindly?
		// Wait, if "x = '#'; system('ls')"
		// Naive strip: "x = '" -> passes. Exploit!
		// Correct strip: "x = '#'; system('ls')" -> blocks. Secure.
		{"String Literal Blindness (Python #)", "x = \"#\"; system('ls')", "python", true},
		{"String Literal Blindness (JS //)", "x = \"//\"; system('ls')", "node", true},

		// False Positives (Should NOT be BLOCKED)
		// System call inside a comment should be ignored.
		{"System call in comment (Python)", "# system('ls')", "python", false},
		{"System call in comment (Node)", "// system('ls')", "node", false},
		{"System call in block comment (Node)", "/* system('ls') */", "node", false},

		// Valid code
		{"Valid Python", "print('hello')", "python", false},
		{"Valid Python with div", "x = 10 / 2", "python", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := checkInterpreterFunctionCalls(tc.val, tc.language, false)
			if tc.shouldBlock {
				if err == nil {
					t.Errorf("SECURITY BYPASS: content %q (lang: %s) was NOT blocked", tc.val, tc.language)
				} else {
					t.Logf("Blocked as expected: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("FALSE POSITIVE: content %q (lang: %s) was blocked: %v", tc.val, tc.language, err)
				}
			}
		})
	}
}
