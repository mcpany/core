// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckAuth(t *testing.T) {
	// Helper to clear env vars for isolation
	clearEnv := func(t *testing.T) {
		vars := []string{
			"ANTHROPIC_API_KEY", "OPENAI_API_KEY", "GEMINI_API_KEY",
			"GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET",
			"GITHUB_CLIENT_ID", "GITHUB_CLIENT_SECRET",
		}
		for _, v := range vars {
			t.Setenv(v, "")
		}
	}

	t.Run("No Environment Variables", func(t *testing.T) {
		clearEnv(t)
		results := CheckAuth()

		assert.Equal(t, "missing", results["ANTHROPIC_API_KEY"].Status)
		assert.Equal(t, "Environment variable not set", results["ANTHROPIC_API_KEY"].Message)

		assert.Equal(t, "missing", results["OPENAI_API_KEY"].Status)
		assert.Equal(t, "missing", results["GEMINI_API_KEY"].Status)

		assert.Equal(t, "info", results["oauth_GOOGLE"].Status)
		assert.Equal(t, "Not configured", results["oauth_GOOGLE"].Message)

		assert.Equal(t, "info", results["oauth_GITHUB"].Status)
	})

	t.Run("API Keys Present & Masked", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("OPENAI_API_KEY", "sk-1234567890")

		results := CheckAuth()

		assert.Equal(t, "ok", results["OPENAI_API_KEY"].Status)
		assert.Equal(t, "Present (...7890)", results["OPENAI_API_KEY"].Message)
	})

	t.Run("Short API Keys", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("GEMINI_API_KEY", "123")

		results := CheckAuth()

		assert.Equal(t, "ok", results["GEMINI_API_KEY"].Status)
		assert.Equal(t, "Present", results["GEMINI_API_KEY"].Message)
	})

	t.Run("OAuth Full Configuration", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("GITHUB_CLIENT_ID", "gh-id")
		t.Setenv("GITHUB_CLIENT_SECRET", "gh-secret")

		results := CheckAuth()

		assert.Equal(t, "ok", results["oauth_GITHUB"].Status)
		assert.Equal(t, "Configured", results["oauth_GITHUB"].Message)
	})

	t.Run("OAuth Partial Configuration", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("GOOGLE_CLIENT_ID", "google-id")
		// Secret missing

		results := CheckAuth()

		assert.Equal(t, "warning", results["oauth_GOOGLE"].Status)
		assert.Equal(t, "Partial configuration (missing ID or Secret)", results["oauth_GOOGLE"].Message)
	})

	t.Run("OAuth Partial Configuration (Only Secret)", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("GOOGLE_CLIENT_SECRET", "google-secret")
		// ID missing

		results := CheckAuth()

		assert.Equal(t, "warning", results["oauth_GOOGLE"].Status)
		assert.Equal(t, "Partial configuration (missing ID or Secret)", results["oauth_GOOGLE"].Message)
	})
}
