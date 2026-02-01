// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckAuth(t *testing.T) {
	tests := []struct {
		name     string
		env      map[string]string
		expected map[string]CheckResult
	}{
		{
			name: "All Missing",
			env:  map[string]string{},
			expected: map[string]CheckResult{
				"ANTHROPIC_API_KEY": {Status: "missing", Message: "Environment variable not set"},
				"OPENAI_API_KEY":    {Status: "missing", Message: "Environment variable not set"},
				"GEMINI_API_KEY":    {Status: "missing", Message: "Environment variable not set"},
				"oauth_GOOGLE":      {Status: "info", Message: "Not configured"},
				"oauth_GITHUB":      {Status: "info", Message: "Not configured"},
			},
		},
		{
			name: "API Keys Present - Short",
			env: map[string]string{
				"ANTHROPIC_API_KEY": "123",
				"OPENAI_API_KEY":    "abc",
				"GEMINI_API_KEY":    "xyz",
			},
			expected: map[string]CheckResult{
				"ANTHROPIC_API_KEY": {Status: "ok", Message: "Present"},
				"OPENAI_API_KEY":    {Status: "ok", Message: "Present"},
				"GEMINI_API_KEY":    {Status: "ok", Message: "Present"},
			},
		},
		{
			name: "API Keys Present - Long (Masking)",
			env: map[string]string{
				"ANTHROPIC_API_KEY": "sk-ant-123456789",
				"OPENAI_API_KEY":    "sk-proj-abcdefghi",
				"GEMINI_API_KEY":    "AIzaSyD-987654321",
			},
			expected: map[string]CheckResult{
				"ANTHROPIC_API_KEY": {Status: "ok", Message: "Present (...6789)"},
				"OPENAI_API_KEY":    {Status: "ok", Message: "Present (...fghi)"},
				"GEMINI_API_KEY":    {Status: "ok", Message: "Present (...4321)"},
			},
		},
		{
			name: "OAuth Fully Configured",
			env: map[string]string{
				"GOOGLE_CLIENT_ID":     "google-id",
				"GOOGLE_CLIENT_SECRET": "google-secret",
				"GITHUB_CLIENT_ID":     "github-id",
				"GITHUB_CLIENT_SECRET": "github-secret",
			},
			expected: map[string]CheckResult{
				"oauth_GOOGLE": {Status: "ok", Message: "Configured"},
				"oauth_GITHUB": {Status: "ok", Message: "Configured"},
			},
		},
		{
			name: "OAuth Partial - Missing Secret",
			env: map[string]string{
				"GOOGLE_CLIENT_ID": "google-id",
			},
			expected: map[string]CheckResult{
				"oauth_GOOGLE": {Status: "warning", Message: "Partial configuration (missing ID or Secret)"},
			},
		},
		{
			name: "OAuth Partial - Missing ID",
			env: map[string]string{
				"GITHUB_CLIENT_SECRET": "github-secret",
			},
			expected: map[string]CheckResult{
				"oauth_GITHUB": {Status: "warning", Message: "Partial configuration (missing ID or Secret)"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear env first to ensure isolation from host env
			// Note: t.Setenv restores the original environment after the test
			t.Setenv("ANTHROPIC_API_KEY", "")
			t.Setenv("OPENAI_API_KEY", "")
			t.Setenv("GEMINI_API_KEY", "")
			t.Setenv("GOOGLE_CLIENT_ID", "")
			t.Setenv("GOOGLE_CLIENT_SECRET", "")
			t.Setenv("GITHUB_CLIENT_ID", "")
			t.Setenv("GITHUB_CLIENT_SECRET", "")

			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			results := CheckAuth()

			for k, expected := range tt.expected {
				assert.Contains(t, results, k)
				assert.Equal(t, expected, results[k], "Check failed for %s", k)
			}
		})
	}
}
