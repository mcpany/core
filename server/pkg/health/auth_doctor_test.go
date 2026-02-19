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
			name:    "No Environment Variables",
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
			name: "API Keys Present",
			envVars: map[string]string{
				"ANTHROPIC_API_KEY": "sk-ant-12345678",
				"OPENAI_API_KEY":    "sk-proj-12345678",
			},
			expected: map[string]CheckResult{
				"ANTHROPIC_API_KEY": {Status: "ok", Message: "Present (...5678)"},
				"OPENAI_API_KEY":    {Status: "ok", Message: "Present (...5678)"},
				"GEMINI_API_KEY":    {Status: "missing", Message: "Environment variable not set"},
				"oauth_GOOGLE":      {Status: "info", Message: "Not configured"},
				"oauth_GITHUB":      {Status: "info", Message: "Not configured"},
			},
		},
		{
			name: "Short API Keys",
			envVars: map[string]string{
				"GEMINI_API_KEY": "123",
			},
			expected: map[string]CheckResult{
				"ANTHROPIC_API_KEY": {Status: "missing", Message: "Environment variable not set"},
				"OPENAI_API_KEY":    {Status: "missing", Message: "Environment variable not set"},
				"GEMINI_API_KEY":    {Status: "ok", Message: "Present"},
				"oauth_GOOGLE":      {Status: "info", Message: "Not configured"},
				"oauth_GITHUB":      {Status: "info", Message: "Not configured"},
			},
		},
		{
			name: "OAuth Complete",
			envVars: map[string]string{
				"GOOGLE_CLIENT_ID":     "google-id",
				"GOOGLE_CLIENT_SECRET": "google-secret",
			},
			expected: map[string]CheckResult{
				"ANTHROPIC_API_KEY": {Status: "missing", Message: "Environment variable not set"},
				"OPENAI_API_KEY":    {Status: "missing", Message: "Environment variable not set"},
				"GEMINI_API_KEY":    {Status: "missing", Message: "Environment variable not set"},
				"oauth_GOOGLE":      {Status: "ok", Message: "Configured"},
				"oauth_GITHUB":      {Status: "info", Message: "Not configured"},
			},
		},
		{
			name: "OAuth Partial (Missing Secret)",
			envVars: map[string]string{
				"GITHUB_CLIENT_ID": "github-id",
			},
			expected: map[string]CheckResult{
				"ANTHROPIC_API_KEY": {Status: "missing", Message: "Environment variable not set"},
				"OPENAI_API_KEY":    {Status: "missing", Message: "Environment variable not set"},
				"GEMINI_API_KEY":    {Status: "missing", Message: "Environment variable not set"},
				"oauth_GOOGLE":      {Status: "info", Message: "Not configured"},
				"oauth_GITHUB":      {Status: "warning", Message: "Partial configuration (missing ID or Secret)"},
			},
		},
		{
			name: "OAuth Partial (Missing ID)",
			envVars: map[string]string{
				"GITHUB_CLIENT_SECRET": "github-secret",
			},
			expected: map[string]CheckResult{
				"ANTHROPIC_API_KEY": {Status: "missing", Message: "Environment variable not set"},
				"OPENAI_API_KEY":    {Status: "missing", Message: "Environment variable not set"},
				"GEMINI_API_KEY":    {Status: "missing", Message: "Environment variable not set"},
				"oauth_GOOGLE":      {Status: "info", Message: "Not configured"},
				"oauth_GITHUB":      {Status: "warning", Message: "Partial configuration (missing ID or Secret)"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment first to ensure isolation
			// t.Setenv restores original values after the test
			t.Setenv("ANTHROPIC_API_KEY", "")
			t.Setenv("OPENAI_API_KEY", "")
			t.Setenv("GEMINI_API_KEY", "")
			t.Setenv("GOOGLE_CLIENT_ID", "")
			t.Setenv("GOOGLE_CLIENT_SECRET", "")
			t.Setenv("GITHUB_CLIENT_ID", "")
			t.Setenv("GITHUB_CLIENT_SECRET", "")

			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			results := CheckAuth()

			// Check exact match for keys we care about
			for k, v := range tt.expected {
				assert.Equal(t, v, results[k], "Mismatch for key %s", k)
			}
			// Verify no extra keys if we were checking strictly, but here we just check expected ones.
			// The implementation returns exactly these keys so it should be fine.
			assert.Equal(t, len(tt.expected), len(results))
		})
	}
}
