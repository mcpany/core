// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"
)

func TestSqlite3Injection(t *testing.T) {
	// sqlite3 "SELECT * FROM users WHERE id={{id}}"
	// id = "1 OR 1=1"
	// Quote level 0 (Unquoted) because {{id}} is not inside quotes in the template "SELECT * FROM users WHERE id={{id}}"

	// Case 1: Unquoted in the template (argument to sqlite3)
	// Template: "SELECT * FROM users WHERE id={{id}}"
	// Placeholder: "{{id}}"
	// Value: "1 OR 1=1"
	// Command: "sqlite3"

	val := "1 OR 1=1"
	template := "SELECT * FROM users WHERE id={{id}}"
	placeholder := "{{id}}"
	command := "sqlite3"
	isShell := false

	err := checkForShellInjection(val, template, placeholder, command, isShell, false)
	if err == nil {
		t.Fatalf("Expected error (blocked), got nil")
	}

	// Case 2: Another common injection
	// Value: "1; DROP TABLE users; --"
	// This one IS blocked by checkUnquotedInjection because of ';'
	val = "1; DROP TABLE users; --"
	err = checkForShellInjection(val, template, placeholder, command, isShell, false)
	if err == nil {
		t.Errorf("Expected error for ';', got nil")
	}

	// Case 3: UNION SELECT
	val = "1 UNION SELECT username, password FROM users"
	err = checkForShellInjection(val, template, placeholder, command, isShell, false)
	if err == nil {
		t.Fatalf("Expected error (blocked), got nil")
	}

	// Case 4: Valid input
	val = "1"
	err = checkForShellInjection(val, template, placeholder, command, isShell, false)
	if err != nil {
		t.Fatalf("Expected nil for valid input, got: %v", err)
	}

	// Case 5: Valid string literal (if user uses quotes in template)
	// Template: "SELECT * FROM users WHERE name='{{name}}'"
	// Value: "foo"
	templateQuoted := "SELECT * FROM users WHERE name='{{name}}'"
	placeholderQuoted := "{{name}}"
	val = "foo"
	err = checkForShellInjection(val, templateQuoted, placeholderQuoted, command, isShell, false)
	if err != nil {
		t.Fatalf("Expected nil for valid quoted input, got: %v", err)
	}

	// Case 6: Attempt injection inside quoted template
	// Value: "foo' OR '1'='1"
	// This should be blocked by checkShellInjection level 2 (single quote check)
	val = "foo' OR '1'='1"
	err = checkForShellInjection(val, templateQuoted, placeholderQuoted, command, isShell, false)
	if err == nil {
		t.Fatalf("Expected error for single quote injection in quoted context, got nil")
	}
}
