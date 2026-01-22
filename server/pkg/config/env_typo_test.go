// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestEnvVarTypoSuggestion(t *testing.T) {
	// Set a similar environment variable
	os.Setenv("OPENAI_API_KEY_123", "secret-value")
	defer os.Unsetenv("OPENAI_API_KEY_123")

	// Create a secret value pointing to a slightly different name
	secret := &configv1.SecretValue{
		Value: &configv1.SecretValue_EnvironmentVariable{
			EnvironmentVariable: "OPENAI_API_KEY",
		},
	}

	// Validate
	err := validateSecretValue(secret)

	// Check if the error suggests the typo
	assert.Error(t, err)

	// We expect the error to be an ActionableError
	ae, ok := err.(*ActionableError)
	assert.True(t, ok, "Error should be ActionableError")

	t.Logf("Error message: %s", ae.Error())
	t.Logf("Suggestion: %s", ae.Suggestion)

	// New expectation: It DOES suggest the typo.
	assert.Contains(t, ae.Suggestion, `Did you mean "OPENAI_API_KEY_123"?`)
}
