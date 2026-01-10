// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package passhash

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"testing"
)

func TestLongPasswordSupport(t *testing.T) {
	// Assert that we CAN handle passwords longer than 72 bytes (bcrypt limit).
	// Generate random password to avoid "Hardcoded Credential" alerts
	bytes := make([]byte, 40)
	if _, err := rand.Read(bytes); err != nil {
		t.Fatal(err)
	}
	base := hex.EncodeToString(bytes) // 80 chars
	longPass := base + strings.Repeat("a", 10) // 90 chars

	hash, err := Password(longPass)
	if err != nil {
		t.Fatalf("Password() failed for long password: %v", err)
	}

	if !CheckPassword(longPass, hash) {
		t.Fatal("CheckPassword() failed for long password")
	}
}

func TestPassword_ShortPassword_Regress(t *testing.T) {
	// Ensure short passwords still work without pre-hashing
	// Generate random password
	bytes := make([]byte, 10)
	if _, err := rand.Read(bytes); err != nil {
		t.Fatal(err)
	}
	shortPass := hex.EncodeToString(bytes)

	hash, err := Password(shortPass)
	if err != nil {
		t.Fatalf("Password() failed: %v", err)
	}
	if !CheckPassword(shortPass, hash) {
		t.Fatal("CheckPassword() failed for short password")
	}
}
