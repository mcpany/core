// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckAuth(t *testing.T) {
	// Clear all relevant environment variables before each test
	cleanupEnv := func(t *testing.T) {
		keys := []string{
			"ANTHROPIC_API_KEY",
			"OPENAI_API_KEY",
			"GEMINI_API_KEY",
			"GOOGLE_CLIENT_ID",
			"GOOGLE_CLIENT_SECRET",
			"GITHUB_CLIENT_ID",
			"GITHUB_CLIENT_SECRET",
		}
		for _, k := range keys {
			t.Setenv(k, "")
		}
	}

	tests := []struct {
		name     string
		env      map[string]string
		expected map[string]string // key -> status
	}{
		{
			name: "All Missing",
			env:  map[string]string{},
			expected: map[string]string{
				"ANTHROPIC_API_KEY": "missing",
				"OPENAI_API_KEY":    "missing",
				"GEMINI_API_KEY":    "missing",
				"oauth_GOOGLE":      "info",
				"oauth_GITHUB":      "info",
			},
		},
		{
			name: "All Present",
			env: map[string]string{
				"ANTHROPIC_API_KEY":    "sk-ant-123456789",
				"OPENAI_API_KEY":       "sk-proj-123456789",
				"GEMINI_API_KEY":       "AIzaSy-123456789",
				"GOOGLE_CLIENT_ID":     "google-id",
				"GOOGLE_CLIENT_SECRET": "google-secret",
				"GITHUB_CLIENT_ID":     "github-id",
				"GITHUB_CLIENT_SECRET": "github-secret",
			},
			expected: map[string]string{
				"ANTHROPIC_API_KEY": "ok",
				"OPENAI_API_KEY":    "ok",
				"GEMINI_API_KEY":    "ok",
				"oauth_GOOGLE":      "ok",
				"oauth_GITHUB":      "ok",
			},
		},
		{
			name: "Partial OAuth",
			env: map[string]string{
				"GOOGLE_CLIENT_ID":     "google-id",
				"GITHUB_CLIENT_SECRET": "github-secret",
			},
			expected: map[string]string{
				"oauth_GOOGLE": "warning",
				"oauth_GITHUB": "warning",
			},
		},
		{
			name: "Short Keys",
			env: map[string]string{
				"ANTHROPIC_API_KEY": "123", // Too short to mask
			},
			expected: map[string]string{
				"ANTHROPIC_API_KEY": "ok",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanupEnv(t)
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			results := CheckAuth()

			for k, expectedStatus := range tt.expected {
				assert.Contains(t, results, k)
				assert.Equal(t, expectedStatus, results[k].Status, "Status mismatch for key %s", k)

				// Additional checks for message content
				if expectedStatus == "ok" {
					if k == "ANTHROPIC_API_KEY" {
						if len(tt.env[k]) > 4 {
							assert.Contains(t, results[k].Message, "Present (...")
							assert.NotContains(t, results[k].Message, tt.env[k][:4], "Key should be masked")
						} else {
							assert.Equal(t, "Present", results[k].Message)
						}
					}
				}
			}
		})
	}
}

// TestCheckAuth_Environment_Restore ensures that environment variables are restored after test.
// This is implicitly handled by t.Setenv, but good to verify if we were doing manual os.Setenv.
// Since we use t.Setenv, this is just a sanity check that it works as expected.
func TestCheckAuth_Manual_Verify(t *testing.T) {
	// Set a value
	key := "OPENAI_API_KEY"
	val := "test-val-12345"
	t.Setenv(key, val)

	results := CheckAuth()
	assert.Equal(t, "ok", results[key].Status)
	assert.Contains(t, results[key].Message, "2345)") // Ends with last 4
}
