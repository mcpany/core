package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import (
	"bytes"
	"encoding/json"
	"math"
)

// maxUnescapeLimit is the maximum size of a key that we will attempt to unescape
// using json.Unmarshal. Keys larger than this will use a streaming scanner
// to avoid excessive allocation.
// Variable for testing purposes.
var maxUnescapeLimit = 1024 * 1024

// unescapeStackLimit is the size of the stack buffer for unescaping keys.
const unescapeStackLimit = 256

// isJSONWhitespace is a lookup table for fast whitespace checking.
// It avoids multiple comparisons in the hot path.
var isJSONWhitespace [256]bool

// isNumberDelimiter is a lookup table for fast number delimiter checking.
var isNumberDelimiter [256]bool

func init() {
	isJSONWhitespace[' '] = true
	isJSONWhitespace['\t'] = true
	isJSONWhitespace['\n'] = true
	isJSONWhitespace['\r'] = true

	// Number delimiters: whitespace, ',', '}', ']'
	// We also treat '/' as delimiter because of comments check in skipNumber
	isNumberDelimiter[' '] = true
	isNumberDelimiter['\t'] = true
	isNumberDelimiter['\n'] = true
	isNumberDelimiter['\r'] = true
	isNumberDelimiter[','] = true
	isNumberDelimiter['}'] = true
	isNumberDelimiter[']'] = true
	isNumberDelimiter['/'] = true
	// Quotes and colons should also stop number scanning to avoid consuming keys in invalid JSON
	isNumberDelimiter['"'] = true
	isNumberDelimiter[':'] = true
}

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
					// Note: skipWhitespaceAndComments handles the full comment and any trailing whitespace.
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

			// Optimization: fast path if previous char is not backslash
			if input[absQ-1] != '\\' {
				endQuote = absQ
				break
			}

			// Count backslashes before absQ
			backslashes := 1 // we already saw one backslash at input[absQ-1]
			for j := absQ - 2; j >= scanStart; j-- {
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
		// It is a key if it is followed by a colon (ignoring whitespace and comments)
		isKey := false
		colonPos := -1

		// Scan after endQuote+1 for colon
		// ⚡ Bolt Optimization: Use optimized skipWhitespaceAndComments with lookup table
		// This replaces the manual loop and correctly handles comments between key and colon.
		j := skipWhitespaceAndComments(input, endQuote+1)
		if j < n && input[j] == ':' {
			isKey = true
			colonPos = j
		}

		if isKey {
			// Check if key is sensitive
			keyContent := input[quotePos+1 : endQuote]
			if isKeySensitive(keyContent) {
				// Identify value start
				valStart := skipWhitespaceAndComments(input, colonPos+1)

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
						// Use int64 for calculation to avoid overflow before comparison
						targetCap := int64(len(input)) + int64(extra)
						// Check against max int to be safe on 32-bit systems, though slices are limited by arch
						if targetCap > math.MaxInt || targetCap < int64(len(input)) {
							targetCap = int64(len(input))
						}
						out = make([]byte, 0, int(targetCap))
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

// skipWhitespaceAndComments skips whitespace and JSON comments (// and /* */).
func skipWhitespaceAndComments(input []byte, start int) int {
	n := len(input)
	i := start
	for i < n {
		c := input[i]
		// ⚡ Bolt Optimization: Fast path for whitespace using lookup table
		if isJSONWhitespace[c] {
			i++
			continue
		}
		// Check for comments (// or /*)
		if c == '/' && i+1 < n {
			next := input[i+1]
			if next == '/' {
				// Line comment: skip until newline
				i += 2
				for i < n {
					if input[i] == '\n' || input[i] == '\r' {
						break
					}
					i++
				}
				continue
			} else if next == '*' {
				// Block comment: skip until */
				i += 2
				for i < n {
					if input[i] == '*' && i+1 < n && input[i+1] == '/' {
						i += 2
						break
					}
					i++
				}
				continue
			}
		}
		break
	}
	return i
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
	n := len(input)
	for i < n {
		c := input[i]
		switch c {
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
		case '/':
			// Check for comments
			if i+1 < n {
				next := input[i+1]
				if next == '/' {
					// Line comment
					i += 2
					for i < n {
						if input[i] == '\n' || input[i] == '\r' {
							break
						}
						i++
					}
					continue
				} else if next == '*' {
					// Block comment
					i += 2
					for i < n {
						if input[i] == '*' && i+1 < n && input[i+1] == '/' {
							i += 2
							break
						}
						i++
					}
					continue
				}
			}
		}
		i++
	}
	return n
}

func skipArray(input []byte, start int) int {
	// Array starts at start, which is '['
	depth := 1
	i := start + 1
	n := len(input)
	for i < n {
		c := input[i]
		switch c {
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
		case '/':
			// Check for comments
			if i+1 < n {
				next := input[i+1]
				if next == '/' {
					// Line comment
					i += 2
					for i < n {
						if input[i] == '\n' || input[i] == '\r' {
							break
						}
						i++
					}
					continue
				} else if next == '*' {
					// Block comment
					i += 2
					for i < n {
						if input[i] == '*' && i+1 < n && input[i+1] == '/' {
							i += 2
							break
						}
						i++
					}
					continue
				}
			}
		}
		i++
	}
	return n
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

func isKeySensitive(keyContent []byte) bool {
	var keyToCheck []byte
	var sensitive bool

	// Stack buffer for unescaping small keys
	var stackBuf [unescapeStackLimit]byte

	if bytes.Contains(keyContent, []byte{'\\'}) {
		// Key contains escape sequences, unescape it to check for sensitivity.
		// This is slower but safer.

		switch {
		case len(keyContent) > maxUnescapeLimit:
			// Use streaming/chunked scan for huge keys to avoid allocation
			if scanEscapedKeyForSensitive(keyContent) {
				sensitive = true
			} else {
				keyToCheck = keyContent
			}
		case len(keyContent) < unescapeStackLimit:
			// Optimization: use stack buffer for small keys
			if unescaped, ok := unescapeKeySmall(keyContent, stackBuf[:]); ok {
				keyToCheck = unescaped
			} else {
				// Fallback if unescape failed (e.g. malformed or weird escape)
				keyToCheck = keyContent
			}
		default:
			// Allocation path for medium keys
			quoted := make([]byte, len(keyContent)+2)
			quoted[0] = '"'
			copy(quoted[1:], keyContent)
			quoted[len(quoted)-1] = '"'

			var unescaped string
			if err := json.Unmarshal(quoted, &unescaped); err == nil {
				keyToCheck = []byte(unescaped)
			} else {
				// Fallback to loose scanning if unmarshal fails (e.g. due to invalid escapes)
				// We use scanEscapedKeyForSensitive which handles invalid escapes gracefully
				// by treating them as literals, which is safer than raw content comparison
				// because scanForSensitiveKeys does not handle escapes.
				if scanEscapedKeyForSensitive(keyContent) {
					sensitive = true
				} else {
					// Still set keyToCheck for the final raw scan, just in case
					keyToCheck = keyContent
				}
			}
		}
	} else {
		keyToCheck = keyContent
	}

	// Optimization: We check the raw key content against sensitive keys.
	// We only unescape if backslashes are detected, avoiding expensive json.Unmarshal calls for the common case.
	// Note: scanForSensitiveKeys (used in the pre-check) also does not handle escapes.
	if !sensitive {
		sensitive = scanForSensitiveKeys(keyToCheck, false)
	}
	return sensitive
}

// unescapeKeySmall unescapes a JSON string into a provided buffer.
// It returns the unescaped slice (backed by buf) and true on success.
// If buf is too small or input is invalid, it returns false.
func unescapeKeySmall(input []byte, buf []byte) ([]byte, bool) {
	bufIdx := 0
	i := 0
	n := len(input)

	for i < n {
		if bufIdx >= len(buf) {
			return nil, false
		}

		c := input[i]
		if c == '\\' {
			i++
			if i >= n {
				return nil, false
			}
			switch input[i] {
			case '"', '\\', '/', '\'':
				c = input[i]
				i++
			case 'b':
				c = '\b'
				i++
			case 'f':
				c = '\f'
				i++
			case 'n':
				c = '\n'
				i++
			case 'r':
				c = '\r'
				i++
			case 't':
				c = '\t'
				i++
			case 'u':
				// \uXXXX
				if i+4 < n {
					val := 0
					valid := true
					for k := 0; k < 4; k++ {
						h := input[i+1+k]
						var d int
						switch {
						case h >= '0' && h <= '9':
							d = int(h - '0')
						case h >= 'a' && h <= 'f':
							d = int(h - 'a' + 10)
						case h >= 'A' && h <= 'F':
							d = int(h - 'A' + 10)
						default:
							valid = false
						}
						if !valid {
							break
						}
						val = (val << 4) | d
					}

					if valid {
						if val <= 127 {
							c = byte(val)
						} else {
							// For non-ASCII, we replace with '?' as scanForSensitiveKeys handles bytes.
							// Sensitive keys are assumed to be ASCII.
							c = '?'
						}
						i += 5
					} else {
						c = 'u'
						i++
					}
				} else {
					c = 'u'
					i++
				}
			default:
				// Unknown escape
				c = input[i]
				i++
			}
		} else {
			i++
		}
		buf[bufIdx] = c
		bufIdx++
	}

	return buf[:bufIdx], true
}

// scanEscapedKeyForSensitive scans a large escaped key for sensitive words using a fixed-size buffer.
// It returns true if any sensitive word is found.
func scanEscapedKeyForSensitive(keyContent []byte) bool {
	// 4KB buffer
	// We use 4097 to allow appending a dummy character when the buffer is full but
	// the stream continues, to avoid false positive matches at the boundary.
	const bufSize = 4097
	const overlap = 64
	var buf [bufSize]byte
	bufIdx := 0

	i := 0
	n := len(keyContent)

	for i < n {
		// Unescape one character
		c := keyContent[i]
		if c == '\\' {
			i++
			if i >= n {
				break
			}
			switch keyContent[i] {
			case '"', '\\', '/', '\'':
				c = keyContent[i]
				i++
			case 'b':
				c = '\b'
				i++
			case 'f':
				c = '\f'
				i++
			case 'n':
				c = '\n'
				i++
			case 'r':
				c = '\r'
				i++
			case 't':
				c = '\t'
				i++
			case 'u':
				// \uXXXX
				if i+4 < n {
					// Parse 4 hex digits
					val := 0
					valid := true
					for k := 0; k < 4; k++ {
						h := keyContent[i+1+k]
						var d int
						switch {
						case h >= '0' && h <= '9':
							d = int(h - '0')
						case h >= 'a' && h <= 'f':
							d = int(h - 'a' + 10)
						case h >= 'A' && h <= 'F':
							d = int(h - 'A' + 10)
						default:
							valid = false
						}
						if !valid {
							break
						}
						val = (val << 4) | d
					}

					if valid {
						// We have a rune value.
						// Sensitive keys are ASCII. If val > 127, it won't match.
						// We can just append '?' or similar if non-ascii, OR handle it if we want full correctness.
						// scanForSensitiveKeys works on bytes.
						// If val <= 127, we append byte(val).
						if val <= 127 {
							c = byte(val)
						} else {
							// For non-ASCII, we can skip or use a placeholder.
							// Assuming sensitive keys are ASCII only.
							c = '?'
						}
						i += 5 // u + 4 digits
					} else {
						// Invalid hex, treat as 'u'
						c = 'u'
						i++
					}
				} else {
					c = 'u'
					i++
				}
			default:
				// Unknown escape, just output the char
				c = keyContent[i]
				i++
			}
		} else {
			i++
		}

		buf[bufIdx] = c
		bufIdx++

		// Check if we reached the processing limit (bufSize - 1)
		// We leave 1 byte for the potential dummy character.
		if bufIdx == bufSize-1 {
			scanLen := bufIdx
			// If we are not at the end of the stream, append a dummy character
			// to avoid matching a word that is cut off by the buffer boundary.
			if i < n {
				buf[bufIdx] = 'z' // Append dummy lowercase letter
				scanLen++
			}

			if scanForSensitiveKeys(buf[:scanLen], false) {
				return true
			}
			// Shift overlap
			// We shift from the valid data, excluding the dummy char if added.
			copy(buf[0:], buf[bufIdx-overlap:bufIdx])
			bufIdx = overlap
		}
	}

	if bufIdx > 0 {
		if scanForSensitiveKeys(buf[:bufIdx], false) {
			return true
		}
	}

	return false
}

func skipNumber(input []byte, start int) int {
	// Number
	// Scan until delimiter: , } ] or whitespace
	// ⚡ Bolt Optimization: Use lookup table for fast delimiter check
	i := start
	for i < len(input) {
		c := input[i]
		if isNumberDelimiter[c] {
			return i
		}
		i++
	}
	return i
}
