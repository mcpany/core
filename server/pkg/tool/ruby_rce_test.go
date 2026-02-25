// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"
)

func TestRubyRCEBypass(t *testing.T) {
	// Target: Ruby Code Execution via Kernel.method("system").call(...)
	// Constraint: 'system' keyword is blocked as a word, but string "system" is allowed inside quotes.

	val := `Kernel.method("system").call("ls")`

	// We simulate a ruby -e "{{code}}" context, so quoteLevel = 1 (Double Quoted)
	// But checkInterpreterFunctionCalls strips quotes from the value itself? No.
	// The value IS the code inside the quotes.

	// checkForShellInjection calls checkInterpreterFunctionCalls(val, "ruby")

	err := checkInterpreterFunctionCalls(val, "ruby")
	if err != nil {
		t.Logf("Blocked Ruby RCE: %v", err)
	} else {
		t.Errorf("Bypass: Ruby RCE accepted: %s", val)
	}
}
