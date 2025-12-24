// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive

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
				} else if trimmed[0] == '[' {
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
	return changed
}

// redactSliceRaw operates on []json.RawMessage to avoid full decoding.
// It returns true if any modification was made.
func redactSliceRaw(s []json.RawMessage) bool {
	changed := false
	for i, v := range s {
		trimmed := bytes.TrimSpace(v)
		if len(trimmed) > 0 {
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
			} else if trimmed[0] == '[' {
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
	return changed
}

// sensitiveKeys is a list of substrings that suggest a key contains sensitive information.
var sensitiveKeys = []string{"api_key", "apikey", "access_token", "token", "secret", "password", "passwd", "credential", "auth", "private_key", "client_secret"}

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

	i := 0
	max := len(s) - len(substr)

	for i <= max {
		c := s[i]
		if c != firstLower && c != firstUpper {
			i++
			continue
		}

		// First character matches, check the rest
		match := true
		for j := 1; j < len(substr); j++ {
			cc := s[i+j]
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
		i++
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
	end := len(s) - len(substr)

	for i := 0; i <= end; i++ {
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
