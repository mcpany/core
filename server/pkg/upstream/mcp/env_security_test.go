// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestStdioTransport_EnvironmentSecurity(t *testing.T) {
	// Set a secret environment variable in the parent process
	secretKey := "SENTINEL_SECRET"
	secretVal := "super_secret_value"
	os.Setenv(secretKey, secretVal)
	defer os.Unsetenv(secretKey)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Configure a command that prints environment variables
	// Use "env" or "printenv"
	stdio := &configv1.McpStdioConnection{
		Command: proto.String("env"),
		Args:    []string{},
		// Add an explicit env var to ensure it is passed
		Env: map[string]*configv1.SecretValue{
			"EXPLICIT_VAR": {
				Value: &configv1.SecretValue_PlainText{
					PlainText: "explicit_value",
				},
			},
		},
	}

	// Use buildCommandFromStdioConfig directly since we want to inspect the generated command
	// or run it and check output.
	// Since buildCommandFromStdioConfig returns *exec.Cmd, we can run it.
	// Note: We are testing the internal function buildCommandFromStdioConfig which is available in package mcp

	cmd, err := buildCommandFromStdioConfig(ctx, stdio, false)
	require.NoError(t, err)

	// Verify cmd.Env directly
	envMap := make(map[string]string)
	for _, e := range cmd.Env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	// 1. Secret from parent should NOT be present
	_, found := envMap[secretKey]
	assert.False(t, found, "Secret env var from parent should NOT be present in child process environment")

	// 2. Explicitly configured var SHOULD be present
	assert.Equal(t, "explicit_value", envMap["EXPLICIT_VAR"], "Explicitly configured env var should be present")

	// 3. Allowed system vars SHOULD be present (if they exist in parent)
	allowedVars := []string{"PATH", "HOME", "USER", "TMPDIR", "TZ", "LANG"}
	for _, key := range allowedVars {
		if val, ok := os.LookupEnv(key); ok {
			assert.Equal(t, val, envMap[key], fmt.Sprintf("Allowed var %s should be preserved", key))
		}
	}
}
