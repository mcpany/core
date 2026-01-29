// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

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
//
// Parameters:
//   - input: The JSON input to walk.
//   - visitor: A function that visits every string value.
//
// Returns:
//   - []byte: The potentially modified JSON output.
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

		// Check for potential comments before the quote
		// Comments start with '/'
		// We loop because we might find a slash that is NOT a comment (e.g. division),
		// but a subsequent slash IS a comment.
		segment := input[i:quotePos]
		searchOffset := 0
		foundComment := false
		for {
			idx := bytes.IndexByte(segment[searchOffset:], '/')
			if idx == -1 {
				break
			}
			slashPos := i + searchOffset + idx
			if slashPos+1 < n {
				next := input[slashPos+1]
				if next == '/' || next == '*' {
					// It is a comment!
					// Skip it and retry scanning from after comment
					i = skipWhitespaceAndComments(input, slashPos)
					foundComment = true
					break
				}
			}
			searchOffset += idx + 1
		}
		if foundComment {
			continue
		}

		// Find end of string using the shared skipString helper
		endQuote := skipString(input, quotePos)
		if endQuote > n {
			endQuote = n
		}

		// Check if this string is a key.
		// It is a key if it is followed by a colon (ignoring whitespace and comments)
		isKey := false
		j := skipWhitespaceAndComments(input, endQuote)
		if j < n && input[j] == ':' {
			isKey = true
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
