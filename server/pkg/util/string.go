// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import "unicode/utf8"

// LevenshteinDistance calculates the Levenshtein distance between two strings.
// It returns the minimum number of single-character edits (insertions, deletions, or substitutions)
// required to change one string into the other.
//
// Parameters:
//   - s1: The first string.
//   - s2: The second string.
//
// Returns:
//   - int: The Levenshtein distance.
func LevenshteinDistance(s1, s2 string) int {
	// Pass a very large limit so it behaves like the unbounded version.
	// The maximum possible distance is max(len(s1), len(s2)).
	// We use max int as limit effectively.
	return LevenshteinDistanceWithLimit(s1, s2, len(s1)+len(s2))
}

// LevenshteinDistanceWithLimit calculates the Levenshtein distance with an upper limit.
// If the distance is strictly greater than limit, it returns a value > limit (specifically limit + 1).
//
// Parameters:
//   - s1: The first string.
//   - s2: The second string.
//   - limit: The maximum distance to compute.
//
// Returns:
//   - int: The Levenshtein distance, or limit + 1 if it exceeds the limit.
func LevenshteinDistanceWithLimit(s1, s2 string, limit int) int {
	// Optimization: If both strings are ASCII, we can avoid rune conversion
	// and use stack-based allocation for small strings.
	if isASCII(s1) && isASCII(s2) {
		return levenshteinASCIIBounded(s1, s2, limit)
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

	// ⚡ BOLT: Optimized small string handling to use stack-allocated rune buffer, avoiding heap allocation.
	// Randomized Selection from Top 5 High-Impact Targets

	// Optimization: Use stack buffer for short strings to avoid allocation
	var r2 []rune
	// Stack buffer for small strings (<= 256 runes).
	// This covers >99% of use cases (names, IDs, short text).
	var runeTmp [256]rune

	// First pass to check length and populate stack buffer.
	// If it fits, we avoid heap allocation entirely.
	count := 0
	fits := true
	for _, r := range s2 {
		if count < len(runeTmp) {
			runeTmp[count] = r
			count++
		} else {
			fits = false
			break
		}
	}

	if fits {
		r2 = runeTmp[:count]
	} else {
		// Fallback for long strings
		r2 = []rune(s2)
	}

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
	// v0[j] is distance from s1[:0] (empty) to s2[:j] which is j
	for j := 0; j <= m; j++ {
		v0[j] = j
	}

	i := 0
	for _, c1 := range s1 {
		i++
		v1[0] = i
		minRow := v1[0]
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
			if v1[j] < minRow {
				minRow = v1[j]
			}
		}

		// If the minimum value in this row exceeds the limit, we can stop early.
		if minRow > limit {
			return limit + 1
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

func levenshteinASCIIBounded(s1, s2 string, limit int) int {
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

	// Optimization: Length check
	// Since these are ASCII, length difference is a lower bound on distance.
	if n-m > limit {
		return limit + 1
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
		minRow := v1[0]
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
			if v1[j] < minRow {
				minRow = v1[j]
			}
		}

		// If the minimum value in this row exceeds the limit, we can stop early.
		if minRow > limit {
			return limit + 1
		}

		// Swap v0 and v1
		v0, v1 = v1, v0
	}

	return v0[m]
}
