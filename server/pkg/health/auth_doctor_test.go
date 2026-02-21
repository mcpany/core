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
		setup    func(t *testing.T)
		validate func(t *testing.T, results map[string]CheckResult)
	}{
		{
			name: "Happy Path - API Keys Present",
			setup: func(t *testing.T) {
				t.Setenv("ANTHROPIC_API_KEY", "sk-ant-api03-valid-key-123456")
				t.Setenv("OPENAI_API_KEY", "sk-proj-valid-key-abcdef")
				t.Setenv("GEMINI_API_KEY", "AIzaSyD-valid-key-789012")
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
				// Set to empty string, which CheckAuth handles as missing
				t.Setenv("ANTHROPIC_API_KEY", "")
				t.Setenv("OPENAI_API_KEY", "")
				t.Setenv("GEMINI_API_KEY", "")
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
				t.Setenv("ANTHROPIC_API_KEY", "123") // Less than 4 chars
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
				t.Setenv("GOOGLE_CLIENT_ID", "google-client-id")
				t.Setenv("GOOGLE_CLIENT_SECRET", "google-client-secret")
			},
			validate: func(t *testing.T, results map[string]CheckResult) {
				assert.Equal(t, "ok", results["oauth_GOOGLE"].Status)
				assert.Equal(t, "Configured", results["oauth_GOOGLE"].Message)
			},
		},
		{
			name: "OAuth Partial - Github (Only ID)",
			setup: func(t *testing.T) {
				t.Setenv("GITHUB_CLIENT_ID", "github-client-id")
				t.Setenv("GITHUB_CLIENT_SECRET", "")
			},
			validate: func(t *testing.T, results map[string]CheckResult) {
				assert.Equal(t, "warning", results["oauth_GITHUB"].Status)
				assert.Equal(t, "Partial configuration (missing ID or Secret)", results["oauth_GITHUB"].Message)
			},
		},
		{
			name: "OAuth Partial - Github (Only Secret)",
			setup: func(t *testing.T) {
				t.Setenv("GITHUB_CLIENT_ID", "")
				t.Setenv("GITHUB_CLIENT_SECRET", "github-client-secret")
			},
			validate: func(t *testing.T, results map[string]CheckResult) {
				assert.Equal(t, "warning", results["oauth_GITHUB"].Status)
				assert.Equal(t, "Partial configuration (missing ID or Secret)", results["oauth_GITHUB"].Message)
			},
		},
		{
			name: "OAuth Not Configured",
			setup: func(t *testing.T) {
				t.Setenv("GOOGLE_CLIENT_ID", "")
				t.Setenv("GOOGLE_CLIENT_SECRET", "")
				t.Setenv("GITHUB_CLIENT_ID", "")
				t.Setenv("GITHUB_CLIENT_SECRET", "")
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
			tc.setup(t)
			results := CheckAuth()
			tc.validate(t, results)
		})
	}
}
