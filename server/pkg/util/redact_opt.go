// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

// sensitiveKeysByStartChar stores sensitive keys grouped by their starting character (lowercase).
// It is populated in init().
var sensitiveKeysByStartChar [256][]string

// containsAnySensitiveKey checks if the input contains any of the sensitive keys
// using a single-pass optimized search.
func containsAnySensitiveKey(input []byte) bool {
	n := len(input)
	for i := 0; i < n; i++ {
		c := input[i]
		// Convert to lowercase for the lookup if it's an uppercase letter
		if c >= 'A' && c <= 'Z' {
			c += 32
		}

		// Check potential patterns starting with this character
		patterns := sensitiveKeysByStartChar[c]
		if len(patterns) > 0 {
			for _, pattern := range patterns {
				if checkMatch(input[i:], pattern) {
					return true
				}
			}
		}
	}
	return false
}

// checkMatch checks if the byte slice starts with the given pattern (case-insensitive).
// The pattern must be lowercase.
func checkMatch(s []byte, pattern string) bool {
	if len(s) < len(pattern) {
		return false
	}
	for j := 0; j < len(pattern); j++ {
		c := s[j]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		if c != pattern[j] {
			return false
		}
	}
	return true
}
