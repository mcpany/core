// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:var-naming // util is a common package names

import (
	"bytes"
)

// WalkJSONStrings visits every string value in the JSON input.
// visitor is called for every string value (not keys).
// visitor receives:
//
//	raw: the raw bytes of the string, including quotes.
//
// visitor returns:
//
//	replacement: the new bytes to replace 'raw' with, or nil to keep original.
//	modified: true if replacement should be used.
func WalkJSONStrings(input []byte, visitor func(raw []byte) ([]byte, bool)) []byte {
	var out []byte
	i := 0
	lastWrite := 0
	n := len(input)

	for i < n {
		// Scan for the next quote which might start a string
		nextQuote := bytes.IndexByte(input[i:], '"')
		if nextQuote == -1 {
			break
		}
		quotePos := i + nextQuote

		// Find end of string using the shared skipString helper
		endQuote := skipString(input, quotePos)
		if endQuote > n {
			endQuote = n
		}

		// Check if this string is a key.
		// It is a key if it is followed by a colon (ignoring whitespace)
		isKey := false
		j := endQuote
		for j < n {
			c := input[j]
			if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
				j++
				continue
			}
			if c == ':' {
				isKey = true
			}
			break
		}

		if !isKey {
			// It is a value
			raw := input[quotePos:endQuote]
			replacement, modified := visitor(raw)
			if modified {
				if out == nil {
					// Allocate buffer
					// Heuristic: start with input size + small buffer
					extra := len(input) / 10
					if extra < 128 {
						extra = 128
					}
					out = make([]byte, 0, len(input)+extra)
				}
				out = append(out, input[lastWrite:quotePos]...)
				out = append(out, replacement...)
				lastWrite = endQuote
			}
		}

		i = endQuote
	}

	if out == nil {
		return input
	}
	out = append(out, input[lastWrite:]...)
	return out
}
