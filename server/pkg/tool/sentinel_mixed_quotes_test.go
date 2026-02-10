package tool

import (
	"testing"
)

func TestSentinel_MixedQuotesRegression(t *testing.T) {
	// Scenario: A template uses the placeholder in BOTH double and single quotes.
	// This creates a conflict where a payload safe for one (Double) is dangerous for the other (Single).
	template := `echo "{{input}}"; echo '{{input}}'`
	placeholder := "{{input}}"

	// 1. Verify analyzeQuoteContext returns 0 (Strict/Unquoted) due to conflict
	level := analyzeQuoteContext(template, placeholder)

	if level != 0 {
		t.Errorf("Expected level 0 (Strict/Unquoted) due to mixed quote context, got %d", level)
	} else {
		t.Log("Correctly detected mixed quote context and fell back to Strict mode.")
	}

	// 2. Verify that a payload containing a single quote is BLOCKED.
	// If it used Level 1 (Double), it would allow single quotes.
	// Since it uses Level 0 (Strict), it should block single quotes.

	val := "'; ls -la; #"
	command := "bash"

	err := checkForShellInjection(val, template, placeholder, command)

	if err == nil {
		t.Errorf("VULNERABILITY: checkForShellInjection failed to block single quote payload in mixed quote context. Payload: %s", val)
	} else {
		t.Logf("Successfully blocked malicious payload: %v", err)
	}

    // 3. Verify that a benign payload (safe for unquoted) still passes
    safeVal := "hello_world"
    err = checkForShellInjection(safeVal, template, placeholder, command)
    if err != nil {
        t.Errorf("False positive: blocked safe payload %q: %v", safeVal, err)
    }
}
