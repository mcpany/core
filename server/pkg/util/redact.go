// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive

import (
	"bytes"
	"encoding/json"
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
	hasSensitiveKey := false
	for _, k := range sensitiveKeysBytes {
		if bytes.Contains(input, k) {
			hasSensitiveKey = true
			break
		}
	}
	if !hasSensitiveKey {
		return input
	}

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

	// Brute force case-insensitive search
	end := len(s) - len(substr)
	for i := 0; i <= end; i++ {
		match := true
		for j := 0; j < len(substr); j++ {
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
	return false
}
