// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"
)

func TestRedactDSN_Leak_AtInPassword(t *testing.T) {
	// Case 3: Password contains @ and no scheme (url.Parse fails)
	// "user:p@ssword@host"
	dsn := "user:p@ssword@host"

	redacted := RedactDSN(dsn)
	t.Logf("Redacted: %s", redacted)

	// We expect "user:[REDACTED]@host"
	if redacted == "user:[REDACTED]@ssword@host" {
		t.Errorf("Leaked part of password containing @: %s", redacted)
	}

	if redacted != "user:[REDACTED]@host" {
		t.Errorf("Expected user:[REDACTED]@host, got %s", redacted)
	}
}

func TestRedactDSN_SpaceFalsePositive(t *testing.T) {
	// "Contact: bob@example.com"
	// Should NOT be redacted because of space.
	dsn := "Contact: bob@example.com"
	redacted := RedactDSN(dsn)
	t.Logf("Redacted: %s", redacted)

	if redacted != dsn {
		t.Errorf("False positive redaction on string with space: %s", redacted)
	}
}

func TestRedactDSN_NoSpaceFalsePositive(t *testing.T) {
	// "email:bob@example.com"
	// This will still be redacted because we can't distinguish from dsn.
	dsn := "email:bob@example.com"
	redacted := RedactDSN(dsn)
	t.Logf("Redacted: %s", redacted)

	if redacted == dsn {
		// If it is NOT redacted, that's fine too (if we improved heuristic), but current expectation is it might be redacted.
		// My proposed fix redacts it.
	} else if redacted == "email:[REDACTED]@example.com" {
		// Expected behavior for now
	} else {
		t.Errorf("Unexpected redaction: %s", redacted)
	}
}
