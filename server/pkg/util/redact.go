// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

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

	// Optimization: Determine if input is an object or array to avoid unnecessary Unmarshal calls.
	// We only need to check the first non-whitespace character.
	// bytes.TrimSpace scans the whole slice (left and right), which is O(N). We only need O(Whitespace).
	var firstByte byte
	for i := 0; i < len(input); i++ {
		b := input[i]
		if b != ' ' && b != '\t' && b != '\n' && b != '\r' {
			firstByte = b
			break
		}
	}

	switch firstByte {
	case '{':
		var m map[string]json.RawMessage
		if err := json.Unmarshal(input, &m); err == nil {
			if redactMapRaw(m) {
				b, _ := json.Marshal(m)
				return b
			}
		}
	case '[':
		var s []json.RawMessage
		if err := json.Unmarshal(input, &s); err == nil {
			if redactSliceRaw(s) {
				b, _ := json.Marshal(s)
				return b
			}
		}
	default:
		// Try both if we can't determine (e.g. unknown format)
		var m map[string]json.RawMessage
		if err := json.Unmarshal(input, &m); err == nil {
			if redactMapRaw(m) {
				b, _ := json.Marshal(m)
				return b
			}
		}
		var s []json.RawMessage
		if err := json.Unmarshal(input, &s); err == nil {
			if redactSliceRaw(s) {
				b, _ := json.Marshal(s)
				return b
			}
		}
	}

	return input
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
		} else {
			// Check if we need to recurse
			// Only recurse if the value looks like an object or array
			trimmed := bytes.TrimSpace(v)
			if len(trimmed) > 0 {
				switch trimmed[0] {
				case '{', '[':
					if !shouldScanRaw(v) {
						continue
					}

					if trimmed[0] == '{' {
						var nested map[string]json.RawMessage
						if err := json.Unmarshal(trimmed, &nested); err == nil {
							if redactMapRaw(nested) {
								if result, err := json.Marshal(nested); err == nil {
									m[k] = result
									changed = true
								}
							}
						}
					} else {
						var nested []json.RawMessage
						if err := json.Unmarshal(trimmed, &nested); err == nil {
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
	}
	return changed
}

// redactSliceRaw operates on []json.RawMessage to avoid full decoding.
// It returns true if any modification was made.
func redactSliceRaw(s []json.RawMessage) bool {
	changed := false
	for i, v := range s {
		trimmed := bytes.TrimSpace(v)
		if len(trimmed) > 0 {
			switch trimmed[0] {
			case '{', '[':
				if !shouldScanRaw(v) {
					continue
				}

				if trimmed[0] == '{' {
					var nested map[string]json.RawMessage
					if err := json.Unmarshal(trimmed, &nested); err == nil {
						if redactMapRaw(nested) {
							if result, err := json.Marshal(nested); err == nil {
								s[i] = result
								changed = true
							}
						}
					}
				} else {
					var nested []json.RawMessage
					if err := json.Unmarshal(trimmed, &nested); err == nil {
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

// sensitiveKeys is a list of substrings that suggest a key contains sensitive information.
var sensitiveKeys = []string{"api_key", "apikey", "access_token", "token", "secret", "password", "passwd", "credential", "auth", "private_key", "client_secret"}

// IsSensitiveKey checks if a key name suggests it contains sensitive information.
func IsSensitiveKey(key string) bool {
	// Use the optimized byte-based scanner for keys as well.
	// Avoid allocation using zero-copy conversion.
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
			} else if idxL == -1 {
				idx = idxU
			} else if idxU == -1 {
				idx = idxL
			} else {
				if idxL < idxU {
					idx = idxL
				} else {
					idx = idxU
				}
			}

			// Found candidate start at offset + idx
			matchStart := offset + idx

			// Check all keys in this group against input starting at matchStart
			for _, key := range keys {
				if matchFoldRest(input[matchStart:], key) {
					return true
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

// bytesContainsFold reports whether substr is within s, interpreting ASCII characters case-insensitively.
// substr must be lower-case.
// Deprecated: Use scanForSensitiveKeys instead for batched checks.
func bytesContainsFold(s, substr []byte) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}

	firstLower := substr[0]
	firstUpper := firstLower - 32 // sensitiveKeys are all lowercase ASCII

	offset := 0
	// We only need to search up to this point
	maxSearch := len(s) - len(substr)

	for offset <= maxSearch {
		// Optimization: Use IndexByte which is assembly optimized (SIMD)
		// Scan for either lowercase or uppercase first char
		slice := s[offset:]
		idxL := bytes.IndexByte(slice, firstLower)

		var idx int
		if idxL == -1 {
			// Lowercase not found, must search for uppercase
			idx = bytes.IndexByte(slice, firstUpper)
			if idx == -1 {
				return false
			}
		} else {
			// Lowercase found at idxL.
			// Check if uppercase appears *before* idxL.
			// We only need to search in slice[:idxL] because if we find it there, it's the winner.
			// If we don't find it there, idxL is the winner (even if uppercase appears later).
			idxU := bytes.IndexByte(slice[:idxL], firstUpper)
			if idxU != -1 {
				idx = idxU
			} else {
				idx = idxL
			}
		}

		// Found a match at offset + idx
		// Check if it's within bounds (IndexByte searches the whole slice, but we only care if it fits)
		if offset+idx > maxSearch {
			return false
		}

		// Check the rest of the string
		match := true
		matchStart := offset + idx
		for j := 1; j < len(substr); j++ {
			cc := s[matchStart+j]
			if cc >= 'A' && cc <= 'Z' {
				cc += 32
			}
			if cc != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}

		// Move past this match
		offset += idx + 1
	}
	return false
}

// containsFold reports whether substr is within s, interpreting ASCII characters case-insensitively.
func containsFold(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	if len(substr) == 0 {
		return true
	}

	// Optimized case-insensitive search
	// We first check the first character to avoid setting up the inner loop for mismatches.
	// Since sensitiveKeys are all lowercase, we can safely assume substr[0] is lowercase.
	first := substr[0]
	maxLen := len(s) - len(substr)

	for i := 0; i <= maxLen; i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32 // to lower
		}
		if c == first {
			// First character matches, check the rest
			match := true
			for j := 1; j < len(substr); j++ {
				charS := s[i+j]
				if charS >= 'A' && charS <= 'Z' {
					charS += 32 // to lower
				}
				if charS != substr[j] {
					match = false
					break
				}
			}
			if match {
				return true
			}
		}
	}
	return false
}
