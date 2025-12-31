// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"bytes"
	"encoding/json"
)

const redactedPlaceholder = "[REDACTED]"

var (
	sensitiveKeysBytes [][]byte
	redactedValue      json.RawMessage
)

func init() {
	for _, k := range sensitiveKeys {
		sensitiveKeysBytes = append(sensitiveKeysBytes, []byte(k))
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
	hasSensitiveKey := false
	for _, k := range sensitiveKeysBytes {
		if bytesContainsFold(input, k) {
			hasSensitiveKey = true
			break
		}
	}
	if !hasSensitiveKey {
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
	// Note: This optimization assumes that if a sensitive key is present in the nested JSON,
	// the key string (e.g. "password") must be present in the raw bytes.
	// Security Note: We also check for backslashes to ensure no escaped keys are present.
	// If backslashes are present, we conservatively fallback to unmarshaling.
	if bytes.IndexByte(v, '\\') != -1 {
		return true
	}
	for _, sk := range sensitiveKeysBytes {
		if bytesContainsFold(v, sk) {
			return true
		}
	}
	return false
}

// bytesContainsFold2 is a proposed optimization that we might use in the future.
// Ideally, we want a function that can search for multiple keys at once (Aho-Corasick),
// but for now we stick to optimizing the single key search or the calling pattern.

// sensitiveKeys is a list of substrings that suggest a key contains sensitive information.
// Note: Shorter keys that are substrings of longer keys (e.g. "token" vs "access_token") cover the longer cases,
// so we only include the shorter ones to optimize performance.
var sensitiveKeys = []string{"api_key", "apikey", "token", "secret", "password", "passwd", "credential", "auth", "private_key"}

// IsSensitiveKey checks if a key name suggests it contains sensitive information.
func IsSensitiveKey(key string) bool {
	for _, s := range sensitiveKeys {
		if containsFold(key, s) {
			return true
		}
	}
	return false
}

// bytesContainsFold reports whether substr is within s, interpreting ASCII characters case-insensitively.
// substr must be lower-case.
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

		// If firstLower is not found, we only need to look for firstUpper.
		if idxL == -1 {
			idx := bytes.IndexByte(slice, firstUpper)
			if idx == -1 {
				return false
			}
			offset += idx
			goto CHECK
		}

		// If firstLower IS found, we check if firstUpper appears *before* it.
		// Optimization: If firstUpper == firstLower (not possible for letters), handled implicitly.
		// We use a small optimization: usually we are looking for lowercase keys in lowercase text.
		// So idxL is likely the one. We only check for Upper if it's closer.
		{
			idxU := bytes.IndexByte(slice[:idxL], firstUpper)
			if idxU != -1 {
				offset += idxU
			} else {
				offset += idxL
			}
		}

	CHECK:

		// Check if we are past the maxSearch limit.
		// Note: offset was updated to point to the potential match start.
		if offset > maxSearch {
			return false
		}

		// Optimization: Check the last character of the substring first.
		// This is a simplified Boyer-Moore idea. If the last character doesn't match,
		// we can fail immediately.
		lastChar := substr[len(substr)-1]
		lastCharS := s[offset+len(substr)-1]
		if lastCharS >= 'A' && lastCharS <= 'Z' {
			lastCharS += 32
		}
		if lastCharS != lastChar {
			offset++
			continue
		}

		// Check the rest of the string
		match := true
		matchStart := offset
		for j := 1; j < len(substr); j++ {
			cc := s[matchStart+j]
			// Optimized ASCII lowercase check:
			// If we know sensitive keys are only letters/digits, we can use a faster check.
			// But for general correctness, we stick to [A-Z] -> +32.
			// We can use a branchless trick: cc | 0x20.
			// However, this affects non-alpha characters too (e.g. '[' -> '{').
			// Since sensitiveKeys contains only [a-z_], we can verify if this optimization is safe.
			// Currently sensitiveKeys are: "api_key", "apikey", "access_token", ...
			// They are all [a-z_].
			// If substr[j] is '_', then s[matchStart+j] must be '_'. '_' is 0x5F.
			// '_' | 0x20 is 0x7F (DEL). So '_' != ('_' | 0x20).
			// So | 0x20 is ONLY safe if we know s[matchStart+j] is a letter.
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
		offset++
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
