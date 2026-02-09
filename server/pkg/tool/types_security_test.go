package tool

import (
	"strings"
	"testing"
)

func TestAnalyzeQuoteContext_MixedQuotes(t *testing.T) {
	template := `echo "{{input}}"; echo '{{input}}'`
	placeholder := "{{input}}"

	// Expected behavior: The system should detect that the input is used in BOTH double and single quotes.
	// With the fix, it should return 0 (Strict) because Double and Single are incompatible.

	level := analyzeQuoteContext(template, placeholder)

	if level != 0 {
		t.Errorf("Expected level 0 (Strict) for mixed quotes, got %d", level)
	}

	// Now let's see if checkForShellInjection blocks the single quote payload when level is 0.

	val := "'; ls -la; #"
	command := "bash"

	err := checkForShellInjection(val, template, placeholder, command)

	if err == nil {
		t.Errorf("VULNERABILITY: checkForShellInjection allowed single quote payload in mixed quote context. Payload: %s", val)
	} else {
		// Verify that it is blocked because of dangerous characters (;)
		if !strings.Contains(err.Error(), "dangerous character") {
			t.Errorf("Expected blocked due to dangerous character, got: %v", err)
		} else {
			t.Logf("Correctly blocked: %v", err)
		}
	}
}

func TestAnalyzeQuoteContext_OtherCombinations(t *testing.T) {
	ph := "{{input}}"

	// Case 1: Double + Backtick
	// Template: echo "{{input}}" `echo {{input}}`
	// Double (1) and Backtick (3). Result should be 3 (Backtick strictness).
	// Because Double allows ';' but Backtick executes it.

	tpl1 := `echo "{{input}}" ` + "`echo {{input}}`"
	if l := analyzeQuoteContext(tpl1, ph); l != 3 {
		t.Errorf("Double + Backtick: Expected 3, got %d", l)
	}

	// Case 2: Single + Backtick
	// Template: echo '{{input}}' `echo {{input}}`
	// Single (2) and Backtick (3). Result should be 3 (Backtick strictness).

	tpl2 := `echo '{{input}}' ` + "`echo {{input}}`"
	if l := analyzeQuoteContext(tpl2, ph); l != 3 {
		t.Errorf("Single + Backtick: Expected 3, got %d", l)
	}

	// Case 3: Unquoted + Anything
	// Template: echo {{input}} "{{input}}"
	// Unquoted (0) + Double (1). Result should be 0.

	tpl3 := `echo {{input}} "{{input}}"`
	if l := analyzeQuoteContext(tpl3, ph); l != 0 {
		t.Errorf("Unquoted + Double: Expected 0, got %d", l)
	}
}

func TestCheckForShellInjection_BacktickWithDouble(t *testing.T) {
	// Template: echo "{{input}}" `echo {{input}}`
	tpl := `echo "{{input}}" ` + "`echo {{input}}`"
	ph := "{{input}}"
	cmd := "bash"

	// Payload: ; rm -rf /
	// This is safe in double quotes ("") but dangerous in backticks (``).
	val := "; rm -rf /"

	err := checkForShellInjection(val, tpl, ph, cmd)
	if err == nil {
		t.Errorf("VULNERABILITY: allowed dangerous payload in backtick context mixed with double quotes. Payload: %s", val)
	} else {
		t.Logf("Correctly blocked: %v", err)
	}
}
