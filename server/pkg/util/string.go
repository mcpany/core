// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

// LevenshteinDistance calculates the Levenshtein distance between two strings.
// It returns the minimum number of single-character edits (insertions, deletions, or substitutions)
// required to change one string into the other.
func LevenshteinDistance(s1, s2 string) int {
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
