// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

// isValidID checks if the ID contains only allowed characters.
// This matches the validation in util.SanitizeID but returns boolean.
func isValidID(id string) bool {
	if id == "" {
		return false
	}
	// Allow alphanumeric, underscore, hyphen.
	// Matches `[a-zA-Z0-9_-]`
	for _, c := range id {
		isLower := c >= 'a' && c <= 'z'
		isUpper := c >= 'A' && c <= 'Z'
		isDigit := c >= '0' && c <= '9'
		isSpecial := c == '_' || c == '-'

		if !isLower && !isUpper && !isDigit && !isSpecial {
			return false
		}
	}
	return true
}
