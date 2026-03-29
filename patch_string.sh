cat << 'INNER_EOF' > server/pkg/util/string.go.tmp
// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import "unicode/utf8"

// LevenshteinDistance calculates the Levenshtein distance between two strings.
func LevenshteinDistance(s1, s2 string) int {
	return LevenshteinDistanceWithLimit(s1, s2, len(s1)+len(s2))
}

// LevenshteinDistanceWithLimit calculates the Levenshtein distance with an upper limit.
func LevenshteinDistanceWithLimit(s1, s2 string, limit int) int {
	if isASCII(s1) && isASCII(s2) {
		return levenshteinASCIIBounded(s1, s2, limit)
	}

	r1 := []rune(s1)
	r2 := []rune(s2)

	if len(r1) < len(r2) {
		r1, r2 = r2, r1
	}

	if len(r2) == 0 {
		return len(r1)
	}

	n, m := len(r1), len(r2)

	if n-m > limit {
		return limit + 1
	}

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
		minRow := v1[0]
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
			if v1[j] < minRow {
				minRow = v1[j]
			}
		}

		if minRow > limit {
			return limit + 1
		}

		for k := 0; k <= m; k++ {
			v0[k] = v1[k]
		}
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

	if n-m > limit {
		return limit + 1
	}

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

		if minRow > limit {
			return limit + 1
		}

		for k := 0; k <= m; k++ {
			v0[k] = v1[k]
		}
	}

	return v0[m]
}
INNER_EOF
mv server/pkg/util/string.go.tmp server/pkg/util/string.go
