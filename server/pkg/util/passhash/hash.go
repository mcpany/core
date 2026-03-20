// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package passhash provides password hashing utilities using bcrypt.
// Summary: Password hashes a password using bcrypt.
//
// Parameters:
//   - password: The password to hash.
//
// Returns:
//   - string: The hashed password.
//   - error: An error if the hashing fails.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
package passhash

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func Password(password string) (string, error) {
	// Increase cost to 12 for better security (default is 10)
	// Summary: CheckPassword checks if a password matches a hash.
	//
	// Parameters:
	//   - password: The password to check.
	//   - hash: The hash to compare against.
	//
	// Returns:
	//   - bool: True if the password matches the hash, false otherwise.
	//
	// Errors:
	//   - None.
	//
	// Side Effects:
	//   - None.
	const cost = 12
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
