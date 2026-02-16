// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"
)

func TestSqlite3DoubleQuotedInjection(t *testing.T) {
	// Case 1: SQL Injection with OR
	t.Run("double_quoted_or_injection", func(t *testing.T) {
		val := "1 OR 1=1"
		template := "\"SELECT * FROM users WHERE id={{id}}\"" // Double quoted
		placeholder := "{{id}}"
		command := "sqlite3"
		isShell := false

		err := checkForShellInjection(val, template, placeholder, command, isShell)
		if err == nil {
			t.Fatalf("VULNERABILITY CONFIRMED: Double-quoted SQL injection passed! Value: %q", val)
		} else {
			t.Logf("Secure: Blocked with error: %v", err)
		}
	})

	// Case 2: Single quote injection
	t.Run("double_quoted_single_quote_injection", func(t *testing.T) {
		val := "foo' bar"
		template := "\"SELECT * FROM users WHERE name='{{name}}'\""
		placeholder := "{{name}}"
		command := "sqlite3"
		isShell := false

		err := checkForShellInjection(val, template, placeholder, command, isShell)
		if err == nil {
			t.Fatalf("VULNERABILITY CONFIRMED: Double-quoted single quote injection passed! Value: %q", val)
		} else {
			t.Logf("Secure: Blocked with error: %v", err)
		}
	})
}
