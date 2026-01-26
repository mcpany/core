// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRedactJSON_SessionKeys(t *testing.T) {
	// Tests for session identifiers that should be redacted
	tests := []struct {
		name     string
		input    map[string]interface{}
		checkKey string // key to check if value is redacted
	}{
		{
			name:     "session_id",
			input:    map[string]interface{}{"session_id": "secret-session-id"},
			checkKey: "session_id",
		},
		{
			name:     "sid",
			input:    map[string]interface{}{"sid": "secret-sid"},
			checkKey: "sid",
		},
		{
			name:     "jsessionid",
			input:    map[string]interface{}{"jsessionid": "secret-jsessionid"},
			checkKey: "jsessionid",
		},
		{
			name:     "sessionid",
			input:    map[string]interface{}{"sessionid": "secret-sessionid"},
			checkKey: "sessionid",
		},
		{
			name:     "access_key",
			input:    map[string]interface{}{"access_key": "secret-access-key"},
			checkKey: "access_key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("Failed to marshal input: %v", err)
			}

			redacted := RedactJSON(jsonBytes)
			redactedStr := string(redacted)

			// The value should be [REDACTED]
			// We check that the original secret value is NOT present
			originalValue := tt.input[tt.checkKey].(string)
			if strings.Contains(redactedStr, originalValue) {
				t.Errorf("Value for key %q was NOT redacted. Output: %s", tt.checkKey, redactedStr)
			}

			// We verify that the key is still present but value is redacted
			expectedSnippet := "\"" + tt.checkKey + "\":\"[REDACTED]\""
			if !strings.Contains(redactedStr, expectedSnippet) {
				t.Errorf("Expected snippet %q not found in output: %s", expectedSnippet, redactedStr)
			}
		})
	}
}

func TestRedactJSON_SessionFalsePositives(t *testing.T) {
	// Tests for keys that contain sensitive substrings but should NOT be redacted
	tests := []struct {
		name     string
		input    map[string]interface{}
		checkKey string
	}{
		{
			name:     "obsidian (contains sid)",
			input:    map[string]interface{}{"obsidian": "safe-value"},
			checkKey: "obsidian",
		},
		{
			name:     "residue (contains sid)",
			input:    map[string]interface{}{"residue": "safe-value"},
			checkKey: "residue",
		},
		{
			name:     "consider (contains sid)",
			input:    map[string]interface{}{"consider": "safe-value"},
			checkKey: "consider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("Failed to marshal input: %v", err)
			}

			redacted := RedactJSON(jsonBytes)
			redactedStr := string(redacted)

			// The value should be VISIBLE
			originalValue := tt.input[tt.checkKey].(string)
			if !strings.Contains(redactedStr, originalValue) {
				t.Errorf("Value for key %q was INCORRECTLY redacted. Output: %s", tt.checkKey, redactedStr)
			}
		})
	}
}
