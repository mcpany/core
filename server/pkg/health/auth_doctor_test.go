// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckAuth(t *testing.T) {
	t.Run("NoEnvVars", func(t *testing.T) {
		// Clear environment variables
		t.Setenv("ANTHROPIC_API_KEY", "")
		t.Setenv("OPENAI_API_KEY", "")
		t.Setenv("GEMINI_API_KEY", "")
		t.Setenv("GOOGLE_CLIENT_ID", "")
		t.Setenv("GOOGLE_CLIENT_SECRET", "")
		t.Setenv("GITHUB_CLIENT_ID", "")
		t.Setenv("GITHUB_CLIENT_SECRET", "")

		results := CheckAuth()

		assert.Equal(t, "missing", results["ANTHROPIC_API_KEY"].Status)
		assert.Equal(t, "Environment variable not set", results["ANTHROPIC_API_KEY"].Message)

		assert.Equal(t, "missing", results["OPENAI_API_KEY"].Status)
		assert.Equal(t, "Environment variable not set", results["OPENAI_API_KEY"].Message)

		assert.Equal(t, "missing", results["GEMINI_API_KEY"].Status)
		assert.Equal(t, "Environment variable not set", results["GEMINI_API_KEY"].Message)

		// OAuth should be "info" (not configured)
		assert.Equal(t, "Not configured", results["oauth_GOOGLE"].Message)
		assert.Equal(t, "info", results["oauth_GOOGLE"].Status)

		assert.Equal(t, "Not configured", results["oauth_GITHUB"].Message)
		assert.Equal(t, "info", results["oauth_GITHUB"].Status)
	})

	t.Run("APIKeysPresent", func(t *testing.T) {
		t.Setenv("ANTHROPIC_API_KEY", "sk-ant-test-1234")
		t.Setenv("OPENAI_API_KEY", "sk-proj-test-5678")
		t.Setenv("GEMINI_API_KEY", "AIzaSyD-test-9012")

		results := CheckAuth()

		// Verify status
		assert.Equal(t, "ok", results["ANTHROPIC_API_KEY"].Status)
		assert.Equal(t, "ok", results["OPENAI_API_KEY"].Status)
		assert.Equal(t, "ok", results["GEMINI_API_KEY"].Status)

		// Verify masking (should contain last 4 chars)
		assert.Contains(t, results["ANTHROPIC_API_KEY"].Message, "1234")
		assert.NotContains(t, results["ANTHROPIC_API_KEY"].Message, "sk-ant-test") // Should not leak prefix
		assert.Contains(t, results["ANTHROPIC_API_KEY"].Message, "Present (...")

		assert.Contains(t, results["OPENAI_API_KEY"].Message, "5678")
		assert.NotContains(t, results["OPENAI_API_KEY"].Message, "sk-proj-test")

		assert.Contains(t, results["GEMINI_API_KEY"].Message, "9012")
		assert.NotContains(t, results["GEMINI_API_KEY"].Message, "AIzaSyD-test")
	})

	t.Run("OAuthPartial", func(t *testing.T) {
		// Only ClientID set
		t.Setenv("GOOGLE_CLIENT_ID", "google-client-id")
		t.Setenv("GOOGLE_CLIENT_SECRET", "")

		// Only ClientSecret set
		t.Setenv("GITHUB_CLIENT_ID", "")
		t.Setenv("GITHUB_CLIENT_SECRET", "github-client-secret")

		results := CheckAuth()

		assert.Equal(t, "warning", results["oauth_GOOGLE"].Status)
		assert.Equal(t, "Partial configuration (missing ID or Secret)", results["oauth_GOOGLE"].Message)

		assert.Equal(t, "warning", results["oauth_GITHUB"].Status)
		assert.Equal(t, "Partial configuration (missing ID or Secret)", results["oauth_GITHUB"].Message)
	})

	t.Run("OAuthFull", func(t *testing.T) {
		t.Setenv("GOOGLE_CLIENT_ID", "google-client-id")
		t.Setenv("GOOGLE_CLIENT_SECRET", "google-client-secret")

		t.Setenv("GITHUB_CLIENT_ID", "github-client-id")
		t.Setenv("GITHUB_CLIENT_SECRET", "github-client-secret")

		results := CheckAuth()

		assert.Equal(t, "ok", results["oauth_GOOGLE"].Status)
		assert.Equal(t, "Configured", results["oauth_GOOGLE"].Message)

		assert.Equal(t, "ok", results["oauth_GITHUB"].Status)
		assert.Equal(t, "Configured", results["oauth_GITHUB"].Message)
	})

	t.Run("ShortKeyMasking", func(t *testing.T) {
		// Key shorter than 4 chars
		t.Setenv("ANTHROPIC_API_KEY", "abc")

		results := CheckAuth()

		assert.Equal(t, "ok", results["ANTHROPIC_API_KEY"].Status)
		// Logic: if len(val) > 4 { masked = "Present (..." + val[len(val)-4:] + ")" } else { masked = "Present" }
		assert.Equal(t, "Present", results["ANTHROPIC_API_KEY"].Message)
	})
}
