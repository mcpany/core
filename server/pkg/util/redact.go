// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import (
	"bytes"
	"encoding/json"
	"unsafe"
)

const redactedPlaceholder = "[REDACTED]"

var (
	sensitiveKeysBytes [][]byte
	redactedValue      json.RawMessage

	// sensitiveStartChars contains the lowercase starting characters of all sensitive keys.
	// Used for optimized scanning.
	sensitiveStartChars []byte

	// sensitiveKeyGroups maps a starting character (lowercase) to the list of sensitive keys starting with it.
	// Optimization: Use array instead of map for faster lookup.
	sensitiveKeyGroups [256][][]byte

	// sensitiveNextCharMask maps a starting character to a bitmask of allowed second characters.
	// Bit 0 = 'a', Bit 1 = 'b', etc.
	// Used to quickly filter out false positives based on the second character.
	sensitiveNextCharMask [256]uint32
)

func init() {
	for _, k := range sensitiveKeys {
		kb := []byte(k)
		sensitiveKeysBytes = append(sensitiveKeysBytes, kb)

		if len(kb) > 0 {
			first := kb[0] // sensitiveKeys are lowercase
			if len(sensitiveKeyGroups[first]) == 0 {
				sensitiveStartChars = append(sensitiveStartChars, first)
			}
			sensitiveKeyGroups[first] = append(sensitiveKeyGroups[first], kb)
		}
	}

	// Build next char masks
	for start, keys := range sensitiveKeyGroups {
		if len(keys) == 0 {
			continue
		}
		var mask uint32
		for _, k := range keys {
			if len(k) > 1 {
				second := k[1] // k is lowercase
				if second >= 'a' && second <= 'z' {
					mask |= 1 << (second - 'a')
				}
			}
		}
		sensitiveNextCharMask[start] = mask
	}

	// Pre-marshal the redacted placeholder to ensure valid JSON and avoid repeated work.
	b, _ := json.Marshal(redactedPlaceholder)
	redactedValue = json.RawMessage(b)
}

// RedactJSON parses a JSON byte slice and redacts sensitive keys.
// If the input is not valid JSON object or array, it returns the input as is.
func RedactJSON(input []byte) []byte {
	// Optimization: Check if any sensitive key is present in the input.
	// If not, we can skip the expensive unmarshal/marshal process.
	if !scanForSensitiveKeys(input, false) {
		return input
	}

	// Use fast zero-allocation redaction path
	// This avoids expensive json.Unmarshal/Marshal for large payloads
	return redactJSONFast(input)
}

// RedactMap recursively redacts sensitive keys in a map.
// Note: This function creates a deep copy of the map with redacted values.
func RedactMap(m map[string]interface{}) map[string]interface{} {
	newMap := make(map[string]interface{})
	for k, v := range m {
		if IsSensitiveKey(k) {
			newMap[k] = redactedPlaceholder
		} else {
			if nestedMap, ok := v.(map[string]interface{}); ok {
				newMap[k] = RedactMap(nestedMap)
			} else if nestedSlice, ok := v.([]interface{}); ok {
				newMap[k] = redactSlice(nestedSlice)
			} else {
				newMap[k] = v
			}
		}
	}
	return newMap
}

func redactSlice(s []interface{}) []interface{} {
	newSlice := make([]interface{}, len(s))
	for i, v := range s {
		if nestedMap, ok := v.(map[string]interface{}); ok {
			newSlice[i] = RedactMap(nestedMap)
		} else if nestedSlice, ok := v.([]interface{}); ok {
			newSlice[i] = redactSlice(nestedSlice)
		} else {
			newSlice[i] = v
		}
	}
	return newSlice
}

