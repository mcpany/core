// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive

import (
	"bytes"
	"encoding/json"
	"strings"
)

const redactedPlaceholder = "[REDACTED]"

var sensitiveKeysBytes [][]byte

func init() {
	for _, k := range sensitiveKeys {
		sensitiveKeysBytes = append(sensitiveKeysBytes, []byte(k))
	}
}

// RedactJSON parses a JSON byte slice and redacts sensitive keys.
// If the input is not valid JSON object or array, it returns the input as is.
func RedactJSON(input []byte) []byte {
	// Optimization: Check if any sensitive key is present in the input.
	// If not, we can skip the expensive unmarshal/marshal process.
	// We use a case-insensitive check to ensure we don't miss keys like "API_KEY".

	// Create a lowercase version of input for fast searching.
	// This allocates but is faster than repeatedly scanning mixed-case input 11 times.
	// Benchmark shows this is ~20% faster than manual scanning and 50% faster if we could use stack/reused buffer.
	// Since we can't easily reuse buffer here without API change, we accept the allocation.
	// But wait, allocating for LARGE input is bad.

	hasSensitiveKey := false

	// Heuristic: If input is small, alloc is cheap. If large, alloc is expensive.
	if len(input) < 4096 {
		lower := bytes.ToLower(input)
		for _, k := range sensitiveKeysBytes {
			if bytes.Contains(lower, k) {
				hasSensitiveKey = true
				break
			}
		}
	} else {
		// For large inputs, avoid allocation and use the slower in-place check.
		for _, k := range sensitiveKeysBytes {
			if bytesContainsFold(input, k) {
				hasSensitiveKey = true
				break
			}
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
		var m map[string]interface{}
		if err := json.Unmarshal(input, &m); err == nil {
			redactMapInPlace(m)
			b, _ := json.Marshal(m)
			return b
		}
	case '[':
		var s []interface{}
		if err := json.Unmarshal(input, &s); err == nil {
			redactSliceInPlace(s)
			b, _ := json.Marshal(s)
			return b
		}
	default:
		// Try both if we can't determine (e.g. unknown format), though unlikely for valid JSON
		var m map[string]interface{}
		if err := json.Unmarshal(input, &m); err == nil {
			redactMapInPlace(m)
			b, _ := json.Marshal(m)
			return b
		}
		var s []interface{}
		if err := json.Unmarshal(input, &s); err == nil {
			redactSliceInPlace(s)
			b, _ := json.Marshal(s)
			return b
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

func redactMapInPlace(m map[string]interface{}) {
	for k, v := range m {
		if IsSensitiveKey(k) {
			m[k] = redactedPlaceholder
		} else {
			if nestedMap, ok := v.(map[string]interface{}); ok {
				redactMapInPlace(nestedMap)
			} else if nestedSlice, ok := v.([]interface{}); ok {
				redactSliceInPlace(nestedSlice)
			}
		}
	}
}

func redactSliceInPlace(s []interface{}) {
	for _, v := range s {
		if nestedMap, ok := v.(map[string]interface{}); ok {
			redactMapInPlace(nestedMap)
		} else if nestedSlice, ok := v.([]interface{}); ok {
			redactSliceInPlace(nestedSlice)
		}
	}
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

// containsFold reports whether substr is within s, interpreting ASCII characters case-insensitively.
func containsFold(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	if len(substr) == 0 {
		return true
	}

	first := substr[0]
	// first is lowercase
	firstUpper := first
	if first >= 'a' && first <= 'z' {
		firstUpper -= 32
	}

	for i := 0; i <= len(s)-len(substr); {
		// Optimization: Scan for the first character using IndexByte (SIMD optimized)
		idx := strings.IndexByte(s[i:], first)
		idxUpper := strings.IndexByte(s[i:], firstUpper)

		if idx < 0 && idxUpper < 0 {
			return false
		}

		minIdx := idx
		if minIdx < 0 || (idxUpper >= 0 && idxUpper < minIdx) {
			minIdx = idxUpper
		}

		i += minIdx

		// Check the rest of the substring
		if i+len(substr) > len(s) {
			return false
		}

		match := true
		for j := 1; j < len(substr); j++ {
			charS := s[i+j]
			if charS >= 'A' && charS <= 'Z' {
				charS += 32 // to lower
			}
			if charS != substr[j] { // substr is assumed lowercase
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

// bytesContainsFold is the byte-slice equivalent of containsFold.
func bytesContainsFold(s, substr []byte) bool {
	if len(substr) > len(s) {
		return false
	}
	if len(substr) == 0 {
		return true
	}

	first := substr[0]
	// first is lowercase
	firstUpper := first
	if first >= 'a' && first <= 'z' {
		firstUpper -= 32
	}

	for i := 0; i <= len(s)-len(substr); {
		// Optimization: Scan for the first character using IndexByte (SIMD optimized)
		idx := bytes.IndexByte(s[i:], first)
		idxUpper := bytes.IndexByte(s[i:], firstUpper)

		if idx < 0 && idxUpper < 0 {
			return false
		}

		minIdx := idx
		if minIdx < 0 || (idxUpper >= 0 && idxUpper < minIdx) {
			minIdx = idxUpper
		}

		i += minIdx

		// Check the rest of the substring
		if i+len(substr) > len(s) {
			return false
		}

		match := true
		for j := 1; j < len(substr); j++ {
			charS := s[i+j]
			if charS >= 'A' && charS <= 'Z' {
				charS += 32 // to lower
			}
			if charS != substr[j] { // substr is assumed lowercase
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
