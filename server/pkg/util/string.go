// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import "unicode/utf8"

// LevenshteinDistance calculates the Levenshtein distance between two strings.
// It returns the minimum number of single-character edits (insertions, deletions, or substitutions)
// required to change one string into the other.
func LevenshteinDistance(s1, s2 string) int {
	// ⚡ Bolt Optimization: Check ASCII once
	s1ASCII := isASCII(s1)
	s2ASCII := isASCII(s2)

	// Optimization: If both strings are ASCII, we can avoid rune conversion
	// and use stack-based allocation for small strings.
	if s1ASCII && s2ASCII {
		return levenshteinASCII(s1, s2)
	}

	// Optimization: Mixed case or both non-ASCII.
	// We handle mixed cases by only converting non-ASCII strings to runes.
	var r1, r2 []rune
	if !s1ASCII {
		r1 = []rune(s1)
	}
	if !s2ASCII {
		r2 = []rune(s2)
	}

	return levenshteinGeneric(s1, r1, s2, r2)
}

// levenshteinGeneric handles distance calculation where at least one string might be non-ASCII (rX != nil).
// If rX is nil, it assumes sX is ASCII and uses byte indexing.
func levenshteinGeneric(s1 string, r1 []rune, s2 string, r2 []rune) int {
	n := len(s1)
	if r1 != nil {
		n = len(r1)
	}
	m := len(s2)
	if r2 != nil {
		m = len(r2)
	}

	if n == 0 {
		return m
	}
	if m == 0 {
		return n
	}

	// Swap so m is smaller
	if m > n {
		return levenshteinGeneric(s2, r2, s1, r1)
	}

	// Optimization: Use stack allocation for small strings even in generic path.
	var stackBuf [512]int
	var v0, v1 []int

	if m+1 <= 256 {
		v0 = stackBuf[:m+1]
		v1 = stackBuf[m+1 : 2*(m+1)]
	} else {
		v0 = make([]int, m+1)
		v1 = make([]int, m+1)
	}

	for j := 0; j <= m; j++ {
		v0[j] = j
	}

	for i := 1; i <= n; i++ {
		v1[0] = i

		var char1 rune
		if r1 != nil {
			char1 = r1[i-1]
		} else {
			char1 = rune(s1[i-1])
		}

		for j := 1; j <= m; j++ {
			var char2 rune
			if r2 != nil {
				char2 = r2[j-1]
			} else {
				char2 = rune(s2[j-1])
			}

			cost := 0
			if char1 != char2 {
				cost = 1
			}
			v1[j] = min(
				v0[j]+1,      // deletion
				v1[j-1]+1,    // insertion
				v0[j-1]+cost, // substitution
			)
		}
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
