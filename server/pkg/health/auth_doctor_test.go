package health

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckAuth(t *testing.T) {
	// Helper to ensure clean state for the keys we care about
	clearEnv := func() {
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
	}

	// Helper to set env vars for test duration
	setEnv := func(key, value string) {
		t.Setenv(key, value)
	}

	t.Run("Empty Environment", func(t *testing.T) {
		clearEnv()
		results := CheckAuth()

		// Verify API Keys are missing
		assert.Equal(t, "missing", results["ANTHROPIC_API_KEY"].Status)
		assert.Equal(t, "missing", results["OPENAI_API_KEY"].Status)
		assert.Equal(t, "missing", results["GEMINI_API_KEY"].Status)

		// Verify OAuth is not configured
		assert.Equal(t, "info", results["oauth_GOOGLE"].Status)
		assert.Equal(t, "info", results["oauth_GITHUB"].Status)
	})

	t.Run("API Keys Present and Masked", func(t *testing.T) {
		clearEnv()
		setEnv("ANTHROPIC_API_KEY", "sk-ant-1234567890")
		setEnv("OPENAI_API_KEY", "sk-proj-short") // len("sk-proj-short") = 13 > 4
		setEnv("GEMINI_API_KEY", "AIzaSyD")       // len("AIzaSyD") = 7 > 4

		results := CheckAuth()

		// Anthropic
		assert.Equal(t, "ok", results["ANTHROPIC_API_KEY"].Status)
		assert.Contains(t, results["ANTHROPIC_API_KEY"].Message, "7890")
		assert.NotContains(t, results["ANTHROPIC_API_KEY"].Message, "1234")

		// OpenAI
		assert.Equal(t, "ok", results["OPENAI_API_KEY"].Status)
		assert.Contains(t, results["OPENAI_API_KEY"].Message, "hort")

		// Edge case: Short key
		setEnv("GEMINI_API_KEY", "1234")
		results = CheckAuth()
		assert.Equal(t, "ok", results["GEMINI_API_KEY"].Status)
		assert.Equal(t, "Present", results["GEMINI_API_KEY"].Message)
	})

	t.Run("OAuth Configuration", func(t *testing.T) {
		t.Run("Complete Config", func(t *testing.T) {
			clearEnv()
			setEnv("GOOGLE_CLIENT_ID", "google-id")
			setEnv("GOOGLE_CLIENT_SECRET", "google-secret")

			results := CheckAuth()
			assert.Equal(t, "ok", results["oauth_GOOGLE"].Status)
			assert.Equal(t, "Configured", results["oauth_GOOGLE"].Message)
		})

		t.Run("Partial Config - Missing Secret", func(t *testing.T) {
			clearEnv()
			setEnv("GITHUB_CLIENT_ID", "github-id")

			results := CheckAuth()
			assert.Equal(t, "warning", results["oauth_GITHUB"].Status)
			assert.Equal(t, "Partial configuration (missing ID or Secret)", results["oauth_GITHUB"].Message)
		})

		t.Run("Partial Config - Missing ID", func(t *testing.T) {
			clearEnv()
			setEnv("GITHUB_CLIENT_SECRET", "github-secret")

			results := CheckAuth()
			assert.Equal(t, "warning", results["oauth_GITHUB"].Status)
		})
	})
}
