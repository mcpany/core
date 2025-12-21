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
		redacted := RedactMap(m)
		b, _ := json.Marshal(redacted)
		return b
	}
	var s []interface{}
	if err := json.Unmarshal(input, &s); err == nil {
		redacted := redactSlice(s)
		b, _ := json.Marshal(redacted)
		return b
	}
	return input
}

// RedactMap recursively redacts sensitive keys in a map.
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

// sensitiveKeys is a list of substrings that suggest a key contains sensitive information.
var sensitiveKeys = []string{"api_key", "apikey", "access_token", "token", "secret", "password", "passwd", "credential", "auth", "private_key", "client_secret"}

// IsSensitiveKey checks if a key name suggests it contains sensitive information.
func IsSensitiveKey(key string) bool {
	k := strings.ToLower(key)
	for _, s := range sensitiveKeys {
		if strings.Contains(k, s) {
			return true
		}
	}
	return false
}