// redactMapRaw operates on map[string]json.RawMessage to avoid full decoding.
// It returns true if any modification was made.
func redactMapRaw(m map[string]json.RawMessage) bool {
	changed := false
	for k, v := range m {
		if IsSensitiveKey(k) {
			m[k] = redactedValue
			changed = true
		} else if len(v) > 0 {
			// Check if we need to recurse
			// Only recurse if the value looks like an object or array
			// Optimization: json.RawMessage from Unmarshal already contains the value bytes without
			// leading/trailing whitespace (unless the value itself is a string with spaces).
			// We can check v[0] directly instead of scanning with bytes.TrimSpace(v) which is O(N).
			switch v[0] {
			case '{', '[':
				if !shouldScanRaw(v) {
					continue
				}

				if v[0] == '{' {
					var nested map[string]json.RawMessage
					if err := json.Unmarshal(v, &nested); err == nil {
						if redactMapRaw(nested) {
							if result, err := json.Marshal(nested); err == nil {
								m[k] = result
								changed = true
							}
						}
					}
				} else {
					var nested []json.RawMessage
					if err := json.Unmarshal(v, &nested); err == nil {
						if redactSliceRaw(nested) {
							if result, err := json.Marshal(nested); err == nil {
								m[k] = result
								changed = true
							}
						}
					}
				}
			}
		}
	}
	return changed
}

// redactSliceRaw operates on []json.RawMessage to avoid full decoding.
// It returns true if any modification was made.
func redactSliceRaw(s []json.RawMessage) bool {
	changed := false
	for i, v := range s {
		if len(v) > 0 {
			switch v[0] {
			case '{', '[':
				if !shouldScanRaw(v) {
					continue
				}

				if v[0] == '{' {
					var nested map[string]json.RawMessage
					if err := json.Unmarshal(v, &nested); err == nil {
						if redactMapRaw(nested) {
							if result, err := json.Marshal(nested); err == nil {
								s[i] = result
								changed = true
							}
						}
					}
				} else {
					var nested []json.RawMessage
					if err := json.Unmarshal(v, &nested); err == nil {
						if redactSliceRaw(nested) {
							if result, err := json.Marshal(nested); err == nil {
								s[i] = result
								changed = true
							}
						}
					}
				}
			}
		}
	}
	return changed
}

// shouldScanRaw returns true if the raw JSON value might contain sensitive keys.
// It is a heuristic to avoid expensive unmarshaling for clean values.
func shouldScanRaw(v []byte) bool {
	// Optimization: Check if any sensitive key is present in the value.
	// If not, we can skip the expensive unmarshal/marshal process.
	// Security Note: We also check for backslashes to ensure no escaped keys are present.
	// If backslashes are present, we conservatively fallback to unmarshaling.
	return scanForSensitiveKeys(v, true)
}

// bytesContainsFold2 is a proposed optimization that we might use in the future.
// Ideally, we want a function that can search for multiple keys at once (Aho-Corasick),
// but for now we stick to optimizing the single key search or the calling pattern.

// sensitiveKeys is a list of substrings that suggest a key contains sensitive information.
// Note: Shorter keys that are substrings of longer keys (e.g. "token" vs "access_token") cover the longer cases,
// so we only include the shorter ones to optimize performance.
var sensitiveKeys = []string{
	"api_key", "apikey", "token", "secret", "password", "passwd", "credential", "auth", "private_key",
	"authorization", "proxy-authorization", "cookie", "set-cookie", "x-api-key",
}

// IsSensitiveKey checks if a key name suggests it contains sensitive information.
func IsSensitiveKey(key string) bool {
	// Use the optimized byte-based scanner for keys as well.
	// Avoid allocation using zero-copy conversion.
	//nolint:gosec // Zero-copy conversion for optimization
	return scanForSensitiveKeys(unsafe.Slice(unsafe.StringData(key), len(key)), false)
}

