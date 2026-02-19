// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckAuth(t *testing.T) {
	t.Run("Empty State", func(t *testing.T) {
		// Ensure clean state
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

		assert.Equal(t, "info", results["oauth_GOOGLE"].Status)
		assert.Equal(t, "Not configured", results["oauth_GOOGLE"].Message)
	})

	t.Run("Long API Key", func(t *testing.T) {
		t.Setenv("ANTHROPIC_API_KEY", "sk-ant-123456789")
		results := CheckAuth()

		assert.Equal(t, "ok", results["ANTHROPIC_API_KEY"].Status)
		assert.Equal(t, "Present (...6789)", results["ANTHROPIC_API_KEY"].Message)
	})

	t.Run("Short API Key", func(t *testing.T) {
		t.Setenv("OPENAI_API_KEY", "123")
		results := CheckAuth()

		assert.Equal(t, "ok", results["OPENAI_API_KEY"].Status)
		assert.Equal(t, "Present", results["OPENAI_API_KEY"].Message)
	})

	t.Run("OAuth Full", func(t *testing.T) {
		t.Setenv("GOOGLE_CLIENT_ID", "google-id")
		t.Setenv("GOOGLE_CLIENT_SECRET", "google-secret")
		results := CheckAuth()

		assert.Equal(t, "ok", results["oauth_GOOGLE"].Status)
		assert.Equal(t, "Configured", results["oauth_GOOGLE"].Message)
	})

	t.Run("OAuth Partial ID Only", func(t *testing.T) {
		t.Setenv("GITHUB_CLIENT_ID", "github-id")
		t.Setenv("GITHUB_CLIENT_SECRET", "")
		results := CheckAuth()

		assert.Equal(t, "warning", results["oauth_GITHUB"].Status)
		assert.Equal(t, "Partial configuration (missing ID or Secret)", results["oauth_GITHUB"].Message)
	})

	t.Run("OAuth Partial Secret Only", func(t *testing.T) {
		t.Setenv("GITHUB_CLIENT_ID", "")
		t.Setenv("GITHUB_CLIENT_SECRET", "github-secret")
		results := CheckAuth()

		assert.Equal(t, "warning", results["oauth_GITHUB"].Status)
		assert.Equal(t, "Partial configuration (missing ID or Secret)", results["oauth_GITHUB"].Message)
	})
}
