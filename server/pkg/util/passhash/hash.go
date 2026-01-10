// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package passhash provides password hashing utilities using bcrypt.
package passhash

import (
	"crypto/sha512"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Password hashes a password using bcrypt.
// If the password is longer than 72 bytes (bcrypt limit), it is pre-hashed with SHA-512.
// SHA-512 produces 64 bytes, which fits within the 72-byte limit of bcrypt.
func Password(password string) (string, error) {
	passBytes := []byte(password)
	if len(passBytes) > 72 {
		h := sha512.Sum512(passBytes)
		passBytes = h[:]
	}

	bytes, err := bcrypt.GenerateFromPassword(passBytes, bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// CheckPassword checks if a password matches a hash.
// It handles passwords longer than 72 bytes by pre-hashing with SHA-512 if needed.
func CheckPassword(password, hash string) bool {
	passBytes := []byte(password)
	if len(passBytes) > 72 {
		h := sha512.Sum512(passBytes)
		passBytes = h[:]
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), passBytes)
	return err == nil
}
