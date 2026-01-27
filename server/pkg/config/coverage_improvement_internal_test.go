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
	"google.golang.org/protobuf/proto"
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

	cfg := &configv1.UpstreamServiceConfig{
		Name: proto.String("mcp-svc"),
		ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
			McpService: &configv1.McpUpstreamService{
				ConnectionType: &configv1.McpUpstreamService_StdioConnection{
					StdioConnection: &configv1.McpStdioConnection{
						Command:          proto.String("dummycmd"),
						WorkingDirectory: proto.String("config_test_dummy_file_internal"),
					},
				},
			},
		},
	}

	// Should pass Command check (mocked) and fail Directory check
	err = ValidateOrError(context.Background(), cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is not a directory")
}