// scanForSensitiveKeys checks if input contains any sensitive key.
// If checkEscape is true, it also returns true if a backslash is found.
// This function replaces the old linear scan (O(N*M)) with a more optimized scan
// that uses SIMD-accelerated IndexByte for grouped start characters.
func scanForSensitiveKeys(input []byte, checkEscape bool) bool {
	if checkEscape {
		if bytes.IndexByte(input, '\\') != -1 {
			return true
		}
	}

	for _, startChar := range sensitiveStartChars {
		keys := sensitiveKeyGroups[startChar]
		// startChar is lowercase. We need to check for uppercase too.
		// Optimized loop: skip directly to the next occurrence of startChar or startChar-32
		upperChar := startChar - 32

		offset := 0
		for offset < len(input) {
			slice := input[offset:]

			// Find first occurrence of startChar or upperChar
			idxL := bytes.IndexByte(slice, startChar)
			idxU := bytes.IndexByte(slice, upperChar)

			var idx int
			if idxL == -1 && idxU == -1 {
				break // No more matches for this char
			}

			switch {
			case idxL == -1:
				idx = idxU
			case idxU == -1:
				idx = idxL
			default:
				if idxL < idxU {
					idx = idxL
				} else {
					idx = idxU
				}
			}
			// Found candidate start at offset + idx
			matchStart := offset + idx

			// Optimization: Check second character
			if matchStart+1 < len(input) {
				second := input[matchStart+1] | 0x20
				if second >= 'a' && second <= 'z' {
					mask := sensitiveNextCharMask[startChar]
					if (mask & (1 << (second - 'a'))) == 0 {
						// Second character doesn't match any key in this group
						offset = matchStart + 1
						continue
					}
				}
			} else {
				// Not enough bytes for any key
				break
			}

			// Check all keys in this group against input starting at matchStart
			for _, key := range keys {
				if matchFoldRest(input[matchStart:], key) {
					endIdx := matchStart + len(key)
					// Check boundary: if the next character is a lowercase letter,
					// it's likely a continuation of a word (e.g. "auth" in "author"), so we skip it.
					// We allow uppercase letters (CamelCase) and other characters (snake_case, end of string).
					if endIdx < len(input) {
						next := input[endIdx]
						if next >= 'a' && next <= 'z' {
							continue
						}
						// Special handling for uppercase keys (e.g. "AUTH" in "AUTHORITY")
						// If the matched key was uppercase, and the next char is uppercase, it's a continuation.
						// However, if the matched key was lowercase (e.g. "auth" in "authToken"), it's CamelCase (boundary).
						if next >= 'A' && next <= 'Z' {
							// Check if the matched key was uppercase.
							// We know input[matchStart] matched the key start.
							// If input[matchStart] is uppercase, assume the whole key match was uppercase (or case-insensitive matching logic holds).
							firstChar := input[matchStart]
							if firstChar >= 'A' && firstChar <= 'Z' {
								continue
							}
						}
					}

					// Optimization: check if it looks like a key (followed by quote and colon)
					// This reduces false positives when sensitive words appear in values.
					if isKey(input, endIdx) {
						return true
					}
				}
			}

			// Move past this match
			offset = matchStart + 1
		}
	}
	return false
}

// matchFoldRest checks if s starts with key (case-insensitive).
// It assumes the first character already matched (case-insensitive).
func matchFoldRest(s, key []byte) bool {
	if len(s) < len(key) {
		return false
	}
	// Skip index 0 as it was already matched
	for i := 1; i < len(key); i++ {
		c := s[i]
		k := key[i] // k is lowercase
		if c != k {
			// Check if c is the uppercase version of k
			if c < 'A' || c > 'Z' || c+32 != k {
				return false
			}
		}
	}
	return true
}

// isKey checks if the string segment starting at startOffset is followed by a closing quote and a colon,
// indicating it is likely a JSON key.
// It conservatively returns true if it hits the scan limit or encounters ambiguity (like escapes).
func isKey(input []byte, startOffset int) bool {
	// Optimization: limit the scan to avoid O(N^2) behavior in pathological cases.
	const maxScan = 256
	endLimit := startOffset + maxScan
	if endLimit > len(input) {
		endLimit = len(input)
	}

	for i := startOffset; i < endLimit; i++ {
		b := input[i]
		if b == '\\' {
			// Found escape sequence. To be safe/conservative, assume it might be a key.
			// properly handling escapes would require tracking state which is complex.
			return true
		}
		if b == '"' {
			// Found potential closing quote.
			// Check if followed by colon (ignoring whitespace).
			for j := i + 1; j < len(input); j++ {
				c := input[j]
				if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
					continue
				}
				return c == ':'
			}
			return false // EOF before finding colon
		}
	}
	// Limit reached or EOF without finding quote.
	// Conservative: assume it might be a key.
	return true
}
