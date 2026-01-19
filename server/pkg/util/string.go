// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import "unicode/utf8"

// LevenshteinDistance calculates the Levenshtein distance between two strings.
// It returns the minimum number of single-character edits (insertions, deletions, or substitutions)
// required to change one string into the other.
func LevenshteinDistance(s1, s2 string) int {
	if s1 == s2 {
		return 0
	}
	// Optimization: Check ASCII once to avoid repeated scans
	s1ASCII := isASCII(s1)
	s2ASCII := isASCII(s2)

	// If both are ASCII, use the highly optimized ASCII-only version
	if s1ASCII && s2ASCII {
		return levenshteinASCII(s1, s2)
	}

	// ⚡ Bolt Optimization: For mixed cases (one ASCII, one non-ASCII),
	// we avoid converting the ASCII string to []rune, saving allocations.
	var r1, r2 []rune
	if !s1ASCII {
		r1 = []rune(s1)
	}
	if !s2ASCII {
		r2 = []rune(s2)
	}

	return levenshteinGeneric(s1, r1, s2, r2)
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
		c1 := s1[i-1]
		for j := 1; j <= m; j++ {
			cost := 0
			if c1 != s2[j-1] {
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

// levenshteinGeneric handles cases where at least one string is non-ASCII.
// It avoids allocations for ASCII strings by using raw string access when possible.
// r1/r2 are the rune slices of s1/s2 if they are non-ASCII, or nil if they are ASCII.
func levenshteinGeneric(s1 string, r1 []rune, s2 string, r2 []rune) int {
	var n, m int
	if r1 != nil {
		n = len(r1)
	} else {
		n = len(s1)
	}
	if r2 != nil {
		m = len(r2)
	} else {
		m = len(s2)
	}

	if n == 0 {
		return m
	}
	if m == 0 {
		return n
	}

	// Ensure m <= n to minimize memory usage
	if m > n {
		return levenshteinGeneric(s2, r2, s1, r1)
	}

	// Optimization: Use stack allocation for small strings.
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
		var c1 rune
		if r1 != nil {
			c1 = r1[i-1]
		} else {
			c1 = rune(s1[i-1])
		}

		if r2 != nil {
			for j := 1; j <= m; j++ {
				cost := 0
				if c1 != r2[j-1] {
					cost = 1
				}
				v1[j] = min(v0[j]+1, v1[j-1]+1, v0[j-1]+cost)
			}
		} else {
			for j := 1; j <= m; j++ {
				cost := 0
				if c1 != rune(s2[j-1]) {
					cost = 1
				}
				v1[j] = min(v0[j]+1, v1[j-1]+1, v0[j-1]+cost)
			}
		}

		// Swap v0 and v1
		v0, v1 = v1, v0
	}

	return v0[m]
}
