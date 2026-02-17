// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckAuth(t *testing.T) {
	// Clear all relevant env vars first to ensure clean state
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

	t.Run("No Env Vars", func(t *testing.T) {
		results := CheckAuth()
		assert.Equal(t, "missing", results["OPENAI_API_KEY"].Status)
		assert.Equal(t, "info", results["oauth_GOOGLE"].Status)
	})

	t.Run("API Keys Present and Masked", func(t *testing.T) {
		t.Setenv("OPENAI_API_KEY", "sk-1234567890")
		results := CheckAuth()

		res := results["OPENAI_API_KEY"]
		assert.Equal(t, "ok", res.Status)
		assert.Contains(t, res.Message, "Present")
		assert.Contains(t, res.Message, "7890")
		assert.NotContains(t, res.Message, "sk-123456") // Ensure masked
	})

	t.Run("API Keys Short", func(t *testing.T) {
		t.Setenv("GEMINI_API_KEY", "123")
		results := CheckAuth()

		res := results["GEMINI_API_KEY"]
		assert.Equal(t, "ok", res.Status)
		assert.Equal(t, "Present", res.Message) // Too short to mask with suffix
	})

	t.Run("OAuth Partial Config", func(t *testing.T) {
		t.Setenv("GOOGLE_CLIENT_ID", "client-id")
		t.Setenv("GOOGLE_CLIENT_SECRET", "")
		results := CheckAuth()

		res := results["oauth_GOOGLE"]
		assert.Equal(t, "warning", res.Status)
		assert.Equal(t, "Partial configuration (missing ID or Secret)", res.Message)
	})

	t.Run("OAuth Full Config", func(t *testing.T) {
		t.Setenv("GITHUB_CLIENT_ID", "client-id")
		t.Setenv("GITHUB_CLIENT_SECRET", "client-secret")
		results := CheckAuth()

		res := results["oauth_GITHUB"]
		assert.Equal(t, "ok", res.Status)
		assert.Equal(t, "Configured", res.Message)
	})
}
