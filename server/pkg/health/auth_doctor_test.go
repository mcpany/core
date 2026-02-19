// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckAuth(t *testing.T) {
	// Helper to unset all relevant env vars
	clearEnv := func() {
		vars := []string{
			"ANTHROPIC_API_KEY", "OPENAI_API_KEY", "GEMINI_API_KEY",
			"GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET",
			"GITHUB_CLIENT_ID", "GITHUB_CLIENT_SECRET",
		}
		for _, v := range vars {
			os.Unsetenv(v)
		}
	}

	tests := []struct {
		name     string
		setup    func()
		expected map[string]CheckResult
	}{
		{
			name:  "Empty State",
			setup: func() {},
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
			setup: func() {
				os.Setenv("ANTHROPIC_API_KEY", "sk-ant-12345")
				os.Setenv("OPENAI_API_KEY", "sk-proj-abcde")
			},
			expected: map[string]CheckResult{
				"ANTHROPIC_API_KEY": {Status: "ok", Message: "Present (...2345)"},
				"OPENAI_API_KEY":    {Status: "ok", Message: "Present (...bcde)"},
				"GEMINI_API_KEY":    {Status: "missing", Message: "Environment variable not set"},
				"oauth_GOOGLE":      {Status: "info", Message: "Not configured"},
				"oauth_GITHUB":      {Status: "info", Message: "Not configured"},
			},
		},
		{
			name: "Short API Key",
			setup: func() {
				os.Setenv("ANTHROPIC_API_KEY", "123")
			},
			expected: map[string]CheckResult{
				"ANTHROPIC_API_KEY": {Status: "ok", Message: "Present"}, // Should not mask if length <= 4
				"OPENAI_API_KEY":    {Status: "missing", Message: "Environment variable not set"},
				"GEMINI_API_KEY":    {Status: "missing", Message: "Environment variable not set"},
				"oauth_GOOGLE":      {Status: "info", Message: "Not configured"},
				"oauth_GITHUB":      {Status: "info", Message: "Not configured"},
			},
		},
		{
			name: "OAuth Complete",
			setup: func() {
				os.Setenv("GOOGLE_CLIENT_ID", "google-id")
				os.Setenv("GOOGLE_CLIENT_SECRET", "google-secret")
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
			name: "OAuth Partial (ID only)",
			setup: func() {
				os.Setenv("GITHUB_CLIENT_ID", "github-id")
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
			name: "OAuth Partial (Secret only)",
			setup: func() {
				os.Setenv("GITHUB_CLIENT_SECRET", "github-secret")
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
			clearEnv()
			tt.setup()
			defer clearEnv() // Cleanup after test

			results := CheckAuth()

			for k, expectedRes := range tt.expected {
				actualRes, ok := results[k]
				assert.True(t, ok, "Expected key %s to be present in results", k)
				assert.Equal(t, expectedRes.Status, actualRes.Status, "Status mismatch for %s", k)
				assert.Equal(t, expectedRes.Message, actualRes.Message, "Message mismatch for %s", k)
			}
		})
	}
}
