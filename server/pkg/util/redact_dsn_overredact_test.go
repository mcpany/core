// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestRedactDSNBug(t *testing.T) {
	// Case where url.Parse works - verified to pass
	input := "postgres://user:password@host.com/db?email=foo@bar.com"
	expected := "postgres://user:[REDACTED]@host.com/db?email=foo@bar.com"
	actual := RedactDSN(input)
	assert.Equal(t, expected, actual, "Should redact password but keep host and parameters (valid URL)")

    // Case where url.Parse FAILS due to invalid port
    // greedy regex should fail here if there is an @ later
    input3 := "postgres://user:password@host:abc/db?email=foo@bar.com"
    expected3 := "postgres://user:[REDACTED]@host:abc/db?email=foo@bar.com"
    actual3 := RedactDSN(input3)
    assert.Equal(t, expected3, actual3, "Should redact password but keep host and parameters (invalid port)")
}
