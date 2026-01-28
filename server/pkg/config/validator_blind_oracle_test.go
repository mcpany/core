// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func strPtr(s string) *string { return &s }

func TestSecurity_BlindOracle_SecretFile_WithSkip(t *testing.T) {
	// Create a secret file in CWD to pass IsAllowedPath check
	cwd, err := os.Getwd()
	require.NoError(t, err)
	tmpFile, err := os.CreateTemp(cwd, "secret")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	_, err = tmpFile.WriteString("SuperSecretValue")
	require.NoError(t, err)
	tmpFile.Close()

	// Helper to validate config with a regex
	check := func(regex string) []config.ValidationError {
		filePath := tmpFile.Name()
		secretValue := configv1.SecretValue_builder{
			ValidationRegex: &regex,
			FilePath:        &filePath,
		}.Build()

		apiKey := configv1.APIKeyAuth_builder{
			ParamName: strPtr("X-API-Key"),
			Value:     secretValue,
		}.Build()

		auth := configv1.Authentication_builder{
			ApiKey: apiKey,
		}.Build()

		user := configv1.User_builder{
			Id:             strPtr("test-user"),
			Authentication: auth,
		}.Build()

		cfg := configv1.McpAnyServerConfig_builder{
			Users: []*configv1.User{user},
		}.Build()

		// Use the Skip context!
		ctx := util.WithSkipSecretValidation(context.Background())
		return config.Validate(ctx, cfg, config.Server)
	}

	// 1. Regex matches -> No error
	errs := check("^Super.*")
	assert.Empty(t, errs, "Expected no errors when regex matches")

	// 2. Regex does not match -> NO Error (Validation skipped)
	errs = check("^Wrong.*")
	assert.Empty(t, errs, "Expected no errors when regex mismatch because validation should be skipped")
}

func TestSecurity_BlindOracle_EnvironmentVariable_WithSkip(t *testing.T) {
	os.Setenv("TEST_SECRET_ENV", "SuperSecretEnvValue")
	defer os.Unsetenv("TEST_SECRET_ENV")

	check := func(regex string) []config.ValidationError {
		envVar := "TEST_SECRET_ENV"
		secretValue := configv1.SecretValue_builder{
			ValidationRegex:     &regex,
			EnvironmentVariable: &envVar,
		}.Build()

		apiKey := configv1.APIKeyAuth_builder{
			ParamName: strPtr("X-API-Key"),
			Value:     secretValue,
		}.Build()

		auth := configv1.Authentication_builder{
			ApiKey: apiKey,
		}.Build()

		user := configv1.User_builder{
			Id:             strPtr("test-user"),
			Authentication: auth,
		}.Build()

		cfg := configv1.McpAnyServerConfig_builder{
			Users: []*configv1.User{user},
		}.Build()

		// Use the Skip context!
		ctx := util.WithSkipSecretValidation(context.Background())
		return config.Validate(ctx, cfg, config.Server)
	}

	// 1. Regex matches -> No error
	errs := check("^Super.*")
	assert.Empty(t, errs)

	// 2. Regex does not match -> NO Error (Validation skipped)
	errs = check("^Wrong.*")
	assert.Empty(t, errs, "Expected no errors when regex mismatch because validation should be skipped")
}

func TestSecurity_BlindOracle_ExistenceCheck_WithSkip(t *testing.T) {
    // This test verifies that we don't leak file existence information.
    // We use a non-existent file in CWD.
    cwd, _ := os.Getwd()
    nonExistentFile := filepath.Join(cwd, "non_existent_secret_file.txt")

    secretValue := configv1.SecretValue_builder{
        FilePath: &nonExistentFile,
    }.Build()

    apiKey := configv1.APIKeyAuth_builder{
        ParamName: strPtr("X-API-Key"),
        Value:     secretValue,
    }.Build()

    auth := configv1.Authentication_builder{
        ApiKey: apiKey,
    }.Build()

    user := configv1.User_builder{
        Id:             strPtr("test-user"),
        Authentication: auth,
    }.Build()

    cfg := configv1.McpAnyServerConfig_builder{
        Users: []*configv1.User{user},
    }.Build()

    // Case 1: Without Skip -> Should Error (File not found)
    errs := config.Validate(context.Background(), cfg, config.Server)
    assert.NotEmpty(t, errs, "Expected error when not skipping validation")
    if len(errs) > 0 {
        // It might be "secret file ... does not exist" or "failed to resolve ... failed to read secret"
        assert.Error(t, errs[0].Err)
    }

    // Case 2: With Skip -> Should NOT Error (Existence check skipped)
    ctx := util.WithSkipSecretValidation(context.Background())
    errs = config.Validate(ctx, cfg, config.Server)
    assert.Empty(t, errs, "Expected no error when skipping validation, even if file does not exist")
}
