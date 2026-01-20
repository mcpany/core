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

	// Optimization: avoid []rune conversion for the longer string.
	// We use len(s) (bytes) as a heuristic for swapping, because converting to []rune
	// just to check length defeats the purpose of avoiding allocation.
	if len(s1) < len(s2) {
		s1, s2 = s2, s1
	}

	// If s2 is empty, the distance is the rune count of s1
	if len(s2) == 0 {
		return utf8.RuneCountInString(s1)
	}

	// Convert shorter string to runes for fast random access
	r2 := []rune(s2)
	m := len(r2)

	// Stack allocation for v0, v1 if m is small enough.
	// Reusing the same logic as levenshteinASCII.
	// 256 ints is a safe stack size.
	var stackBuf [512]int
	var v0, v1 []int

	if m+1 <= 256 {
		v0 = stackBuf[:m+1]
		v1 = stackBuf[m+1 : 2*(m+1)]
	} else {
		v0 = make([]int, m+1)
		v1 = make([]int, m+1)
	}

	// Initialize v0 (the first row, where one string is empty)
	for j := 0; j <= m; j++ {
		v0[j] = j
	}

	i := 0
	for _, c1 := range s1 {
		i++
		v1[0] = i
		for j := 1; j <= m; j++ {
			cost := 0
			if c1 != r2[j-1] {
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
	// If m <= 255, we can use a stack array of 512 ints (256 * 2).
	// This covers common cases like tool names and medium descriptions.
	var stackBuf [512]int
	var v0, v1 []int

	if m+1 <= 256 {
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
