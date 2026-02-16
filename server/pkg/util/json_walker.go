// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import (
	"bytes"
	"math"
)

// jsonWhitespaceTable is a lookup table for fast whitespace checking.
// We define it here to ensure it's available for skipWhitespace even if redact_fast.go
// is not initialized yet or if build tags exclude it (though they are in the same package).
var jsonWhitespaceTable [256]bool

func init() {
	jsonWhitespaceTable[' '] = true
	jsonWhitespaceTable['\t'] = true
	jsonWhitespaceTable['\n'] = true
	jsonWhitespaceTable['\r'] = true
}

// WalkJSONStrings visits every string value in the JSON input.
//
// Summary: Walks through a JSON bytes slice and applies a visitor function to all string values.
//
// It supports non-standard JSON with comments (// and /* */).
//
// Parameters:
//   - input: []byte. The JSON input to walk.
//   - visitor: func(raw []byte) ([]byte, bool). A function that takes the raw string bytes (including quotes) and returns a replacement and a modified flag.
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
					out = make([]byte, 0, calculateCapacity(len(input)))
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

// WalkStandardJSONStrings visits every string value in the JSON input.
//
// Summary: Optimized JSON walker for standard JSON (no comments).
//
// It visits every string value (not keys) and applies the visitor.
//
// Parameters:
//   - input: []byte. The standard JSON input.
//   - visitor: func(raw []byte) ([]byte, bool). A function that takes the raw string bytes and returns a replacement and a modified flag.
//
// Returns:
//   - []byte: The potentially modified JSON output.
func WalkStandardJSONStrings(input []byte, visitor func(raw []byte) ([]byte, bool)) []byte {
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

		// Optimization: Standard JSON has no comments, so we skip the expensive slash scan loop.

		// Find end of string using the shared skipString helper
		endQuote := skipString(input, quotePos)
		if endQuote > n {
			endQuote = n
		}

		// Check if this string is a key.
		// It is a key if it is followed by a colon (ignoring whitespace)
		isKey := false
		j := skipWhitespace(input, endQuote)

		// ⚡ Bolt Fix: Handle edge case where a comment might exist between key and colon.
		// Although we target standard JSON, regression tests cover non-standard JSON (comments).
		// If we hit a slash instead of a colon, we fallback to the comment-aware skipper for just this check.
		if j < n && input[j] == '/' {
			j = skipWhitespaceAndComments(input, endQuote)
		}

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
					out = make([]byte, 0, calculateCapacity(len(input)))
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

// skipWhitespace returns the index of the first non-whitespace character starting from start.
func skipWhitespace(input []byte, start int) int {
	// ⚡ BOLT: Optimized whitespace skipping using lookup table to reduce branching.
	// Randomized Selection from Top 5 High-Impact Targets
	i := start
	n := len(input)
	for i < n {
		c := input[i]
		if jsonWhitespaceTable[c] {
			i++
			continue
		}
		break
	}
	return i
}

// calculateCapacity calculates a safe initial capacity for the output buffer.
// It uses int64 to avoid overflow during calculation and checks against MaxInt.
func calculateCapacity(inputLen int) int {
	// Heuristic: input size + 10% or at least 128 bytes
	extra := int64(inputLen) / 10
	if extra < 128 {
		extra = 128
	}
	targetCap := int64(inputLen) + extra

	// Check for overflow against int limit
	if targetCap > int64(math.MaxInt) {
		return math.MaxInt
	}
	return int(targetCap)
}
