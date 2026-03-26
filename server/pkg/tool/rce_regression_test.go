// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"
)

// TestRegression_RCE_ShellInjection verifies that nested command substitution attempts
// inside double-quoted templates are correctly identified as strict/unquoted context,
// thus preventing RCE payloads like `'); ls #`.
func TestRegression_RCE_ShellInjection(t *testing.T) {
	// Payload that attempts to break out of single quotes inside $()
	// Template: echo "$(echo '{{input}}')"
	// This template uses Double Quotes, but contains $() which nests quotes.
	// If context is misidentified as Double Quotes (Level 1), this payload is accepted.
	// If context is correctly identified as Complex/Strict (Level 0), this payload is rejected
	// because it contains `'` and `)` and `;`.
	val := "'); ls #"
	template := `echo "$(echo '{{input}}')"`
	placeholder := "{{input}}"
	command := "bash"
	isShell := true

	// Call the internal function checkForShellInjection
	// Note: We are in package tool, so we can access unexported functions.
	err := checkForShellInjection(val, template, placeholder, command, isShell)

	if err == nil {
		t.Errorf("VULNERABILITY DETECTED: Payload %q was accepted for template %q. Expected rejection due to shell injection.", val, template)
	} else {
		t.Logf("SUCCESS: Payload correctly rejected: %v", err)
	}

	// Also test with backticks inside double quotes
	templateBacktick := "echo \"`echo '{{input}}'`\""
	errBacktick := checkForShellInjection(val, templateBacktick, placeholder, command, isShell)
	if errBacktick == nil {
		t.Errorf("VULNERABILITY DETECTED: Payload %q was accepted for template %q. Expected rejection.", val, templateBacktick)
	} else {
		t.Logf("SUCCESS: Payload correctly rejected (backtick): %v", errBacktick)
	}

	// Test with ${...} variable expansion inside double quotes
	// echo "${var:-{{input}}}"
	templateVar := `echo "${var:-{{input}}}"`
	errVar := checkForShellInjection(val, templateVar, placeholder, command, isShell)
	if errVar == nil {
		t.Errorf("VULNERABILITY DETECTED: Payload %q was accepted for template %q. Expected rejection.", val, templateVar)
	} else {
		t.Logf("SUCCESS: Payload correctly rejected (variable expansion): %v", errVar)
	}
}
