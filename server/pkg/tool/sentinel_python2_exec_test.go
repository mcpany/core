// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"
)

func TestPython2Exec(t *testing.T) {
	// Vulnerability: Python 2 allowed `exec "print(1)"` without parentheses.
	// Our fix should block `exec` followed by space or quote.

	val := `exec "print(1)"`

	// Python is non-strict.
	// exec is in functionKeywords.

	// We use a template that puts it in quoteLevel 2 (Single) to trigger checkInterpreterFunctionCalls
	// but skip checkUnquotedKeywords (which is skipped for Level 2).
	// Although checkUnquotedKeywords is ALSO run inside checkInterpreterFunctionCalls for STRICT languages.
	// But Python is NOT strict.

	template := "'{{val}}'"

	err := checkForShellInjection(val, template, "{{val}}", "python", false)

	if err == nil {
		t.Fatalf("Vulnerability confirmed: 'exec \"print(1)\"' passed validation in Single Quote context (Python). Error was nil.")
	} else {
        t.Logf("Blocked with error: %v", err)
    }
}
