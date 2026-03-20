// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"
)

func TestRubySystemSpace(t *testing.T) {
	// Vulnerability: checkInterpreterFunctionCalls allows 'system "id"' (no parens)
	// This is dangerous if the input is used in an eval context or similar.

	val := `system "id"`
	template := "'{{val}}'"

	// QuoteLevel 2 (Single Quoted).

	err := checkForShellInjection(val, template, "{{val}}", "ruby", false)

	if err == nil {
		t.Fatalf("Vulnerability confirmed: 'system \"id\"' passed validation in Single Quote context (Ruby). Error was nil.")
	} else {
        t.Logf("Blocked with error: %v", err)
    }

	valPerl := `system "id"`
	errPerl := checkForShellInjection(valPerl, template, "{{val}}", "perl", false)
	if errPerl == nil {
		t.Fatalf("Vulnerability confirmed: 'system \"id\"' passed validation in Single Quote context (Perl). Error was nil.")
	} else {
        t.Logf("Blocked with error: %v", errPerl)
    }
}
