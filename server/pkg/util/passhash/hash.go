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
// Summary: Password hashes a password using bcrypt.
//
// Parameters:
//   - password (string): The password parameter.
//
// Returns:
//   - string: The string result.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
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
// Summary: CheckPassword checks if a password matches a hash.
//
// Parameters:
//   - password (string): The password parameter.
//   - hash (string): The hash parameter.
//
// Returns:
//   - bool: The bool result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
