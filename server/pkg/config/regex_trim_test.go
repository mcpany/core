// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestRegexTrimValidationBug(t *testing.T) {
	// Setup: plain_text value with spaces
	config := func() *configv1.McpAnyServerConfig {
		cfg := &configv1.McpAnyServerConfig{}
		svc := &configv1.UpstreamServiceConfig{}
		svc.SetName("test-trim-bug")

		mcp := &configv1.McpUpstreamService{}
		stdio := &configv1.McpStdioConnection{}
		stdio.SetCommand("ls")

		secret := &configv1.SecretValue{}
		secret.SetPlainText(" value ")
		secret.SetValidationRegex("^value$")

		stdio.SetEnv(map[string]*configv1.SecretValue{
			"TEST_TRIM": secret,
		})

		mcp.SetStdioConnection(stdio)
		svc.SetMcpService(mcp)

		cfg.SetUpstreamServices([]*configv1.UpstreamServiceConfig{svc})
		return cfg
	}()

	// This assumes that execLookPath is mocked or "ls" exists.
	// We need to mock execLookPath to avoid dependency on system "ls" or PATH.
	// But `validator.go` uses `execLookPath` var which we can override in test package.

	oldLookPath := execLookPath
	defer func() { execLookPath = oldLookPath }()
	execLookPath = func(file string) (string, error) {
		return "/bin/ls", nil
	}

	errs := Validate(context.Background(), config, Server)

	// Expectation: It should pass because validation logic now trims the value.
	assert.Empty(t, errs, "Validation errors not expected")
}
