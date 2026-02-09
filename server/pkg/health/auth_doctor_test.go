// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckAuth_APIKeys(t *testing.T) {
	// Clean slate
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("GEMINI_API_KEY", "")

	// Test 1: All Missing
	results := CheckAuth()
	for _, key := range []string{"ANTHROPIC_API_KEY", "OPENAI_API_KEY", "GEMINI_API_KEY"} {
		assert.Equal(t, "missing", results[key].Status)
		assert.Equal(t, "Environment variable not set", results[key].Message)
	}

	// Test 2: Present and Masked
	t.Setenv("ANTHROPIC_API_KEY", "sk-ant-1234567890abcdef")
	results = CheckAuth()
	assert.Equal(t, "ok", results["ANTHROPIC_API_KEY"].Status)
	assert.True(t, strings.HasPrefix(results["ANTHROPIC_API_KEY"].Message, "Present (..."))
	assert.True(t, strings.HasSuffix(results["ANTHROPIC_API_KEY"].Message, "cdef)"), "Should show last 4 chars")
	assert.NotContains(t, results["ANTHROPIC_API_KEY"].Message, "1234567890ab", "Should not leak middle part")

	// Test 3: Short Key (Edge Case)
	t.Setenv("OPENAI_API_KEY", "123")
	results = CheckAuth()
	assert.Equal(t, "ok", results["OPENAI_API_KEY"].Status)
	assert.Equal(t, "Present", results["OPENAI_API_KEY"].Message, "Short keys should not be partially revealed")
}

func TestCheckAuth_OAuth(t *testing.T) {
	// Clean slate
	t.Setenv("GOOGLE_CLIENT_ID", "")
	t.Setenv("GOOGLE_CLIENT_SECRET", "")
	t.Setenv("GITHUB_CLIENT_ID", "")
	t.Setenv("GITHUB_CLIENT_SECRET", "")

	// Test 1: Not Configured
	results := CheckAuth()
	assert.Equal(t, "info", results["oauth_GOOGLE"].Status)
	assert.Equal(t, "Not configured", results["oauth_GOOGLE"].Message)

	// Test 2: Partial Configuration (Missing Secret)
	t.Setenv("GOOGLE_CLIENT_ID", "google-client-id")
	results = CheckAuth()
	assert.Equal(t, "warning", results["oauth_GOOGLE"].Status)
	assert.Equal(t, "Partial configuration (missing ID or Secret)", results["oauth_GOOGLE"].Message)

	// Test 3: Partial Configuration (Missing ID)
	t.Setenv("GOOGLE_CLIENT_ID", "")
	t.Setenv("GOOGLE_CLIENT_SECRET", "google-client-secret")
	results = CheckAuth()
	assert.Equal(t, "warning", results["oauth_GOOGLE"].Status)

	// Test 4: Fully Configured
	t.Setenv("GITHUB_CLIENT_ID", "github-client-id")
	t.Setenv("GITHUB_CLIENT_SECRET", "github-client-secret")
	results = CheckAuth()
	assert.Equal(t, "ok", results["oauth_GITHUB"].Status)
	assert.Equal(t, "Configured", results["oauth_GITHUB"].Message)
}
