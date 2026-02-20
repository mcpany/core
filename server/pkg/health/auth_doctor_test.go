// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckAuth(t *testing.T) {
	t.Run("API Keys", func(t *testing.T) {
		tests := []struct {
			name     string
			key      string
			value    string
			expected CheckResult
		}{
			{
				name:  "Present Long Key",
				key:   "ANTHROPIC_API_KEY",
				value: "sk-ant-1234567890",
				expected: CheckResult{
					Status:  "ok",
					Message: "Present (...7890)",
				},
			},
			{
				name:  "Present Short Key",
				key:   "OPENAI_API_KEY",
				value: "123",
				expected: CheckResult{
					Status:  "ok",
					Message: "Present",
				},
			},
			{
				name:  "Missing Key",
				key:   "GEMINI_API_KEY",
				value: "",
				expected: CheckResult{
					Status:  "missing",
					Message: "Environment variable not set",
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Setenv(tt.key, tt.value)

				results := CheckAuth()
				assert.Equal(t, tt.expected, results[tt.key])
			})
		}
	})

	t.Run("OAuth Configuration", func(t *testing.T) {
		tests := []struct {
			name           string
			provider       string
			clientID       string
			clientSecret   string
			expectedStatus string
			expectedMsg    string
		}{
			{
				name:           "Fully Configured",
				provider:       "GOOGLE",
				clientID:       "google-id",
				clientSecret:   "google-secret",
				expectedStatus: "ok",
				expectedMsg:    "Configured",
			},
			{
				name:           "Partial Config (Only ID)",
				provider:       "GITHUB",
				clientID:       "github-id",
				clientSecret:   "",
				expectedStatus: "warning",
				expectedMsg:    "Partial configuration (missing ID or Secret)",
			},
			{
				name:           "Partial Config (Only Secret)",
				provider:       "GITHUB", // Reusing provider but different test case
				clientID:       "",
				clientSecret:   "github-secret",
				expectedStatus: "warning",
				expectedMsg:    "Partial configuration (missing ID or Secret)",
			},
			{
				name:           "Not Configured",
				provider:       "GOOGLE",
				clientID:       "",
				clientSecret:   "",
				expectedStatus: "info",
				expectedMsg:    "Not configured",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Ensure clean slate by setting to empty first (t.Setenv restores original)
				t.Setenv(tt.provider+"_CLIENT_ID", "")
				t.Setenv(tt.provider+"_CLIENT_SECRET", "")

				if tt.clientID != "" {
					t.Setenv(tt.provider+"_CLIENT_ID", tt.clientID)
				}
				if tt.clientSecret != "" {
					t.Setenv(tt.provider+"_CLIENT_SECRET", tt.clientSecret)
				}

				results := CheckAuth()
				key := "oauth_" + tt.provider
				result, ok := results[key]
				assert.True(t, ok, "Expected result for %s", key)
				assert.Equal(t, tt.expectedStatus, result.Status)
				assert.Equal(t, tt.expectedMsg, result.Message)
			})
		}
	})

	t.Run("Key Masking Logic", func(t *testing.T) {
		// Test specifically the logic for masking
		key := "ANTHROPIC_API_KEY"

		t.Run("Length > 4", func(t *testing.T) {
			t.Setenv(key, "12345")
			res := CheckAuth()[key]
			assert.Equal(t, "ok", res.Status)
			assert.True(t, strings.HasPrefix(res.Message, "Present (..."))
			assert.True(t, strings.HasSuffix(res.Message, "2345)"))
		})

		t.Run("Length == 4", func(t *testing.T) {
			t.Setenv(key, "1234")
			res := CheckAuth()[key]
			assert.Equal(t, "ok", res.Status)
			assert.Equal(t, "Present", res.Message)
		})

		t.Run("Length < 4", func(t *testing.T) {
			t.Setenv(key, "123")
			res := CheckAuth()[key]
			assert.Equal(t, "ok", res.Status)
			assert.Equal(t, "Present", res.Message)
		})
	})
}
