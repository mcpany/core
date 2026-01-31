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
		env      map[string]string
		expected map[string]CheckResult
	}{
		{
			name: "Empty Environment",
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
			name: "Full Configuration with Long Keys",
			env: map[string]string{
				"ANTHROPIC_API_KEY":    "sk-ant-1234567890",
				"OPENAI_API_KEY":       "sk-proj-abcdefghij",
				"GEMINI_API_KEY":       "AIzaSyD-klmnopqrst",
				"GOOGLE_CLIENT_ID":     "google-id",
				"GOOGLE_CLIENT_SECRET": "google-secret",
				"GITHUB_CLIENT_ID":     "github-id",
				"GITHUB_CLIENT_SECRET": "github-secret",
			},
			expected: map[string]CheckResult{
				"ANTHROPIC_API_KEY": {Status: "ok", Message: "Present (...7890)"},
				"OPENAI_API_KEY":    {Status: "ok", Message: "Present (...ghij)"},
				"GEMINI_API_KEY":    {Status: "ok", Message: "Present (...qrst)"},
				"oauth_GOOGLE":      {Status: "ok", Message: "Configured"},
				"oauth_GITHUB":      {Status: "ok", Message: "Configured"},
			},
		},
		{
			name: "Short Secrets",
			env: map[string]string{
				"ANTHROPIC_API_KEY": "abc",
				"OPENAI_API_KEY":    "1234",
			},
			expected: map[string]CheckResult{
				"ANTHROPIC_API_KEY": {Status: "ok", Message: "Present"},
				"OPENAI_API_KEY":    {Status: "ok", Message: "Present"},
			},
		},
		{
			name: "Partial OAuth Configuration",
			env: map[string]string{
				"GOOGLE_CLIENT_ID":     "google-id-only",
				"GITHUB_CLIENT_SECRET": "github-secret-only",
			},
			expected: map[string]CheckResult{
				"oauth_GOOGLE": {Status: "warning", Message: "Partial configuration (missing ID or Secret)"},
				"oauth_GITHUB": {Status: "warning", Message: "Partial configuration (missing ID or Secret)"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment first to ensure isolation from the host environment
			keys := []string{
				"ANTHROPIC_API_KEY", "OPENAI_API_KEY", "GEMINI_API_KEY",
				"GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET",
				"GITHUB_CLIENT_ID", "GITHUB_CLIENT_SECRET",
			}
			for _, k := range keys {
				t.Setenv(k, "")
			}

			// Set test environment
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			results := CheckAuth()

			// Check specific expectations
			for k, expectedResult := range tt.expected {
				assert.Contains(t, results, k)
				assert.Equal(t, expectedResult, results[k], "Mismatch for key %s", k)
			}
		})
	}
}
