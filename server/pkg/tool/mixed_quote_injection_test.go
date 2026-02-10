package tool

import (
	"strings"
	"testing"
)

func TestCheckForShellInjection_MixedQuotes(t *testing.T) {
	// Vulnerability Reproduction: Mixed Double and Single Quotes
	// The user input is used in both contexts.
	template := `echo "{{input}}"; echo '{{input}}'`
	placeholder := "{{input}}"
	command := "bash"

	// Payload that is safe for Double Quotes (no ", $, `) but breaks Single Quotes (')
	val := "'; ls -la; #"

	err := checkForShellInjection(val, template, placeholder, command)

	if err == nil {
		t.Errorf("VULNERABILITY: checkForShellInjection allowed single quote payload in mixed quote context. Payload: %s", val)
	} else {
		t.Logf("Correctly blocked: %v", err)
		if !strings.Contains(err.Error(), "single quote") {
			t.Errorf("Expected error about single quote, got: %v", err)
		}
	}
}

func TestCheckForShellInjection_MixedDoubleAndUnquoted(t *testing.T) {
	// Mixed Double and Unquoted
	// echo "{{input}}" {{input}}
	template := `echo "{{input}}" {{input}}`
	placeholder := "{{input}}"
	command := "bash"

	// Payload that is safe for Double Quotes (e.g. spaces) but unsafe for Unquoted
	val := "hello world"

	err := checkForShellInjection(val, template, placeholder, command)

	if err == nil {
		t.Errorf("VULNERABILITY: checkForShellInjection allowed space payload in mixed unquoted context. Payload: %s", val)
	} else {
		t.Logf("Correctly blocked: %v", err)
		if !strings.Contains(err.Error(), "dangerous character") { // Unquoted check returns "dangerous character"
			t.Errorf("Expected error about dangerous character, got: %v", err)
		}
	}
}

func TestCheckForShellInjection_MixedBacktickAndSingle(t *testing.T) {
	// Mixed Backtick and Single
	// `echo {{input}}` '{{input}}'
	template := "`echo {{input}}` '{{input}}'"
	placeholder := "{{input}}"
	command := "node" // Node allows backticks as template literals (safeish?) but checkForShellInjection treats as Backtick context

	// Payload containing single quote. Safe for backtick (usually), unsafe for single quote.
	val := "'"

	err := checkForShellInjection(val, template, placeholder, command)

	if err == nil {
		t.Errorf("VULNERABILITY: checkForShellInjection allowed single quote payload in mixed backtick/single context. Payload: %s", val)
	} else {
		t.Logf("Correctly blocked: %v", err)
	}
}

func TestAnalyzeQuoteContext_Bitmask(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		placeholder string
		expected    int
	}{
		{
			name:        "Unquoted",
			template:    `echo {{input}}`,
			placeholder: "{{input}}",
			expected:    QuoteLevelUnquoted,
		},
		{
			name:        "Double",
			template:    `echo "{{input}}"`,
			placeholder: "{{input}}",
			expected:    QuoteLevelDouble,
		},
		{
			name:        "Single",
			template:    `echo '{{input}}'`,
			placeholder: "{{input}}",
			expected:    QuoteLevelSingle,
		},
		{
			name:        "Backtick",
			template:    `echo ` + "`{{input}}`",
			placeholder: "{{input}}",
			expected:    QuoteLevelBacktick,
		},
		{
			name:        "Mixed Double Single",
			template:    `echo "{{input}}" '{{input}}'`,
			placeholder: "{{input}}",
			expected:    QuoteLevelDouble | QuoteLevelSingle,
		},
		{
			name:        "Mixed All",
			template:    `{{input}} "{{input}}" '{{input}}'`,
			placeholder: "{{input}}",
			expected:    QuoteLevelUnquoted | QuoteLevelDouble | QuoteLevelSingle,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzeQuoteContext(tt.template, tt.placeholder)
			if got != tt.expected {
				t.Errorf("analyzeQuoteContext() = %v, expected %v", got, tt.expected)
			}
		})
	}
}
