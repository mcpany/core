// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func mockExecLookPath() func() {
	oldLookPath := execLookPath
	execLookPath = func(file string) (string, error) {
		return "/bin/ls", nil
	}
	return func() { execLookPath = oldLookPath }
}

func TestPlainTextSecretValidation(t *testing.T) {
	cleanup := mockExecLookPath()
	defer cleanup()

	config := func() *configv1.McpAnyServerConfig {
		cfg := &configv1.McpAnyServerConfig{}
		svc := &configv1.UpstreamServiceConfig{}
		svc.SetName("test-plaintext-secret")

		mcp := &configv1.McpUpstreamService{}
		stdio := &configv1.McpStdioConnection{}
		stdio.SetCommand("ls")

		secret := &configv1.SecretValue{}
		secret.SetPlainText("invalid-key")
		secret.SetValidationRegex("^sk-[a-zA-Z0-9]{10}$")

		stdio.SetEnv(map[string]*configv1.SecretValue{
			"TEST_KEY": secret,
		})

		mcp.SetStdioConnection(stdio)
		svc.SetMcpService(mcp)

		cfg.SetUpstreamServices([]*configv1.UpstreamServiceConfig{svc})
		return cfg
	}()

	errs := Validate(context.Background(), config, Server)

	assert.NotEmpty(t, errs, "Validation errors expected for invalid PlainText")
	if len(errs) > 0 {
		assert.Contains(t, errs[0].Err.Error(), "secret value does not match validation regex")
	}
}

func TestEnvSecretValidation(t *testing.T) {
	cleanup := mockExecLookPath()
	defer cleanup()

	os.Setenv("TEST_ENV_KEY", "invalid-key")
	defer os.Unsetenv("TEST_ENV_KEY")

	config := func() *configv1.McpAnyServerConfig {
		cfg := &configv1.McpAnyServerConfig{}
		svc := &configv1.UpstreamServiceConfig{}
		svc.SetName("test-env-secret")

		mcp := &configv1.McpUpstreamService{}
		stdio := &configv1.McpStdioConnection{}
		stdio.SetCommand("ls")

		secret := &configv1.SecretValue{}
		secret.SetEnvironmentVariable("TEST_ENV_KEY")
		secret.SetValidationRegex("^sk-[a-zA-Z0-9]{10}$")

		stdio.SetEnv(map[string]*configv1.SecretValue{
			"TEST_KEY": secret,
		})

		mcp.SetStdioConnection(stdio)
		svc.SetMcpService(mcp)

		cfg.SetUpstreamServices([]*configv1.UpstreamServiceConfig{svc})
		return cfg
	}()

	errs := Validate(context.Background(), config, Server)

	assert.NotEmpty(t, errs, "Validation errors expected for invalid Env var")
	if len(errs) > 0 {
		assert.Contains(t, errs[0].Err.Error(), "secret value does not match validation regex")
	}
}

func TestEmptyPlainTextSecretValidation(t *testing.T) {
	cleanup := mockExecLookPath()
	defer cleanup()

	config := func() *configv1.McpAnyServerConfig {
		cfg := &configv1.McpAnyServerConfig{}
		svc := &configv1.UpstreamServiceConfig{}
		svc.SetName("test-empty-plaintext")

		mcp := &configv1.McpUpstreamService{}
		stdio := &configv1.McpStdioConnection{}
		stdio.SetCommand("ls")

		secret := &configv1.SecretValue{}
		secret.SetPlainText("")
		secret.SetValidationRegex("^.+$")

		stdio.SetEnv(map[string]*configv1.SecretValue{
			"TEST_KEY": secret,
		})

		mcp.SetStdioConnection(stdio)
		svc.SetMcpService(mcp)

		cfg.SetUpstreamServices([]*configv1.UpstreamServiceConfig{svc})
		return cfg
	}()

	errs := Validate(context.Background(), config, Server)

	// This should fail because empty string doesn't match ^.+$
	assert.NotEmpty(t, errs, "Validation errors expected for empty PlainText")
	if len(errs) > 0 {
		assert.Contains(t, errs[0].Err.Error(), "secret value does not match validation regex")
	}
}

func TestWhitespaceInEnvVar_WithRegex(t *testing.T) {
	cleanup := mockExecLookPath()
	defer cleanup()

	// Set env var with whitespace
	os.Setenv("TEST_WHITESPACE_KEY", "  valid-key  ")
	defer os.Unsetenv("TEST_WHITESPACE_KEY")

	config := func() *configv1.McpAnyServerConfig {
		cfg := &configv1.McpAnyServerConfig{}
		svc := &configv1.UpstreamServiceConfig{}
		svc.SetName("test-whitespace")

		mcp := &configv1.McpUpstreamService{}
		stdio := &configv1.McpStdioConnection{}
		stdio.SetCommand("ls")

		secret := &configv1.SecretValue{}
		secret.SetEnvironmentVariable("TEST_WHITESPACE_KEY")
		secret.SetValidationRegex("^valid-key$")

		stdio.SetEnv(map[string]*configv1.SecretValue{
			"TEST_KEY": secret,
		})

		mcp.SetStdioConnection(stdio)
		svc.SetMcpService(mcp)

		cfg.SetUpstreamServices([]*configv1.UpstreamServiceConfig{svc})
		return cfg
	}()

	errs := Validate(context.Background(), config, Server)

	// Should be empty if we trim whitespace
	assert.Empty(t, errs, "Validation errors not expected for env var with whitespace")
}

func TestWhitespaceInPlainText_WithRegex(t *testing.T) {
	cleanup := mockExecLookPath()
	defer cleanup()

	config := func() *configv1.McpAnyServerConfig {
		cfg := &configv1.McpAnyServerConfig{}
		svc := &configv1.UpstreamServiceConfig{}
		svc.SetName("test-whitespace-plain")

		mcp := &configv1.McpUpstreamService{}
		stdio := &configv1.McpStdioConnection{}
		stdio.SetCommand("ls")

		secret := &configv1.SecretValue{}
		secret.SetPlainText("  valid-key  ")
		secret.SetValidationRegex("^valid-key$")

		stdio.SetEnv(map[string]*configv1.SecretValue{
			"TEST_KEY": secret,
		})

		mcp.SetStdioConnection(stdio)
		svc.SetMcpService(mcp)

		cfg.SetUpstreamServices([]*configv1.UpstreamServiceConfig{svc})
		return cfg
	}()

	errs := Validate(context.Background(), config, Server)

	// Should be empty if we trim whitespace
	assert.Empty(t, errs, "Validation errors not expected for plain text with whitespace")
}
