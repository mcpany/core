// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import "unicode/utf8"

// LevenshteinDistance calculates the Levenshtein distance between two strings.
// It returns the minimum number of single-character edits (insertions, deletions, or substitutions)
// required to change one string into the other.
func LevenshteinDistance(s1, s2 string) int {
	// Optimization: If both strings are ASCII, we can avoid rune conversion
	// and use stack-based allocation for small strings.
	if isASCII(s1) && isASCII(s2) {
		return levenshteinASCII(s1, s2)
	}

	r1, r2 := []rune(s1), []rune(s2)
	n, m := len(r1), len(r2)

	if n == 0 {
		return m
	}
	if m == 0 {
		return n
	}

	// We want to iterate over the longer string in the outer loop
	// and use the shorter string for the column vector to minimize memory usage.
	// So if m > n, we swap them so that m is always the smaller (or equal) length.
	if m > n {
		r1, r2 = r2, r1
		n, m = m, n
	}

	// v0 represents the previous row of distances
	v0 := make([]int, m+1)
	// v1 represents the current row of distances
	v1 := make([]int, m+1)

	// Initialize v0 (the first row, where one string is empty)
	for j := 0; j <= m; j++ {
		v0[j] = j
	}

	for i := 1; i <= n; i++ {
		v1[0] = i
		for j := 1; j <= m; j++ {
			cost := 0
			if r1[i-1] != r2[j-1] {
				cost = 1
			}
			v1[j] = min(
				v0[j]+1,      // deletion
				v1[j-1]+1,    // insertion
				v0[j-1]+cost, // substitution
			)
		}
		// Swap v0 and v1 for the next iteration
		v0, v1 = v1, v0
	}

	return v0[m]
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}

func levenshteinASCII(s1, s2 string) int {
	n, m := len(s1), len(s2)

	if n == 0 {
		return m
	}
	if m == 0 {
		return n
	}

	if m > n {
		s1, s2 = s2, s1
		n, m = m, n
	}

	// Optimization: Use stack allocation for small strings.
	// We need m+1 ints for each vector.
	// If m < 64, we can use a stack array of 128 ints (64 * 2).
	// This covers common cases like tool names.
	var stackBuf [128]int
	var v0, v1 []int

	if m+1 <= 64 {
		v0 = stackBuf[:m+1]
		v1 = stackBuf[m+1 : 2*(m+1)]
	} else {
		v0 = make([]int, m+1)
		v1 = make([]int, m+1)
	}

	// Initialize v0
	for j := 0; j <= m; j++ {
		v0[j] = j
	}

	for i := 1; i <= n; i++ {
		v1[0] = i
		for j := 1; j <= m; j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}
			v1[j] = min(
				v0[j]+1,      // deletion
				v1[j-1]+1,    // insertion
				v0[j-1]+cost, // substitution
			)
		}
		// Swap v0 and v1
		v0, v1 = v1, v0
	}

	return v0[m]
}
