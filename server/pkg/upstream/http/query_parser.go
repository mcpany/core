// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"net/url"
	"strings"
)

type queryPart struct {
	raw        string
	key        string
	isInvalid  bool
	keyDecoded bool
}

// parseQueryManual parses a raw query string into parts, preserving invalid encodings.
// ⚡ BOLT: Optimized query parsing to avoid allocation in hot path during startup.
// Randomized Selection from Top 5 High-Impact Targets
func parseQueryManual(rawQuery string) []queryPart {
	if rawQuery == "" {
		return nil
	}

	// Count expected parts to pre-allocate slice (heuristic: count '&')
	// This is O(N) but simple byte scan.
	count := 1
	for i := 0; i < len(rawQuery); i++ {
		if rawQuery[i] == '&' {
			count++
		}
	}
	parts := make([]queryPart, 0, count)

	start := 0
	for i := 0; i <= len(rawQuery); i++ {
		if i == len(rawQuery) || rawQuery[i] == '&' {
			p := rawQuery[start:i]
			start = i + 1

			if p == "" {
				continue
			}

			qp := queryPart{raw: p}
			var key, value string

			// Find '='
			eqIdx := strings.IndexByte(p, '=')
			if eqIdx >= 0 {
				key = p[:eqIdx]
				value = p[eqIdx+1:]
			} else {
				key = p
			}

			decodedKey, errKey := url.QueryUnescape(key)
			if errKey == nil {
				qp.key = decodedKey
				qp.keyDecoded = true
			}

			_, errVal := url.QueryUnescape(value)

			if errKey != nil || errVal != nil {
				qp.isInvalid = true
			}
			parts = append(parts, qp)
		}
	}
	return parts
}
