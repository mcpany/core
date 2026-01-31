// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatorDirectoryExistsWithMock(t *testing.T) {
	// Mock execLookPath
	origLookPath := execLookPath
	defer func() { execLookPath = origLookPath }()
	execLookPath = func(file string) (string, error) {
		return "/usr/bin/dummycmd", nil
	}

	// File as directory
	f, err := os.Create("config_test_dummy_file_internal")
	require.NoError(t, err)
	f.Close()
	defer os.Remove("config_test_dummy_file_internal")

	stdio := &configv1.McpStdioConnection{}
	stdio.SetCommand("dummycmd")
	stdio.SetWorkingDirectory("config_test_dummy_file_internal")

	mcpSvc := &configv1.McpUpstreamService{}
	mcpSvc.SetStdioConnection(stdio)

	cfg := &configv1.UpstreamServiceConfig{}
	cfg.SetName("mcp-svc")
	cfg.SetMcpService(mcpSvc)

	// Should pass Command check (mocked) and fail Directory check
	err = ValidateOrError(context.Background(), cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is not a directory")
}
