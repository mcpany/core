// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckAuth(t *testing.T) {
	// Helper to ensure clean slate for env vars we care about
	resetEnv := func(t *testing.T) {
		keys := []string{
			"ANTHROPIC_API_KEY", "OPENAI_API_KEY", "GEMINI_API_KEY",
			"GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET",
			"GITHUB_CLIENT_ID", "GITHUB_CLIENT_SECRET",
		}
		for _, k := range keys {
			t.Setenv(k, "")
		}
	}

	t.Run("Empty State", func(t *testing.T) {
		resetEnv(t)
		results := CheckAuth()

		assert.Equal(t, "missing", results["ANTHROPIC_API_KEY"].Status)
		assert.Equal(t, "missing", results["OPENAI_API_KEY"].Status)
		assert.Equal(t, "missing", results["GEMINI_API_KEY"].Status)

		assert.Equal(t, "info", results["oauth_GOOGLE"].Status)
		assert.Equal(t, "info", results["oauth_GITHUB"].Status)
	})

	t.Run("API Keys Present", func(t *testing.T) {
		resetEnv(t)
		t.Setenv("ANTHROPIC_API_KEY", "sk-ant-123456789")
		results := CheckAuth()

		assert.Equal(t, "ok", results["ANTHROPIC_API_KEY"].Status)
		assert.Equal(t, "Present (...6789)", results["ANTHROPIC_API_KEY"].Message)
	})

	t.Run("Short API Keys", func(t *testing.T) {
		resetEnv(t)
		t.Setenv("OPENAI_API_KEY", "123")
		results := CheckAuth()

		assert.Equal(t, "ok", results["OPENAI_API_KEY"].Status)
		assert.Equal(t, "Present", results["OPENAI_API_KEY"].Message)
	})

	t.Run("OAuth Complete", func(t *testing.T) {
		resetEnv(t)
		t.Setenv("GOOGLE_CLIENT_ID", "google-id")
		t.Setenv("GOOGLE_CLIENT_SECRET", "google-secret")
		results := CheckAuth()

		assert.Equal(t, "ok", results["oauth_GOOGLE"].Status)
		assert.Equal(t, "Configured", results["oauth_GOOGLE"].Message)
	})

	t.Run("OAuth Partial", func(t *testing.T) {
		resetEnv(t)
		t.Setenv("GITHUB_CLIENT_ID", "github-id")
		// GITHUB_CLIENT_SECRET is explicitly unset by resetEnv
		results := CheckAuth()

		assert.Equal(t, "warning", results["oauth_GITHUB"].Status)
		assert.Contains(t, results["oauth_GITHUB"].Message, "Partial configuration")
	})
}
