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

		// Check for potential comments before the quote.
		// We scan the segment between the last processed position (i) and the next quote (quotePos)
		// for any occurrence of a comment start sequence ('//' or '/*').
		// We must check ALL slashes because a division operator ('/') might precede a real comment.
		segment := input[i:quotePos]
		searchOffset := 0
		foundComment := false
		for {
			idx := bytes.IndexByte(segment[searchOffset:], '/')
			if idx == -1 {
				// No more slashes in this segment
				break
			}
			// Calculate absolute position of the slash
			slashIdx := searchOffset + idx
			slashPos := i + slashIdx

			// Check if this slash starts a comment
			if slashPos+1 < n {
				next := input[slashPos+1]
				if next == '/' || next == '*' {
					// It is a comment!
					// Skip it and retry scanning from the position after the comment.
					i = skipWhitespaceAndComments(input, slashPos)
					foundComment = true
					break
				}
			}
			// Not a comment (e.g. division operator), continue searching the segment
			searchOffset = slashIdx + 1
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
