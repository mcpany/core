// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestValidate_Security_StdioArgs_NoExtension_Block(t *testing.T) {
	// 1. Create a directory OUTSIDE the current working directory.
	tempDir, err := os.MkdirTemp("", "mcpany-repro-outside-noext")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 2. Create a "malicious" script in that external directory WITHOUT an extension.
	scriptPath := filepath.Join(tempDir, "malicious_script")
	err = os.WriteFile(scriptPath, []byte("print('pwned')"), 0644)
	require.NoError(t, err)

	// 3. Configure a tool that uses python to execute this script.
	jsonConfig := `{
		"upstream_services": [
			{
				"name": "test-service-bypass-noext",
				"mcp_service": {
					"stdio_connection": {
						"command": "python3",
						"args": ["` + scriptPath + `"]
					}
				}
			}
		]
	}`

	cfg := configv1.McpAnyServerConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(jsonConfig), cfg))

	// Mock execLookPath
	oldLookPath := execLookPath
	defer func() { execLookPath = oldLookPath }()
	execLookPath = func(file string) (string, error) {
		return "/usr/bin/python3", nil
	}

	// 4. Validate the config.
	// We expect validation errors because the script is outside CWD.
	errors := Validate(context.Background(), cfg, Server)

	assert.NotEmpty(t, errors, "Expected validation error for script without extension outside CWD (vulnerability blocked)")
	if len(errors) > 0 {
		assert.Contains(t, errors[0].Error(), "file path")
		assert.Contains(t, errors[0].Error(), "is not allowed")
	}
}

func TestValidate_Security_StdioArgs_NoExt_ButPath_Block(t *testing.T) {
	// This test ensures that arguments with path separators (but no extension) are treated as paths and validated.

	// 1. Create a directory OUTSIDE the current working directory.
	tempDir, err := os.MkdirTemp("", "mcpany-repro-outside-path")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 2. Create a script there.
	scriptPath := filepath.Join(tempDir, "myscript")
	err = os.WriteFile(scriptPath, []byte("echo hi"), 0755)
	require.NoError(t, err)

	// 3. Configure service using the absolute path (which has separators)
	// Even though it has no extension, the presence of separators should trigger validation.
	jsonConfig := `{
		"upstream_services": [
			{
				"name": "test-service-path-check",
				"mcp_service": {
					"stdio_connection": {
						"command": "bash",
						"args": ["` + scriptPath + `"]
					}
				}
			}
		]
	}`

	cfg := configv1.McpAnyServerConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(jsonConfig), cfg))

	oldLookPath := execLookPath
	defer func() { execLookPath = oldLookPath }()
	execLookPath = func(file string) (string, error) {
		return "/bin/bash", nil
	}

	errors := Validate(context.Background(), cfg, Server)

	assert.NotEmpty(t, errors, "Expected validation error for absolute path without extension")
	if len(errors) > 0 {
		assert.Contains(t, errors[0].Error(), "file path")
		assert.Contains(t, errors[0].Error(), "is not allowed")
	}
}
