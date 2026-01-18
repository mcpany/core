// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"strings"
)

// LevenshteinDistance calculates the Levenshtein distance between two strings.
// It returns the minimum number of single-character edits (insertions, deletions or substitutions)
// required to change one string into the other.
func LevenshteinDistance(s, t string) int {
	// Optimization: convert to runes to handle Unicode correctly
	r1, r2 := []rune(s), []rune(t)
	n, m := len(r1), len(r2)

	// Safety check for DoS prevention
	if n > 256 || m > 256 {
		return n + m // Return a large enough distance to not match
	}

	if n == 0 {
		return m
	}
	if m == 0 {
		return n
	}

	// Optimization: Swap to use less memory if m is larger
	if m < n {
		return LevenshteinDistance(t, s)
	}

	// Use two rows instead of a full matrix to save memory
	// previous row
	prev := make([]int, m+1)
	// current row
	curr := make([]int, m+1)

	// Initialize first row
	for j := 0; j <= m; j++ {
		prev[j] = j
	}

	for i := 1; i <= n; i++ {
		curr[0] = i
		for j := 1; j <= m; j++ {
			cost := 1
			if r1[i-1] == r2[j-1] {
				cost = 0
			}
			// min(insertion, deletion, substitution)
			curr[j] = min(
				curr[j-1]+1,     // insertion
				prev[j]+1,       // deletion
				prev[j-1]+cost,  // substitution
			)
		}
		// Copy current to previous for next iteration
		copy(prev, curr)
	}

	return prev[m]
}


// FindClosestMatch finds the closest match for a target string from a list of candidates.
// It returns the best match and true if a match is found within the maxDistance threshold.
// Otherwise it returns empty string and false.
func FindClosestMatch(target string, candidates []string, maxDistance int) (string, bool) {
	bestMatch := ""
	minDist := maxDistance + 1

	targetLower := strings.ToLower(target)

	for _, candidate := range candidates {
		// Optimization: Exact match check (should be handled by caller, but safe to have)
		if candidate == target {
			return candidate, true
		}

		dist := LevenshteinDistance(targetLower, strings.ToLower(candidate))
		if dist < minDist {
			minDist = dist
			bestMatch = candidate
		}
	}

	if minDist <= maxDistance {
		return bestMatch, true
	}

	return "", false
}
