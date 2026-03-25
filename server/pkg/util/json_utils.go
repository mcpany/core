// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import (
	"bytes"
)

// skipString returns the index after the JSON string starting at start.
// start must point to the opening quote.
func skipString(input []byte, start int) int {
	// String starts at start, which is '"'
	scanStart := start + 1
	for {
		q := bytes.IndexByte(input[scanStart:], '"')
		if q == -1 {
			return len(input)
		}
		absQ := scanStart + q
		// Check escape

		// Optimization: fast path if previous char is not backslash
		if input[absQ-1] != '\\' {
			return absQ + 1
		}

		backslashes := 1 // we already saw one backslash at input[absQ-1]
		for j := absQ - 2; j >= scanStart; j-- {
			if input[j] == '\\' {
				backslashes++
			} else {
				break
			}
		}
		if backslashes%2 == 0 {
			return absQ + 1
		}
		scanStart = absQ + 1
	}
}
