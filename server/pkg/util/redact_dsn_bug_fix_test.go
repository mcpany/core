// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"
)

func TestRedactDSN_Bug_RedisNoUser(t *testing.T) {
	// Scenario: Redis DSN with empty user and password, no host (or implicit).
	// redis://:mysecretpassword
	// url.Parse fails because ":mysecretpassword" is not a valid port.
	// RedactDSN fallback regex expects '@', so it fails to match.
	// Password is leaked.

	dsn := "redis://:mysecretpassword"
	redacted := RedactDSN(dsn)

	// We expect the password to be redacted.
	expected := "redis://:[REDACTED]"

	if redacted != expected {
		t.Errorf("Failed to redact redis password. Got: %s, Expected: %s", redacted, expected)
	}
}

func TestRedactDSN_Bug_UserPasswordNoHost(t *testing.T) {
    // Scenario: User:Password but no host.
    // postgres://user:mysecretpassword
    // url.Parse fails.
    // Regex fails (no @).

    dsn := "postgres://user:mysecretpassword"
    redacted := RedactDSN(dsn)

    expected := "postgres://user:[REDACTED]"

    if redacted != expected {
        t.Errorf("Failed to redact password without host. Got: %s, Expected: %s", redacted, expected)
    }
}

func TestRedactDSN_Regression_HostPort(t *testing.T) {
    // Scenario: Standard URL with host:port.
    // http://localhost:8080
    // Should NOT be redacted as http://localhost:[REDACTED].

    dsn := "http://localhost:8080"
    redacted := RedactDSN(dsn)

    expected := "http://localhost:8080"

    if redacted != expected {
        t.Errorf("Regression! Valid host:port was redacted. Got: %s, Expected: %s", redacted, expected)
    }
}
