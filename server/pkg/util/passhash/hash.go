// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package passhash provides password hashing utilities using bcrypt.
package passhash

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Password hashes a password using bcrypt.
//
// Summary: Hashes a password using bcrypt with cost 12.
//
// Parameters:
//   - password: string. The password to hash.
//
// Returns:
//   - string: The hashed password.
//   - error: An error if the hashing fails.
//
// Side Effects:
//   - CPU intensive operation.
func Password(password string) (string, error) {
	// Increase cost to 12 for better security (default is 10)
	const cost = 12
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// CheckPassword checks if a password matches a hash.
//
// Summary: Verifies a password against a bcrypt hash.
//
// Parameters:
//   - password: string. The password to check.
//   - hash: string. The hash to compare against.
//
// Returns:
//   - bool: True if the password matches the hash, false otherwise.
//
// Side Effects:
//   - CPU intensive operation.
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
