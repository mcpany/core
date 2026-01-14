// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"context"
	"os"
	"testing"

	"github.com/mcpany/core/server/pkg/config"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvVarOverrideListWithComma(t *testing.T) {
	// Create a temporary config file with one service
	fs := afero.NewMemMapFs()
	configContent := `
upstream_services:
  - name: "my-service"
    mcp_service:
      stdio_connection:
        command: "echo"
        args: ["original"]
`
	err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)

	// Set env var to override args. We want to pass two arguments:
	// 1. "--msg"
	// 2. "hello, world"
	//
	// Current implementation splits by comma, so we expect this to result in 3 args: "--msg", "hello", " world"
	// causing the test to fail if we assert for 2 args.
	// We will demonstrate the bug by asserting the incorrect behavior, then fix it.
	envVar := "MCPANY__UPSTREAM_SERVICES__0__MCP_SERVICE__STDIO_CONNECTION__ARGS"
	val := "--msg,\"hello, world\"" // intended: ["--msg", "hello, world"]

	_ = os.Setenv(envVar, val)
	defer func() {
		_ = os.Unsetenv(envVar)
	}()

	// Load the config
	store := config.NewFileStore(fs, []string{"/config.yaml"})
	cfg, err := store.Load(context.Background())
	require.NoError(t, err)

	require.Len(t, cfg.UpstreamServices, 1)
	service := cfg.UpstreamServices[0]
	require.NotNil(t, service.GetMcpService())
	require.NotNil(t, service.GetMcpService().GetStdioConnection())

	args := service.GetMcpService().GetStdioConnection().GetArgs()

	// We expect 2 args now that we handle CSV quoting.
	assert.Len(t, args, 2)
	assert.Equal(t, "--msg", args[0])
	assert.Equal(t, "hello, world", args[1])
}
