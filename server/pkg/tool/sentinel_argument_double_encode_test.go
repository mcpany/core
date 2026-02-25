// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"
)

func TestArgumentInjectionBypass(t *testing.T) {
	// Test double encoding
	// %252d -> %2d -> -
	// If the code only unescapes once, it sees %2d (safe).
	// If the downstream system unescapes again, it sees - (unsafe).
	// But validateSafePathAndInjection only checks val and Unescape(val).

	// Case 1: Double Encoded Hyphen
	val := "%252d"
	err := validateSafePathAndInjection(val, false, "ls")
	if err != nil {
		t.Logf("Blocked double encoding as expected: %v", err)
	} else {
		t.Errorf("Bypass: Double encoded hyphen accepted: %s", val)
	}

	// Case 2: Unicode Homoglyphs?
	// e.g. U+2010 (Hyphen) or U+2212 (Minus Sign)
	// If the tool normalizes unicode, this could become a hyphen.
	val = "\u2010la"
	err = validateSafePathAndInjection(val, false, "ls")
	if err == nil {
		t.Logf("Accepted unicode hyphen: %s (Check if target tool normalizes this)", val)
	} else {
		t.Logf("Blocked unicode hyphen: %v", err)
	}
}

func TestLocalFileAccessBypass(t *testing.T) {
    // Case 1: file:// with capital letters
    val := "FiLe:///etc/passwd"
    err := validateSafePathAndInjection(val, false, "curl")
    if err != nil {
        t.Logf("Blocked mixed case file scheme: %v", err)
    } else {
        t.Errorf("Bypass: Mixed case file scheme accepted: %s", val)
    }

    // Case 2: file:// with whitespace
    val = "file :///etc/passwd"
    err = validateSafePathAndInjection(val, false, "curl")
    if err != nil {
        t.Logf("Blocked file scheme with space: %v", err)
    } else {
        t.Logf("Accepted file scheme with space: %s", val)
    }
}
