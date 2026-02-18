// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckAuth(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected map[string]CheckResult
	}{
		{
			name:    "NoEnvVars",
			envVars: map[string]string{},
			expected: map[string]CheckResult{
				"ANTHROPIC_API_KEY": {Status: "missing", Message: "Environment variable not set"},
				"OPENAI_API_KEY":    {Status: "missing", Message: "Environment variable not set"},
				"GEMINI_API_KEY":    {Status: "missing", Message: "Environment variable not set"},
				"oauth_GOOGLE":      {Status: "info", Message: "Not configured"},
				"oauth_GITHUB":      {Status: "info", Message: "Not configured"},
			},
		},
		{
			name: "APIKeysPresent",
			envVars: map[string]string{
				"ANTHROPIC_API_KEY": "sk-ant-1234567890",
			},
			expected: map[string]CheckResult{
				"ANTHROPIC_API_KEY": {Status: "ok", Message: "Present (...7890)"},
				"OPENAI_API_KEY":    {Status: "missing", Message: "Environment variable not set"},
			},
		},
		{
			name: "ShortAPIKey",
			envVars: map[string]string{
				"OPENAI_API_KEY": "123",
			},
			expected: map[string]CheckResult{
				"OPENAI_API_KEY": {Status: "ok", Message: "Present"},
			},
		},
		{
			name: "OAuthComplete",
			envVars: map[string]string{
				"GOOGLE_CLIENT_ID":     "google-id",
				"GOOGLE_CLIENT_SECRET": "google-secret",
			},
			expected: map[string]CheckResult{
				"oauth_GOOGLE": {Status: "ok", Message: "Configured"},
			},
		},
		{
			name: "OAuthPartialID",
			envVars: map[string]string{
				"GITHUB_CLIENT_ID": "github-id",
			},
			expected: map[string]CheckResult{
				"oauth_GITHUB": {Status: "warning", Message: "Partial configuration (missing ID or Secret)"},
			},
		},
		{
			name: "OAuthPartialSecret",
			envVars: map[string]string{
				"GITHUB_CLIENT_SECRET": "github-secret",
			},
			expected: map[string]CheckResult{
				"oauth_GITHUB": {Status: "warning", Message: "Partial configuration (missing ID or Secret)"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear env vars first to ensure clean state
			// Note: In real scenarios, parallel tests might interfere if not handled carefully.
			// t.Setenv handles restoration after test.
			// However, CheckAuth reads specifically named env vars, so we just need to ensure
			// we set/unset what we care about or what might interfere.
			// For simplicity, we just set the ones in the test case.
			// But wait, if the environment already has ANTHROPIC_API_KEY set (e.g. in CI),
			// `t.Setenv` will override it for the test duration, which is good.
			// But for "NoEnvVars", we need to make sure they are UNSET.
			// `t.Setenv(k, "")` sets it to empty string, but `os.Getenv` returns empty string for both unset and set-to-empty.
			// `CheckAuth` checks `if val != ""`. So `t.Setenv(k, "")` works effectively as "unset" for this logic.

			// Explicitly unset all known keys to ensure clean slate for "NoEnvVars" case and others
			keys := []string{
				"ANTHROPIC_API_KEY", "OPENAI_API_KEY", "GEMINI_API_KEY",
				"GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET",
				"GITHUB_CLIENT_ID", "GITHUB_CLIENT_SECRET",
			}
			for _, k := range keys {
				t.Setenv(k, "")
			}

			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			results := CheckAuth()

			for k, v := range tt.expected {
				assert.Equal(t, v, results[k], "CheckResult mismatch for key: %s", k)
			}
		})
	}
}
