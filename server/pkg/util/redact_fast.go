// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import (
	"bytes"
	"encoding/json"
)

// redactJSONFast is a zero-allocation (mostly) implementation of RedactJSON.
// It scans the input byte slice and constructs a new slice with redacted values.
// It avoids full JSON parsing.
func redactJSONFast(input []byte) []byte {
	// Lazy allocation: only allocate result buffer if we actually need to redact something.
	var out []byte

	i := 0
	lastWrite := 0
	n := len(input)

	for i < n {
		// Scan for the next quote which might start a key
		nextQuote := bytes.IndexByte(input[i:], '"')
		if nextQuote == -1 {
			// No more strings, break to write rest
			break
		}
		quotePos := i + nextQuote

		// Parse string
		// We need to find the matching closing quote
		// Handle escapes: \\ and \"
		var endQuote int
		// Optimization: fast scan for closing quote
		scanStart := quotePos + 1
		malformed := false
		for {
			q := bytes.IndexByte(input[scanStart:], '"')
			if q == -1 {
				// Malformed JSON (unclosed string)
				malformed = true
				break
			}
			absQ := scanStart + q
			// Check for escape
			// Count backslashes before absQ
			backslashes := 0
			for j := absQ - 1; j >= scanStart; j-- {
				if input[j] == '\\' {
					backslashes++
				} else {
					break
				}
			}
			if backslashes%2 == 0 {
				// Even number of backslashes means the quote is NOT escaped
				endQuote = absQ
				break
			}
			// Odd number means it IS escaped, continue
			scanStart = absQ + 1
		}

		if malformed {
			// Stop processing and flush
			// i points to where we started scanning for strings
			// But we want to preserve input as is if malformed
			break
		}

		// Check if this string is a key.
		// It is a key if it is followed by a colon (ignoring whitespace)
		isKey := false
		colonPos := -1
		// Scan after endQuote+1 for colon
		j := endQuote + 1
		for j < n {
			c := input[j]
			if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
				j++
				continue
			}
			if c == ':' {
				isKey = true
				colonPos = j
			}
			break
		}

		if isKey {
			// Check if key is sensitive
			keyContent := input[quotePos+1 : endQuote]

			var keyToCheck []byte
			if bytes.Contains(keyContent, []byte{'\\'}) {
				// Key contains escape sequences, unescape it to check for sensitivity.
				// This is slower but safer.
				// Limit the key size to prevent excessive allocation on malicious input
				if len(keyContent) > 1024 {
					keyToCheck = keyContent
				} else {
					quoted := make([]byte, len(keyContent)+2)
					quoted[0] = '"'
					copy(quoted[1:], keyContent)
					quoted[len(quoted)-1] = '"'

					var unescaped string
					if err := json.Unmarshal(quoted, &unescaped); err == nil {
						keyToCheck = []byte(unescaped)
					} else {
						// Fallback to raw content if unmarshal fails
						keyToCheck = keyContent
					}
				}
			} else {
				keyToCheck = keyContent
			}

			// Optimization: We check the raw key content against sensitive keys.
			// We only unescape if backslashes are detected, avoiding expensive json.Unmarshal calls for the common case.
			// Note: scanForSensitiveKeys (used in the pre-check) also does not handle escapes.
			sensitive := scanForSensitiveKeys(keyToCheck, false)

			if sensitive {
				// Identify value start
				valStart := colonPos + 1
				// Skip whitespace
				for valStart < n {
					c := input[valStart]
					if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
						valStart++
						continue
					}
					break
				}

				if valStart < n {
					// Initialize buffer if needed
					if out == nil {
						// Allocate slightly more than input size to avoid reallocations when replacement is longer than original
						// 1.1x is a heuristic, but we cap it to avoid excessive memory usage for large inputs.
						extra := len(input) / 10
						if extra > 16384 {
							extra = 16384
						}
						if extra < 128 {
							extra = 128
						}
						out = make([]byte, 0, len(input)+extra)
					}

					// Determine value end
					valEnd := skipJSONValue(input, valStart)

					// Write pending data up to value start
					out = append(out, input[lastWrite:valStart]...)

					// Write replacement
					out = append(out, redactedValue...)

					// Advance pointers
					i = valEnd
					lastWrite = valEnd
					continue
				}
			}
		}

		// Not sensitive key, or just string value
		// Advance past the string
		i = endQuote + 1
	}

	if out == nil {
		// No redaction occurred
		return input
	}

	// Write remaining
	out = append(out, input[lastWrite:]...)
	return out
}

// skipJSONValue returns the index after the JSON value starting at start.
func skipJSONValue(input []byte, start int) int {
	if start >= len(input) {
		return start
	}

	c := input[start]
	switch c {
	case '"':
		return skipString(input, start)
	case '{':
		return skipObject(input, start)
	case '[':
		return skipArray(input, start)
	case 't', 'f', 'n':
		return skipLiteral(input, start)
	default:
		return skipNumber(input, start)
	}
}

func skipObject(input []byte, start int) int {
	// Object starts at start, which is '{'
	depth := 1
	i := start + 1
	for i < len(input) {
		switch input[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i + 1
			}
		case '"':
			// Skip string to avoid confusion with braces inside strings
			i = skipString(input, i)
			continue // skipString returns index after string, so continue loop
		}
		i++
	}
	return len(input)
}

func skipArray(input []byte, start int) int {
	// Array starts at start, which is '['
	depth := 1
	i := start + 1
	for i < len(input) {
		switch input[i] {
		case '[':
			depth++
		case ']':
			depth--
			if depth == 0 {
				return i + 1
			}
		case '"':
			i = skipString(input, i)
			continue
		}
		i++
	}
	return len(input)
}

func skipLiteral(input []byte, start int) int {
	// true, false, null
	i := start
	for i < len(input) {
		c := input[i]
		if c < 'a' || c > 'z' {
			return i
		}
		i++
	}
	return i
}

func skipNumber(input []byte, start int) int {
	// Number
	// Scan until delimiter: , } ] or whitespace
	i := start
	for i < len(input) {
		c := input[i]
		if c == ',' || c == '}' || c == ']' || c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			return i
		}
		i++
	}
	return i
}
