// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckAuth(t *testing.T) {
	t.Run("MissingAllKeys", func(t *testing.T) {
		// Use t.Setenv with empty string to simulate missing keys
		// The code treats empty string same as missing (os.Getenv returns "")
		keys := []string{
			"ANTHROPIC_API_KEY",
			"OPENAI_API_KEY",
			"GEMINI_API_KEY",
			"GOOGLE_CLIENT_ID",
			"GOOGLE_CLIENT_SECRET",
			"GITHUB_CLIENT_ID",
			"GITHUB_CLIENT_SECRET",
		}
		for _, k := range keys {
			t.Setenv(k, "")
		}

		results := CheckAuth()

		assert.Equal(t, "missing", results["ANTHROPIC_API_KEY"].Status)
		assert.Equal(t, "missing", results["OPENAI_API_KEY"].Status)
		assert.Equal(t, "missing", results["GEMINI_API_KEY"].Status)
		assert.Equal(t, "info", results["oauth_GOOGLE"].Status)
		assert.Equal(t, "info", results["oauth_GITHUB"].Status)
	})

	t.Run("PresentAPIKeys", func(t *testing.T) {
		t.Setenv("ANTHROPIC_API_KEY", "sk-ant-12345678")
		t.Setenv("OPENAI_API_KEY", "sk-proj-12345678")
		t.Setenv("GEMINI_API_KEY", "AIzaSy12345678")

		results := CheckAuth()

		assert.Equal(t, "ok", results["ANTHROPIC_API_KEY"].Status)
		assert.Contains(t, results["ANTHROPIC_API_KEY"].Message, "5678)")
		assert.Equal(t, "ok", results["OPENAI_API_KEY"].Status)
		assert.Contains(t, results["OPENAI_API_KEY"].Message, "5678)")
		assert.Equal(t, "ok", results["GEMINI_API_KEY"].Status)
		assert.Contains(t, results["GEMINI_API_KEY"].Message, "5678)")
	})

	t.Run("ShortAPIKey", func(t *testing.T) {
		t.Setenv("ANTHROPIC_API_KEY", "123") // Too short to mask last 4

		results := CheckAuth()

		assert.Equal(t, "ok", results["ANTHROPIC_API_KEY"].Status)
		assert.Equal(t, "Present", results["ANTHROPIC_API_KEY"].Message)
	})

	t.Run("OAuthConfiguration", func(t *testing.T) {
		// Full Google
		t.Setenv("GOOGLE_CLIENT_ID", "google-id")
		t.Setenv("GOOGLE_CLIENT_SECRET", "google-secret")

		// Partial Github (Missing Secret)
		t.Setenv("GITHUB_CLIENT_ID", "github-id")
		t.Setenv("GITHUB_CLIENT_SECRET", "")

		results := CheckAuth()

		assert.Equal(t, "ok", results["oauth_GOOGLE"].Status)
		assert.Equal(t, "Configured", results["oauth_GOOGLE"].Message)

		assert.Equal(t, "warning", results["oauth_GITHUB"].Status)
		assert.Equal(t, "Partial configuration (missing ID or Secret)", results["oauth_GITHUB"].Message)
	})

	t.Run("OAuthConfiguration_MissingID", func(t *testing.T) {
		// Partial Github (Missing ID)
		t.Setenv("GITHUB_CLIENT_ID", "")
		t.Setenv("GITHUB_CLIENT_SECRET", "github-secret")

		results := CheckAuth()

		assert.Equal(t, "warning", results["oauth_GITHUB"].Status)
	})
}
