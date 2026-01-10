// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package passhash

import (
	"strings"
	"testing"
)

func TestLongPasswordSupport(t *testing.T) {
	// Assert that we CAN handle passwords longer than 72 bytes (bcrypt limit).
	longPassword := strings.Repeat("a", 73)

	hash, err := Password(longPassword)
	if err != nil {
		t.Fatalf("Password() failed for long password: %v", err)
	}

	if !CheckPassword(longPassword, hash) {
		t.Fatal("CheckPassword() failed for long password")
	}
}

func TestPassword_ShortPassword_Regress(t *testing.T) {
	// Ensure short passwords still work without pre-hashing
	shortPassword := "short"
	hash, err := Password(shortPassword)
	if err != nil {
		t.Fatalf("Password() failed: %v", err)
	}
	if !CheckPassword(shortPassword, hash) {
		t.Fatal("CheckPassword() failed for short password")
	}
}
