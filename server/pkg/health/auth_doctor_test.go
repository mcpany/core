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
		validate func(t *testing.T, results map[string]CheckResult)
	}{
		{
			name:    "Empty Environment",
			envVars: map[string]string{},
			validate: func(t *testing.T, results map[string]CheckResult) {
				// Check that all known API keys are reported as missing
				expectedKeys := []string{"ANTHROPIC_API_KEY", "OPENAI_API_KEY", "GEMINI_API_KEY"}
				for _, k := range expectedKeys {
					assert.Contains(t, results, k)
					assert.Equal(t, "missing", results[k].Status)
					assert.Equal(t, "Environment variable not set", results[k].Message)
				}
				// Check that OAuth is not configured
				expectedOAuth := []string{"oauth_GOOGLE", "oauth_GITHUB"}
				for _, k := range expectedOAuth {
					assert.Contains(t, results, k)
					assert.Equal(t, "info", results[k].Status)
					assert.Equal(t, "Not configured", results[k].Message)
				}
			},
		},
		{
			name: "Full Environment",
			envVars: map[string]string{
				"ANTHROPIC_API_KEY":    "sk-ant-123456789",
				"OPENAI_API_KEY":       "sk-proj-987654321",
				"GEMINI_API_KEY":       "AIzaSyD-abcdefg",
				"GOOGLE_CLIENT_ID":     "google-id",
				"GOOGLE_CLIENT_SECRET": "google-secret",
				"GITHUB_CLIENT_ID":     "github-id",
				"GITHUB_CLIENT_SECRET": "github-secret",
			},
			validate: func(t *testing.T, results map[string]CheckResult) {
				assert.Equal(t, "ok", results["ANTHROPIC_API_KEY"].Status)
				assert.Contains(t, results["ANTHROPIC_API_KEY"].Message, "Present (...")
				assert.Contains(t, results["ANTHROPIC_API_KEY"].Message, "6789)")

				assert.Equal(t, "ok", results["OPENAI_API_KEY"].Status)
				assert.Equal(t, "ok", results["GEMINI_API_KEY"].Status)

				assert.Equal(t, "ok", results["oauth_GOOGLE"].Status)
				assert.Equal(t, "Configured", results["oauth_GOOGLE"].Message)

				assert.Equal(t, "ok", results["oauth_GITHUB"].Status)
				assert.Equal(t, "Configured", results["oauth_GITHUB"].Message)
			},
		},
		{
			name: "Short API Keys",
			envVars: map[string]string{
				"OPENAI_API_KEY": "123",
			},
			validate: func(t *testing.T, results map[string]CheckResult) {
				assert.Equal(t, "ok", results["OPENAI_API_KEY"].Status)
				assert.Equal(t, "Present", results["OPENAI_API_KEY"].Message)
			},
		},
		{
			name: "Partial OAuth (Client ID only)",
			envVars: map[string]string{
				"GOOGLE_CLIENT_ID": "google-id",
			},
			validate: func(t *testing.T, results map[string]CheckResult) {
				assert.Equal(t, "warning", results["oauth_GOOGLE"].Status)
				assert.Equal(t, "Partial configuration (missing ID or Secret)", results["oauth_GOOGLE"].Message)
			},
		},
		{
			name: "Partial OAuth (Client Secret only)",
			envVars: map[string]string{
				"GITHUB_CLIENT_SECRET": "github-secret",
			},
			validate: func(t *testing.T, results map[string]CheckResult) {
				assert.Equal(t, "warning", results["oauth_GITHUB"].Status)
				assert.Equal(t, "Partial configuration (missing ID or Secret)", results["oauth_GITHUB"].Message)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keysToManage := []string{
				"ANTHROPIC_API_KEY", "OPENAI_API_KEY", "GEMINI_API_KEY",
				"GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET",
				"GITHUB_CLIENT_ID", "GITHUB_CLIENT_SECRET",
			}

			// Clear env first to ensure isolation from host env.
			// t.Setenv(k, "") sets the environment variable to an empty string.
			// Since CheckAuth uses os.Getenv(k) != "", this effectively simulates
			// the variable being "unset" or empty, which is sufficient for our tests.
			for _, k := range keysToManage {
				t.Setenv(k, "")
			}

			// Now set the specific test vars
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			results := CheckAuth()
			tt.validate(t, results)
		})
	}
}
