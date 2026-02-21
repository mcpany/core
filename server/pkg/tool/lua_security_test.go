package tool

import (
	"testing"
)

func TestLuaInjection(t *testing.T) {
	// Lua allows calling functions with a single string argument without parentheses.
	// e.g. io.popen "ls"
	// Currently checkInterpreterFunctionCalls treats lua as non-strict.
	// "io" is not in objectKeywords.
	// "popen" is in functionKeywords, but only blocked if followed by '('.

	val := "io.popen \"ls\""
	template := "{{script}}"
	placeholder := "{{script}}"
	command := "lua"
	isShell := false

	// This should fail if we properly block lua injection
	err := checkForShellInjection(val, template, placeholder, command, isShell)
	if err == nil {
		t.Fatalf("VULNERABILITY: lua injection '%s' was NOT blocked", val)
	}
}
