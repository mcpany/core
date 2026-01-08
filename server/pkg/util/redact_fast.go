// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import (
	"bytes"
	"unicode/utf8"
)

// redactJSONFast is a zero-allocation (mostly) implementation of RedactJSON.
// It scans the input byte slice and constructs a new slice with redacted values.
// It avoids full JSON parsing.
func redactJSONFast(input []byte) []byte {
	// Pre-allocate result buffer.
	// We use len(input) as a safe estimate.
	// Redaction usually shrinks the output (replacing long values with [REDACTED]),
	// but can expand it if replacing short values (like "") with [REDACTED].
	// append will handle growth if needed.
	out := make([]byte, 0, len(input))

	i := 0
	n := len(input)

	for i < n {
		// Scan for the next quote which might start a key
		nextQuote := bytes.IndexByte(input[i:], '"')
		if nextQuote == -1 {
			// No more strings, just copy the rest
			out = append(out, input[i:]...)
			break
		}
		quotePos := i + nextQuote

		// Copy everything up to the quote
		out = append(out, input[i:quotePos]...)
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
				out = append(out, input[i:]...)
				return out
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
				// Avoid json.Unmarshal for performance
				unquoted := unquoteBytes(input[i : endQuote+1])
				sensitive = scanForSensitiveKeys(unquoted, false)
			} else {
				// Use scanForSensitiveKeys to check if the key matches any sensitive pattern.
				// scanForSensitiveKeys checks for substrings and handles case folding as implemented in its logic.
				// We don't need to validate key context here because we know we are inside a key.
				sensitive = scanForSensitiveKeys(keyContent, false)
			}

			if sensitive {
				// Write key and colon
				out = append(out, input[i:colonPos+1]...)

				// Identify value
				valStart := colonPos + 1
				// Skip whitespace and write it to output
				for valStart < n {
					c := input[valStart]
					if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
						out = append(out, c) // Preserve whitespace
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
				out = append(out, redactedValue...)

				// Advance i to valEnd
				i = valEnd
			} else {
				// Not sensitive, copy key and let loop continue
				out = append(out, input[i:endQuote+1]...)
				i = endQuote + 1
			}
		} else {
			// Not a key (just a string value), copy it
			out = append(out, input[i:endQuote+1]...)
			i = endQuote + 1
		}
	}

	return out
}

// unquoteBytes unquotes a JSON string (with quotes) into a byte slice.
// It handles basic escapes and unicode sequences.
func unquoteBytes(input []byte) []byte {
	// input includes quotes, e.g. "key"
	n := len(input)
	if n < 2 {
		return input // Should not happen for valid quoted string
	}
	// content is input[1 : n-1]
	// Conservative estimate for capacity: len(input)-2
	out := make([]byte, 0, n-2)

	// We iterate manually
	for i := 1; i < n-1; i++ {
		c := input[i]
		if c == '\\' {
			i++
			if i >= n-1 {
				// Trailing backslash?
				out = append(out, '\\')
				break
			}
			switch input[i] {
			case '"', '\\', '/', '\'':
				out = append(out, input[i])
			case 'b':
				out = append(out, '\b')
			case 'f':
				out = append(out, '\f')
			case 'n':
				out = append(out, '\n')
			case 'r':
				out = append(out, '\r')
			case 't':
				out = append(out, '\t')
			case 'u':
				// Parse hex
				if i+4 < n-1 {
					r, ok := decodeUnicode(input[i+1 : i+5])
					if ok {
						out = utf8.AppendRune(out, r)
						i += 4
					} else {
						// Invalid unicode escape, keep literal
						out = append(out, '\\', 'u')
					}
				} else {
					out = append(out, '\\', 'u')
				}
			default:
				// Unknown escape, keep literal
				out = append(out, '\\', input[i])
			}
		} else {
			out = append(out, c)
		}
	}
	return out
}

func decodeUnicode(src []byte) (rune, bool) {
	var r rune
	for i := 0; i < 4; i++ {
		b, ok := hexToByte(src[i])
		if !ok {
			return 0, false
		}
		r = r<<4 | rune(b)
	}
	return r, true
}

func hexToByte(c byte) (byte, bool) {
	if c >= '0' && c <= '9' {
		return c - '0', true
	}
	if c >= 'a' && c <= 'f' {
		return c - 'a' + 10, true
	}
	if c >= 'A' && c <= 'F' {
		return c - 'A' + 10, true
	}
	return 0, false
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
