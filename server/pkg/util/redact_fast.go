// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"bytes"
	"encoding/json"
)

// redactJSONFast is a zero-allocation (mostly) implementation of RedactJSON.
// It scans the input byte slice and constructs a new slice with redacted values.
// It avoids full JSON parsing.
func redactJSONFast(input []byte) []byte {
	// Pre-allocate result buffer. Usually it will be same size or slightly smaller/larger.
	// 1.1x input size is a safe bet to avoid reallocations if we replace short values with long placeholders.
	out := bytes.NewBuffer(make([]byte, 0, len(input)))

	i := 0
	n := len(input)

	for i < n {
		// Scan for the next quote which might start a key
		nextQuote := bytes.IndexByte(input[i:], '"')
		if nextQuote == -1 {
			// No more strings, just copy the rest
			out.Write(input[i:])
			break
		}
		quotePos := i + nextQuote

		// Copy everything up to the quote
		out.Write(input[i:quotePos])
		i = quotePos

		// Parse string
		// We need to find the matching closing quote
		// Handle escapes: \\ and \"
		var endQuote int
		// Optimization: fast scan for closing quote
		scanStart := i + 1
		for {
			q := bytes.IndexByte(input[scanStart:], '"')
			if q == -1 {
				// Malformed JSON (unclosed string), just copy the rest
				out.Write(input[i:])
				return out.Bytes()
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
			keyContent := input[i+1 : endQuote]

			// Check for escapes in the key
			hasEscape := bytes.IndexByte(keyContent, '\\') != -1

			var sensitive bool
			if hasEscape {
				// Unescape key to check sensitivity
				// We need to include quotes for json.Unmarshal
				var keyStr string
				if err := json.Unmarshal(input[i:endQuote+1], &keyStr); err == nil {
					// Check the unescaped key string
					// We convert string to bytes to use existing helpers if needed, or pass string.
					// scanForSensitiveKeys expects bytes.
					sensitive = scanForSensitiveKeys([]byte(keyStr), false, false)
				} else {
					// Failed to unmarshal key, treat as not sensitive or fallback?
					// If key is invalid JSON, we probably shouldn't be here or it's not a valid key.
					sensitive = false
				}
			} else {
				// Use scanForSensitiveKeys to check if the key matches any sensitive pattern.
				// scanForSensitiveKeys checks for substrings and handles case folding as implemented in its logic.
				sensitive = scanForSensitiveKeys(keyContent, false, false)
			}

			if sensitive {
				// Write key and colon
				out.Write(input[i : colonPos+1])

				// Identify value
				valStart := colonPos + 1
				// Skip whitespace and write it to output
				for valStart < n {
					c := input[valStart]
					if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
						out.WriteByte(c) // Preserve whitespace
						valStart++
						continue
					}
					break
				}

				if valStart >= n {
					break
				}

				// Determine value end
				valEnd := skipJSONValue(input, valStart)

				// Redact
				out.Write([]byte(redactedValue)) // redactedValue is "[REDACTED]" (json.RawMessage)

				// Advance i to valEnd
				i = valEnd
			} else {
				// Not sensitive, copy key and let loop continue
				out.Write(input[i : endQuote+1])
				i = endQuote + 1
			}
		} else {
			// Not a key (just a string value), copy it
			out.Write(input[i : endQuote+1])
			i = endQuote + 1
		}
	}

	return out.Bytes()
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
		backslashes := 0
		for j := absQ - 1; j >= scanStart; j-- {
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
