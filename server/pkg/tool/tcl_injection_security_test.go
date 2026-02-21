package tool

import (
	"testing"
)

func TestTclInjection(t *testing.T) {
	// Tcl (tclsh) allows "exec ls" without parentheses.
	// We have now enforced strict mode for tclsh, so "exec" should be blocked as a keyword.

	val := "exec ls"
	template := "{{script}}"
	placeholder := "{{script}}"
	command := "tclsh"
	isShell := false // tclsh is interpreter, not shell

	// This should now RETURN an error (blocked)
	err := checkForShellInjection(val, template, placeholder, command, isShell)
	if err == nil {
		t.Fatalf("VULNERABILITY: tclsh injection '%s' was NOT blocked (expected error)", val)
	}
}

func TestLuaInjection(t *testing.T) {
	// Lua allows "os.execute 'ls'"
	// We enforce strict mode, so "os" or "execute" should be blocked?
	// strict mode blocks "statementKeywords".
	// "os" is in objectKeywords.
	// But in strict mode, ALL keywords (statement, object, function) are added to statementKeywords (universal).
	// So "os" and "execute" are blocked as words.

	val := "os.execute 'ls'"
	template := "{{script}}"
	placeholder := "{{script}}"
	command := "lua"
	isShell := false

	err := checkForShellInjection(val, template, placeholder, command, isShell)
	if err == nil {
		t.Fatalf("VULNERABILITY: lua injection '%s' was NOT blocked (expected error)", val)
	}
}
