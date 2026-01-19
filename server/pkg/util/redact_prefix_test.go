// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactJSON_PrefixFalsePositives(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		// Should NOT Redact (False Positives)
		{
			name:     "Suffix Match - betoken",
			input:    `{"betoken": "safe"}`,
			expected: `{"betoken": "safe"}`,
		},
		{
			name:     "Suffix Match - author",
			input:    `{"author": "Shakespeare"}`,
			expected: `{"author": "Shakespeare"}`,
		},
		{
			name:     "Suffix Match - authorized",
			input:    `{"authorized": "true"}`,
			expected: `{"authorized": "true"}`,
		},
		{
			name:     "Middle Match - pretokenized",
			input:    `{"pretokenized": "value"}`,
			expected: `{"pretokenized": "value"}`,
		},
		{
			name:     "Suffix Match - glass_auth", // "auth" is sensitive
			input:    `{"glass_auth": "[REDACTED]"}`, // Wait, underscore IS a separator. So this SHOULD be redacted.
			// Let's check logic. "_" is NOT alphanumeric. So boundary check passes.
			// So "glass_auth" -> contains "auth" preceded by "_". Should match.
			expected: `{"glass_auth": "[REDACTED]"}`,
		},
		{
			name:     "Suffix Match - glassauth", // "auth" is sensitive
			input:    `{"glassauth": "safe"}`,
			expected: `{"glassauth": "safe"}`,
		},

		// Should Redact (True Positives)
		{
			name:     "Exact Match",
			input:    `{"token": "secret"}`,
			expected: `{"token": "[REDACTED]"}`,
		},
		{
			name:     "Underscore Prefix",
			input:    `{"access_token": "secret"}`,
			expected: `{"access_token": "[REDACTED]"}`,
		},
		{
			name:     "Dash Prefix",
			input:    `{"x-api-key": "secret"}`,
			expected: `{"x-api-key": "[REDACTED]"}`,
		},
		{
			name:     "Start of String",
			input:    `{"token_id": "secret"}`,
			expected: `{"token_id": "[REDACTED]"}`,
		},
		{
			name:     "CamelCase Boundary (lower to Upper)",
			input:    `{"authToken": "secret"}`, // "auth" is sensitive.
			// "auth" matches "auth". Next is 'T'.
			// matchFoldRest handles "auth" vs "authToken".
			// But checkPotentialMatch has special logic for CamelCase at the END.
			// If match is "auth", and next is 'T' (upper), it considers it a match (boundary).
			// So "authToken" contains "auth". It IS redacted.
			expected: `{"authToken": "[REDACTED]"}`,
		},
		{
			name:     "CamelCase Boundary (Upper to Upper)",
			input:    `{"AuthToken": "secret"}`,
			expected: `{"AuthToken": "[REDACTED]"}`,
		},
		{
			name:     "Non-Boundary Continuation",
			input:    `{"author": "value"}`, // "auth" matches start of "author". Next is 'o' (lower).
			// checkPotentialMatch logic: if next is lower, continue (skip).
			// So "author" is NOT redacted.
			expected: `{"author": "value"}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := RedactJSON([]byte(tc.input))
			assert.Equal(t, tc.expected, string(result))
		})
	}
}
