package tool

import (
	"testing"
)

// Tests in this file verify protection against "Backslash Injection" in interpreters.
// This attack vector involves using a backslash to escape a closing quote in a single-quoted string,
// allowing the attacker to break out of the string literal and inject commands.

func TestMySQLBackslashInjection(t *testing.T) {
	// Case: MySQL SQL Injection using backslash escaping
	// Command: mysql -e "SELECT * FROM users WHERE name = '{{NAME}}' AND role = '{{ROLE}}'"
	// Input NAME: \
	// Input ROLE:  OR 1=1 --
	//
	// Resulting query: SELECT * FROM users WHERE name = '\' AND role = ' OR 1=1 --'
	// The string 'name' becomes '\' AND role = '.
	// The rest ' OR 1=1 --' is executed as SQL.

	valName := "\\"
	valRole := " OR 1=1 --"

	// Simulate the template after NAME substitution (partially) or check individually.
	// We check each input against the template.

	template := "SELECT * FROM users WHERE name = '{{NAME}}' AND role = '{{ROLE}}'"
	placeholderName := "{{NAME}}"
	placeholderRole := "{{ROLE}}"
	command := "mysql"

	// Check NAME input
	// It is inside single quotes in the template -> QuoteLevel 2
	err := checkForShellInjection(valName, template, placeholderName, command, false)
	if err == nil {
		t.Errorf("Expected error for backslash injection in MySQL single-quoted context, got nil (VULNERABLE)")
	}

	// Check ROLE input
	// It is inside single quotes in the template -> QuoteLevel 2
	// But it contains spaces and --.
	// If the first input succeeded (backslash), this input becomes the payload.
	// Individually, this input is safe because it doesn't break out of single quotes (no single quote in it).
	// The vulnerability relies on the first input escaping the quote.
	err = checkForShellInjection(valRole, template, placeholderRole, command, false)
	if err != nil {
		// This is expected to be nil currently because it's just a string inside quotes
		// t.Logf("Role input checks out as safe: %v", err)
	}
}

func TestPythonBackslashInjection(t *testing.T) {
	// Case: Python Injection using backslash escaping
	// Command: python -c "print('{{A}}', '{{B}}')"
	// Input A: \
	// Input B: ); import os; os.system('id'); #

	valA := "\\"
	valB := "); import os; os.system('id'); #"

	template := "print('{{A}}', '{{B}}')"
	placeholderA := "{{A}}"
	placeholderB := "{{B}}"
	command := "python"

	// Check A input
	// Inside single quotes -> QuoteLevel 2
	err := checkForShellInjection(valA, template, placeholderA, command, false)
	if err == nil {
		t.Errorf("Expected error for backslash injection in Python single-quoted context, got nil (VULNERABLE)")
	}

	// Check B input
	// Inside single quotes -> QuoteLevel 2
	// Individually safe (no single quotes)
	err = checkForShellInjection(valB, template, placeholderB, command, false)
	if err != nil {
		// Expected safe individually
	}
}

func TestRubyBackslashInjection(t *testing.T) {
	// Case: Ruby Injection using backslash escaping
	valA := "\\"
	template := "puts('{{A}}', '{{B}}')"
	placeholderA := "{{A}}"
	command := "ruby"

	err := checkForShellInjection(valA, template, placeholderA, command, false)
	if err == nil {
		t.Errorf("Expected error for backslash injection in Ruby single-quoted context, got nil (VULNERABLE)")
	}
}

func TestPerlBackslashInjection(t *testing.T) {
	// Case: Perl Injection using backslash escaping
	valA := "\\"
	template := "print('{{A}}', '{{B}}')"
	placeholderA := "{{A}}"
	command := "perl"

	err := checkForShellInjection(valA, template, placeholderA, command, false)
	if err == nil {
		t.Errorf("Expected error for backslash injection in Perl single-quoted context, got nil (VULNERABLE)")
	}
}

func TestNodeBackslashInjection(t *testing.T) {
	// Case: Node Injection using backslash escaping
	valA := "\\"
	template := "console.log('{{A}}', '{{B}}')"
	placeholderA := "{{A}}"
	command := "node"

	err := checkForShellInjection(valA, template, placeholderA, command, false)
	if err == nil {
		t.Errorf("Expected error for backslash injection in Node single-quoted context, got nil (VULNERABLE)")
	}
}
