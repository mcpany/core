// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckAuth(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want map[string]string // simplified expectation: key -> status
		// Additional assertions can be done in the loop
		checkMasking bool
	}{
		{
			name: "No Env Vars",
			env: map[string]string{
				"ANTHROPIC_API_KEY":    "",
				"OPENAI_API_KEY":       "",
				"GEMINI_API_KEY":       "",
				"GOOGLE_CLIENT_ID":     "",
				"GOOGLE_CLIENT_SECRET": "",
				"GITHUB_CLIENT_ID":     "",
				"GITHUB_CLIENT_SECRET": "",
			},
			want: map[string]string{
				"ANTHROPIC_API_KEY": "missing",
				"OPENAI_API_KEY":    "missing",
				"GEMINI_API_KEY":    "missing",
				"oauth_GOOGLE":      "info",
				"oauth_GITHUB":      "info",
			},
		},
		{
			name: "Full Config",
			env: map[string]string{
				"ANTHROPIC_API_KEY":    "sk-ant-12345",
				"OPENAI_API_KEY":       "sk-proj-12345",
				"GEMINI_API_KEY":       "AIzaSy-12345",
				"GOOGLE_CLIENT_ID":     "google-id",
				"GOOGLE_CLIENT_SECRET": "google-secret",
				"GITHUB_CLIENT_ID":     "github-id",
				"GITHUB_CLIENT_SECRET": "github-secret",
			},
			want: map[string]string{
				"ANTHROPIC_API_KEY": "ok",
				"OPENAI_API_KEY":    "ok",
				"GEMINI_API_KEY":    "ok",
				"oauth_GOOGLE":      "ok",
				"oauth_GITHUB":      "ok",
			},
			checkMasking: true,
		},
		{
			name: "Short Key",
			env: map[string]string{
				"OPENAI_API_KEY": "123",
			},
			want: map[string]string{
				"OPENAI_API_KEY": "ok",
			},
		},
		{
			name: "Partial OAuth - Google",
			env: map[string]string{
				"GOOGLE_CLIENT_ID": "google-id",
			},
			want: map[string]string{
				"oauth_GOOGLE": "warning",
			},
		},
		{
			name: "Partial OAuth - Github Secret only",
			env: map[string]string{
				"GITHUB_CLIENT_SECRET": "github-secret",
			},
			want: map[string]string{
				"oauth_GITHUB": "warning",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear env first (or rely on t.Setenv to restore, but we need to clear potential pollution from other tests or this test loop?)
			// t.Setenv restores to original state. But inside the loop, we are accumulating?
			// No, t.Setenv in t.Run only affects that subtest.
			// However, we need to ensure the "base" state is clean.
			// Since we can't easily unset all env vars, we assume the test runner environment doesn't have these set.
			// Just in case, we can explicitly unset them if they exist, but let's assume clean env.

			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			results := CheckAuth()

			for k, wantStatus := range tt.want {
				assert.Contains(t, results, k)
				assert.Equal(t, wantStatus, results[k].Status, "Status mismatch for %s", k)
			}

			if tt.checkMasking {
				// Verify masking for long keys
				res := results["OPENAI_API_KEY"]
				assert.Equal(t, "Present (...2345)", res.Message)
				assert.NotContains(t, res.Message, "sk-proj-1")

				res = results["ANTHROPIC_API_KEY"]
				assert.Equal(t, "Present (...2345)", res.Message)
			}

			if tt.name == "Short Key" {
				res := results["OPENAI_API_KEY"]
				assert.Equal(t, "Present", res.Message)
			}
		})
	}
}
