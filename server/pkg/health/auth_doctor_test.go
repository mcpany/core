// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckAuth(t *testing.T) {
	// Helper to reset env vars is not needed because t.Setenv handles cleanup automatically.

	t.Run("No Keys Set", func(t *testing.T) {
		// Ensure clean state
		t.Setenv("ANTHROPIC_API_KEY", "")
		t.Setenv("OPENAI_API_KEY", "")
		t.Setenv("GEMINI_API_KEY", "")
		t.Setenv("GOOGLE_CLIENT_ID", "")
		t.Setenv("GITHUB_CLIENT_ID", "")

		results := CheckAuth()

		assert.Equal(t, "missing", results["ANTHROPIC_API_KEY"].Status)
		assert.Equal(t, "Environment variable not set", results["ANTHROPIC_API_KEY"].Message)

		assert.Equal(t, "info", results["oauth_GOOGLE"].Status)
		assert.Equal(t, "Not configured", results["oauth_GOOGLE"].Message)
	})

	t.Run("API Keys Present and Masked", func(t *testing.T) {
		t.Setenv("ANTHROPIC_API_KEY", "sk-ant-123456789")

		results := CheckAuth()

		assert.Equal(t, "ok", results["ANTHROPIC_API_KEY"].Status)
		// Check masking: should end with last 4 chars "6789"
		assert.True(t, strings.HasSuffix(results["ANTHROPIC_API_KEY"].Message, "6789)"))
		assert.Contains(t, results["ANTHROPIC_API_KEY"].Message, "Present (...")
		assert.NotContains(t, results["ANTHROPIC_API_KEY"].Message, "12345") // Shouldn't show the middle
	})

	t.Run("Short API Keys Masking", func(t *testing.T) {
		// Edge case: Key length <= 4
		t.Setenv("OPENAI_API_KEY", "123")

		results := CheckAuth()

		assert.Equal(t, "ok", results["OPENAI_API_KEY"].Status)
		assert.Equal(t, "Present", results["OPENAI_API_KEY"].Message)
		assert.NotContains(t, results["OPENAI_API_KEY"].Message, "123") // Should just say "Present" if too short to mask safely?
		// Wait, looking at code:
		// if len(val) > 4 { masked = "Present (..." + val[len(val)-4:] + ")" } else { masked = "Present" }
		// So "123" -> "Present"
	})

	t.Run("OAuth Complete Configuration", func(t *testing.T) {
		t.Setenv("GOOGLE_CLIENT_ID", "google-id")
		t.Setenv("GOOGLE_CLIENT_SECRET", "google-secret")

		results := CheckAuth()

		assert.Equal(t, "ok", results["oauth_GOOGLE"].Status)
		assert.Equal(t, "Configured", results["oauth_GOOGLE"].Message)
	})

	t.Run("OAuth Partial Configuration", func(t *testing.T) {
		// ID set, Secret missing
		t.Setenv("GITHUB_CLIENT_ID", "github-id")
		t.Setenv("GITHUB_CLIENT_SECRET", "")

		results := CheckAuth()

		assert.Equal(t, "warning", results["oauth_GITHUB"].Status)
		assert.Equal(t, "Partial configuration (missing ID or Secret)", results["oauth_GITHUB"].Message)

		// ID missing, Secret set
		t.Setenv("GITHUB_CLIENT_ID", "")
		t.Setenv("GITHUB_CLIENT_SECRET", "github-secret")

		results2 := CheckAuth()
		assert.Equal(t, "warning", results2["oauth_GITHUB"].Status)
	})
}
