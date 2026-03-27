// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package passhash provides password hashing utilities using bcrypt.
package passhash

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Password provides password functionality.
//
// Summary: Password.
//
// Parameters.
//   - password: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func Password(password string) (string, error) {
	// Increase cost to 12 for better security (default is 10)
	const cost = 12
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// CheckPassword provides checkpassword functionality.
//
// Summary: CheckPassword.
//
// Parameters.
//   - password: The parameter.
//   - hash: The parameter.
//
// Returns.
//   - result: The result.
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
