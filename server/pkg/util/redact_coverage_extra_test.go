// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactSlice_Coverage(t *testing.T) {
	// Cover the unused redactSlice function

	// Case 1: Slice with no secrets
	s1 := []interface{}{"a", "b", 1}
	res1 := redactSlice(s1)
	// Since no change, it might return nil or same slice depending on implementation details
	// actually redactSlice checks changed. If false, it creates deep copy?
	// Implementation:
	// res, changed := redactSliceMaybe(s)
	// if changed { return res }
	// newSlice := make...
	// ... copy ...
	// return newSlice

	assert.Equal(t, s1, res1)

	// Case 2: Slice with secrets in map
	s2 := []interface{}{
		map[string]interface{}{
			"secret": "value",
			"safe":   "value",
		},
	}
	res2 := redactSlice(s2)
	assert.Equal(t, "[REDACTED]", res2[0].(map[string]interface{})["secret"])
	assert.Equal(t, "value", res2[0].(map[string]interface{})["safe"])

	// Case 3: Slice with nested slice with secrets
	s3 := []interface{}{
		[]interface{}{
			map[string]interface{}{
				"password": "123",
			},
		},
	}
	res3 := redactSlice(s3)
	nested := res3[0].([]interface{})
	nestedMap := nested[0].(map[string]interface{})
	assert.Equal(t, "[REDACTED]", nestedMap["password"])
}

func TestMatchFoldRest_Coverage(t *testing.T) {
	// matchFoldRest is internal, but we can exercise it via IsSensitiveKey.
	// We need to pick keys from the sensitiveKeys list in redact.go
	// "api_key", "apikey", "token", "secret", "password", "passwd", "credential", "auth", "private_key", ...

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// 1. len(s) < len(key)
		// "api_key" is length 7. "api_ke" is 6.
		// IsSensitiveKey("api_ke") -> checkPotentialMatch("api_ke") -> matchFoldRest("api_ke", "api_key")
		// However, checkPotentialMatch might iterate other keys.
		// "auth" is length 4. "aut" is 3.
		// We need to ensure we hit the len check for a specific key where matching started.
		// "api_key" starts with 'a'. "api_ke" starts with 'a'.
		// sensitiveKeyGroups['a'] includes "api_key", "apikey", "auth", "authorization", "api_keys", "apikeys", "authentication", "authenticator".
		// "api_ke" will try to match "api_key". "api_ke" is shorter.
		{"ShortInput", "api_ke", false},

		// 2. c == k (Exact match)
		{"ExactMatch", "api_key", true},

		// 3. k in [a-z] AND (c | 0x20) == k (Case insensitive match)
		{"CaseInsensitiveMatch", "API_KEY", true},

		// 4. k in [a-z] AND (c | 0x20) != k (Mismatch on letter)
		// "api_key": match 'a', 'p', 'i', '_'. Then 'k' vs 'z'.
		{"MismatchLetter", "api_zey", false},

		// 5. k NOT in [a-z] AND c != k (Mismatch on non-letter)
		// "api_key" has '_'. Match 'a', 'p', 'i'. Next k='_'. Input has '-'.
		// c='-', k='_'. c!=k. k is not [a-z]. Should return false.
		{"MismatchNonLetter", "api-key", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSensitiveKey(tt.input); got != tt.expected {
				t.Errorf("IsSensitiveKey(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
