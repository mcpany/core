// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckAuth(t *testing.T) {
	// Helper to set environment variables for the duration of the test
	setEnv := func(t *testing.T, key, value string) {
		t.Helper()
		t.Setenv(key, value)
	}

	// Helper to unset environment variables for the duration of the test
	unsetEnv := func(t *testing.T, key string) {
		t.Helper()
		originalValue, exists := os.LookupEnv(key)
		err := os.Unsetenv(key)
		if err != nil {
			t.Fatalf("Failed to unset env var %s: %v", key, err)
		}
		t.Cleanup(func() {
			if exists {
				os.Setenv(key, originalValue)
			} else {
				os.Unsetenv(key)
			}
		})
	}

	tests := []struct {
		name     string
		setup    func(t *testing.T)
		validate func(t *testing.T, results map[string]CheckResult)
	}{
		{
			name: "Happy Path - API Keys Present",
			setup: func(t *testing.T) {
				setEnv(t, "ANTHROPIC_API_KEY", "sk-ant-api03-valid-key-123456")
				setEnv(t, "OPENAI_API_KEY", "sk-proj-valid-key-abcdef")
				setEnv(t, "GEMINI_API_KEY", "AIzaSyD-valid-key-789012")
			},
			validate: func(t *testing.T, results map[string]CheckResult) {
				assert.Equal(t, "ok", results["ANTHROPIC_API_KEY"].Status)
				assert.Contains(t, results["ANTHROPIC_API_KEY"].Message, "Present (...")
				assert.Contains(t, results["ANTHROPIC_API_KEY"].Message, "3456)") // Last 4 chars

				assert.Equal(t, "ok", results["OPENAI_API_KEY"].Status)
				assert.Contains(t, results["OPENAI_API_KEY"].Message, "cdef)")

				assert.Equal(t, "ok", results["GEMINI_API_KEY"].Status)
				assert.Contains(t, results["GEMINI_API_KEY"].Message, "9012)")
			},
		},
		{
			name: "Missing Keys",
			setup: func(t *testing.T) {
				unsetEnv(t, "ANTHROPIC_API_KEY")
				unsetEnv(t, "OPENAI_API_KEY")
				unsetEnv(t, "GEMINI_API_KEY")
			},
			validate: func(t *testing.T, results map[string]CheckResult) {
				assert.Equal(t, "missing", results["ANTHROPIC_API_KEY"].Status)
				assert.Equal(t, "Environment variable not set", results["ANTHROPIC_API_KEY"].Message)

				assert.Equal(t, "missing", results["OPENAI_API_KEY"].Status)
				assert.Equal(t, "Environment variable not set", results["OPENAI_API_KEY"].Message)

				assert.Equal(t, "missing", results["GEMINI_API_KEY"].Status)
				assert.Equal(t, "Environment variable not set", results["GEMINI_API_KEY"].Message)
			},
		},
		{
			name: "Short Keys Masking",
			setup: func(t *testing.T) {
				setEnv(t, "ANTHROPIC_API_KEY", "123") // Less than 4 chars
			},
			validate: func(t *testing.T, results map[string]CheckResult) {
				assert.Equal(t, "ok", results["ANTHROPIC_API_KEY"].Status)
				assert.Equal(t, "Present", results["ANTHROPIC_API_KEY"].Message)
				assert.NotContains(t, results["ANTHROPIC_API_KEY"].Message, "...")
			},
		},
		{
			name: "OAuth Configured - Google",
			setup: func(t *testing.T) {
				setEnv(t, "GOOGLE_CLIENT_ID", "google-client-id")
				setEnv(t, "GOOGLE_CLIENT_SECRET", "google-client-secret")
			},
			validate: func(t *testing.T, results map[string]CheckResult) {
				assert.Equal(t, "ok", results["oauth_GOOGLE"].Status)
				assert.Equal(t, "Configured", results["oauth_GOOGLE"].Message)
			},
		},
		{
			name: "OAuth Partial - Github (Only ID)",
			setup: func(t *testing.T) {
				setEnv(t, "GITHUB_CLIENT_ID", "github-client-id")
				unsetEnv(t, "GITHUB_CLIENT_SECRET")
			},
			validate: func(t *testing.T, results map[string]CheckResult) {
				assert.Equal(t, "warning", results["oauth_GITHUB"].Status)
				assert.Equal(t, "Partial configuration (missing ID or Secret)", results["oauth_GITHUB"].Message)
			},
		},
		{
			name: "OAuth Partial - Github (Only Secret)",
			setup: func(t *testing.T) {
				unsetEnv(t, "GITHUB_CLIENT_ID")
				setEnv(t, "GITHUB_CLIENT_SECRET", "github-client-secret")
			},
			validate: func(t *testing.T, results map[string]CheckResult) {
				assert.Equal(t, "warning", results["oauth_GITHUB"].Status)
				assert.Equal(t, "Partial configuration (missing ID or Secret)", results["oauth_GITHUB"].Message)
			},
		},
		{
			name: "OAuth Not Configured",
			setup: func(t *testing.T) {
				unsetEnv(t, "GOOGLE_CLIENT_ID")
				unsetEnv(t, "GOOGLE_CLIENT_SECRET")
				unsetEnv(t, "GITHUB_CLIENT_ID")
				unsetEnv(t, "GITHUB_CLIENT_SECRET")
			},
			validate: func(t *testing.T, results map[string]CheckResult) {
				assert.Equal(t, "info", results["oauth_GOOGLE"].Status)
				assert.Equal(t, "Not configured", results["oauth_GOOGLE"].Message)

				assert.Equal(t, "info", results["oauth_GITHUB"].Status)
				assert.Equal(t, "Not configured", results["oauth_GITHUB"].Message)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Clear environment before each test case to ensure isolation
			// Although t.Setenv restores, explicitly unsetting relevant keys first ensures a clean slate
			// if the test environment already had them set.
			// However, t.Setenv will restore to the *original* state, so if the environment had them set,
			// we should rely on that or explicitly unset inside setup if needed.
			// The safest way here is to just run setup.

			tc.setup(t)
			results := CheckAuth()
			tc.validate(t, results)
		})
	}
}
