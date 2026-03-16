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
//   - password (string): The textual representation of password.
//
// Returns:
//   - string: The resulting text.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//   - string: The resulting text.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func Password(password string) (string, error) {
	// Increase cost to 12 for better security (default is 10)
// CheckPassword checks if a password matches a hash.
//
// Summary: CheckPassword checks if a password matches a hash.
//
// Parameters:
//   - password (string): The textual representation of password.
//   - hash (string): The textual representation of hash.
//
// Returns:
//   - bool: True if successful or valid, false otherwise.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
// CheckPassword checks if a password matches a hash.
//
// Summary: CheckPassword checks if a password matches a hash.
//
// Parameters:
//   - password (string): The textual representation of password.
//   - hash (string): The textual representation of hash.
//
// Returns:
//   - bool: True if successful or valid, false otherwise.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
