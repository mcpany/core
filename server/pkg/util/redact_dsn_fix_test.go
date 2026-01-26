// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactDSN_SlashInPassword_Fix(t *testing.T) {
	// Scenario: DSN without scheme, password contains a slash.
	// This mimics some non-standard DSNs or user input where slash is not encoded.
	// Previously, the regex strictly forbade slash to avoid matching "://" in URLs.
	// We want to support slash if it doesn't look like "://".

	dsn := "user:pass/word@host"
	redacted := RedactDSN(dsn)

	expected := "user:[REDACTED]@host"

	assert.Equal(t, expected, redacted, "Should redact password containing slash")
}

func TestRedactDSN_SlashAtStartOfPassword(t *testing.T) {
	dsn := "user:/password@host"
	redacted := RedactDSN(dsn)
	expected := "user:[REDACTED]@host"
	assert.Equal(t, expected, redacted)
}

func TestRedactDSN_ConsecutiveSlashes_ShouldNotRedact(t *testing.T) {
	// This looks like a URL scheme "user://..." or path "//..."
	// We should NOT redact this as a password, because it's ambiguous and likely a URL.
	// If we redacted it, we might break "http://host@path".

	// Case 1: Likely a URL with authority
	dsn := "http://user@host"
	// Regex check:
	// If we allow //, matches ://user@.
	// Result http:[REDACTED]host. BAD.

	redacted := RedactDSN(dsn)
	assert.Equal(t, dsn, redacted, "Should NOT redact URL scheme as password")
}
