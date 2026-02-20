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
			name: "All Present",
			env: map[string]string{
				"ANTHROPIC_API_KEY":    "sk-ant-12345678",
				"OPENAI_API_KEY":       "sk-proj-12345678",
				"GEMINI_API_KEY":       "AIzaSyD-12345678",
				"GOOGLE_CLIENT_ID":     "google-id",
				"GOOGLE_CLIENT_SECRET": "google-secret",
				"GITHUB_CLIENT_ID":     "github-id",
				"GITHUB_CLIENT_SECRET": "github-secret",
			},
			expected: map[string]CheckResult{
				"ANTHROPIC_API_KEY": {Status: "ok", Message: "Present (...5678)"},
				"OPENAI_API_KEY":    {Status: "ok", Message: "Present (...5678)"},
				"GEMINI_API_KEY":    {Status: "ok", Message: "Present (...5678)"},
				"oauth_GOOGLE":      {Status: "ok", Message: "Configured"},
				"oauth_GITHUB":      {Status: "ok", Message: "Configured"},
			},
		},
		{
			name: "Short Keys",
			env: map[string]string{
				"ANTHROPIC_API_KEY": "123",
			},
			expected: map[string]CheckResult{
				"ANTHROPIC_API_KEY": {Status: "ok", Message: "Present"},
			},
		},
		{
			name: "Partial OAuth (ID only)",
			env: map[string]string{
				"GOOGLE_CLIENT_ID": "google-id",
			},
			expected: map[string]CheckResult{
				"oauth_GOOGLE": {Status: "warning", Message: "Partial configuration (missing ID or Secret)"},
			},
		},
		{
			name: "Partial OAuth (Secret only)",
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
			// Clear all relevant env vars first to ensure clean state
			keys := []string{
				"ANTHROPIC_API_KEY", "OPENAI_API_KEY", "GEMINI_API_KEY",
				"GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET",
				"GITHUB_CLIENT_ID", "GITHUB_CLIENT_SECRET",
			}
			for _, k := range keys {
				t.Setenv(k, "")
			}

			// Set test env vars
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			results := CheckAuth()

			for k, expectedResult := range tt.expected {
				assert.Contains(t, results, k)
				assert.Equal(t, expectedResult.Status, results[k].Status)
				assert.Equal(t, expectedResult.Message, results[k].Message)
			}
		})
	}
}
