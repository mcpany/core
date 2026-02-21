package tool

import (
	"testing"
)

func TestTclInjection(t *testing.T) {
	// Tcl (tclsh) allows "exec ls" without parentheses.
	// Currently checkInterpreterFunctionCalls treats tclsh as non-strict,
	// so "exec" is only blocked if followed by '(', '=', ':'.

	val := "exec ls"
	template := "{{script}}"
	placeholder := "{{script}}"
	command := "tclsh"
	isShell := false // tclsh is interpreter, not shell

	// This should fail if we properly block tcl injection
	err := checkForShellInjection(val, template, placeholder, command, isShell)
	if err == nil {
		t.Fatalf("VULNERABILITY: tclsh injection '%s' was NOT blocked", val)
	}
}
