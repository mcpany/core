// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

// LevenshteinDistance calculates the Levenshtein distance between two strings.
// It returns the minimum number of single-character edits (insertions, deletions, or substitutions)
// required to change one string into the other.
func LevenshteinDistance(s1, s2 string) int {
	// Optimization: If both strings are ASCII, we can avoid rune conversion
	if isASCII(s1) && isASCII(s2) {
		return levenshteinASCII(s1, s2)
	}
	return levenshteinRunes([]rune(s1), []rune(s2))
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 128 {
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

	v := make([]int, m+1)
	for j := 0; j <= m; j++ {
		v[j] = j
	}

	for i := 1; i <= n; i++ {
		diag := v[0]
		v[0] = i

		for j := 1; j <= m; j++ {
			temp := v[j]
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}
			v[j] = min(v[j]+1, v[j-1]+1, diag+cost)
			diag = temp
		}
	}

	return v[m]
}

func levenshteinRunes(s1, s2 []rune) int {
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

	v := make([]int, m+1)
	for j := 0; j <= m; j++ {
		v[j] = j
	}

	for i := 1; i <= n; i++ {
		diag := v[0]
		v[0] = i

		for j := 1; j <= m; j++ {
			temp := v[j]
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}
			v[j] = min(v[j]+1, v[j-1]+1, diag+cost)
			diag = temp
		}
	}

	return v[m]
}
