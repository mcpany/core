// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

// TestReproWhitespaceInURL checks validation for URLs with trailing whitespace.
func TestReproWhitespaceInURL(t *testing.T) {
	// 1. Setup Env Var with trailing whitespace
	envVar := "TEST_URL_WITH_SPACE"
	val := "http://example.com " // Trailing space!
	os.Setenv(envVar, val)
	defer os.Unsetenv(envVar)

    // Note: We simulate the validator receiving a string with a space,
    // which happens if 'expand' injects it or if it's hardcoded in YAML.

    badURL := "http://example.com "

    // Validate http_service
    svc := &configv1.HttpUpstreamService{
        Address: &badURL,
    }

    err := validateHTTPService(svc)
    assert.Error(t, err)

    // Assert the new friendly error message
    if err != nil {
        assert.True(t, strings.Contains(err.Error(), "contains hidden whitespace"), "Error should mention hidden whitespace. Got: %v", err)

        // If it's an ActionableError, check the suggestion
        if ae, ok := err.(*ActionableError); ok {
            assert.Contains(t, ae.Suggestion, "remove any trailing spaces", "Suggestion should advise removing spaces")
        }
    }
}
